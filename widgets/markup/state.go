//go:build windows

package markup

import (
	"reflect"
	"strconv"
	"strings"
	"sync"
)

// State stores markup binding data and notifies bound documents after updates.
//
// The state store is optimized for map-based snapshots. Use Set/Patch when the
// root object is a map[string]any. Use Replace when you want to swap in an
// arbitrary Go value such as a struct snapshot.
type State struct {
	mu        sync.RWMutex
	data      any
	watchers  map[uint64]func(stateChange)
	nextWatch uint64
}

type stateChange struct {
	paths []string
}

// NewState creates a binding state store with an optional initial snapshot.
func NewState(data any) *State {
	if data == nil {
		data = map[string]any{}
	}
	return &State{
		data:     data,
		watchers: map[uint64]func(stateChange){},
	}
}

// Get returns the value located at the provided dot-separated path.
//
// An empty path returns the entire root snapshot.
func (s *State) Get(path string) (any, bool) {
	if s == nil {
		return nil, false
	}
	s.mu.RLock()
	data := s.data
	s.mu.RUnlock()
	return lookupStateValue(data, path)
}

// Replace swaps the entire root snapshot and refreshes all bindings.
func (s *State) Replace(data any) bool {
	if s == nil {
		return false
	}
	if data == nil {
		data = map[string]any{}
	}

	s.mu.Lock()
	if reflect.DeepEqual(s.data, data) {
		s.mu.Unlock()
		return false
	}
	s.data = data
	watchers := s.watchersLocked()
	s.mu.Unlock()

	s.notify(watchers, stateChange{})
	return true
}

// Set writes a value to a dot-separated path and refreshes bindings that
// depend on the path. Missing intermediate nodes are created as map[string]any.
//
// Set only mutates map-based snapshots. If the current root snapshot is not a
// map[string]any, use Replace with a new snapshot instead.
func (s *State) Set(path string, value any) bool {
	if s == nil {
		return false
	}
	path = normalizeBindingPath(path)
	if path == "" {
		return s.Replace(value)
	}

	s.mu.Lock()
	changed := setStateMapPath(&s.data, path, value)
	if !changed {
		s.mu.Unlock()
		return false
	}
	watchers := s.watchersLocked()
	s.mu.Unlock()

	s.notify(watchers, stateChange{paths: []string{path}})
	return true
}

// Patch applies multiple Set operations and refreshes once if any value changed.
func (s *State) Patch(values map[string]any) bool {
	if s == nil || len(values) == 0 {
		return false
	}

	s.mu.Lock()
	changedPaths := make([]string, 0, len(values))
	for path, value := range values {
		normalized := normalizeBindingPath(path)
		if normalized == "" {
			if reflect.DeepEqual(s.data, value) {
				continue
			}
			s.data = value
			changedPaths = append(changedPaths, "")
			continue
		}
		if setStateMapPath(&s.data, normalized, value) {
			changedPaths = append(changedPaths, normalized)
		}
	}
	if len(changedPaths) == 0 {
		s.mu.Unlock()
		return false
	}
	watchers := s.watchersLocked()
	s.mu.Unlock()

	s.notify(watchers, stateChange{paths: changedPaths})
	return true
}

func (s *State) snapshot() any {
	if s == nil {
		return nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data
}

func (s *State) subscribe(fn func(stateChange)) func() {
	if s == nil || fn == nil {
		return func() {}
	}

	s.mu.Lock()
	s.nextWatch++
	id := s.nextWatch
	s.watchers[id] = fn
	s.mu.Unlock()

	return func() {
		s.mu.Lock()
		delete(s.watchers, id)
		s.mu.Unlock()
	}
}

func (s *State) watchersLocked() []func(stateChange) {
	if len(s.watchers) == 0 {
		return nil
	}
	watchers := make([]func(stateChange), 0, len(s.watchers))
	for _, watcher := range s.watchers {
		watchers = append(watchers, watcher)
	}
	return watchers
}

func (s *State) notify(watchers []func(stateChange), change stateChange) {
	for _, watcher := range watchers {
		if watcher != nil {
			watcher(change)
		}
	}
}

func setStateMapPath(root *any, path string, value any) bool {
	if root == nil {
		return false
	}
	if *root == nil {
		*root = map[string]any{}
	}

	current, ok := (*root).(map[string]any)
	if !ok {
		return false
	}

	segments := splitBindingPath(path)
	if len(segments) == 0 {
		if reflect.DeepEqual(*root, value) {
			return false
		}
		*root = value
		return true
	}

	for _, segment := range segments[:len(segments)-1] {
		next, exists := current[segment]
		if !exists || next == nil {
			child := map[string]any{}
			current[segment] = child
			current = child
			continue
		}
		child, ok := next.(map[string]any)
		if !ok {
			return false
		}
		current = child
	}

	last := segments[len(segments)-1]
	if reflect.DeepEqual(current[last], value) {
		return false
	}
	current[last] = value
	return true
}

func lookupStateValue(data any, path string) (any, bool) {
	path = normalizeBindingPath(path)
	if path == "" {
		return data, data != nil
	}

	current := reflect.ValueOf(data)
	for _, segment := range splitBindingPath(path) {
		next, ok := descendStateValue(current, segment)
		if !ok {
			return nil, false
		}
		current = next
	}
	if !current.IsValid() {
		return nil, false
	}
	return current.Interface(), true
}

func descendStateValue(current reflect.Value, segment string) (reflect.Value, bool) {
	current = unwrapStateValue(current)
	if !current.IsValid() {
		return reflect.Value{}, false
	}

	switch current.Kind() {
	case reflect.Map:
		if current.Type().Key().Kind() != reflect.String {
			return reflect.Value{}, false
		}
		next := current.MapIndex(reflect.ValueOf(segment))
		if !next.IsValid() {
			return reflect.Value{}, false
		}
		return next, true
	case reflect.Struct:
		return stateStructField(current, segment)
	case reflect.Slice, reflect.Array:
		index, err := strconv.Atoi(segment)
		if err != nil || index < 0 || index >= current.Len() {
			return reflect.Value{}, false
		}
		return current.Index(index), true
	default:
		return reflect.Value{}, false
	}
}

func unwrapStateValue(value reflect.Value) reflect.Value {
	for value.IsValid() {
		switch value.Kind() {
		case reflect.Interface, reflect.Pointer:
			if value.IsNil() {
				return reflect.Value{}
			}
			value = value.Elem()
		default:
			return value
		}
	}
	return value
}

func stateStructField(current reflect.Value, segment string) (reflect.Value, bool) {
	fieldType, ok := stateFieldType(current.Type(), segment)
	if !ok {
		return reflect.Value{}, false
	}
	return current.FieldByIndex(fieldType.Index), true
}

func stateFieldType(typ reflect.Type, segment string) (reflect.StructField, bool) {
	lowerSegment := strings.ToLower(strings.TrimSpace(segment))
	for index := 0; index < typ.NumField(); index++ {
		field := typ.Field(index)
		if field.PkgPath != "" {
			continue
		}
		if strings.ToLower(field.Name) == lowerSegment {
			return field, true
		}
		if tag := strings.ToLower(strings.TrimSpace(stateTagName(field))); tag == lowerSegment && tag != "" {
			return field, true
		}
	}
	return reflect.StructField{}, false
}

func stateTagName(field reflect.StructField) string {
	for _, tagName := range []string{"json", "state", "markup"} {
		tagValue := field.Tag.Get(tagName)
		if tagValue == "" {
			continue
		}
		name := strings.TrimSpace(strings.Split(tagValue, ",")[0])
		if name != "" && name != "-" {
			return name
		}
	}
	return ""
}
