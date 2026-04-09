//go:build windows

package jsonui

import (
	"reflect"
	"strconv"
	"strings"
	"sync"
)

// Change reports which data paths changed in the host data source.
type Change struct {
	Paths []string
}

// DataSource supplies read access plus change notifications for JSON UI bindings.
//
// JSON documents only declare binding relationships. The host owns all data
// mutations and can implement this interface directly, or use Store.
type DataSource interface {
	Get(path string) (any, bool)
	Subscribe(func(Change)) func()
}

// Store is the default in-memory DataSource implementation for JSON UI.
type Store struct {
	mu        sync.RWMutex
	data      any
	watchers  map[uint64]func(Change)
	nextWatch uint64
}

// NewStore creates a Store with an optional initial snapshot.
func NewStore(data any) *Store {
	if data == nil {
		data = map[string]any{}
	}
	return &Store{
		data:     data,
		watchers: map[uint64]func(Change){},
	}
}

// Get returns the value stored at a dot-separated path.
func (s *Store) Get(path string) (any, bool) {
	if s == nil {
		return nil, false
	}
	s.mu.RLock()
	data := s.data
	s.mu.RUnlock()
	return lookupPathValue(data, path)
}

// Replace swaps the entire snapshot and notifies all bindings.
func (s *Store) Replace(data any) bool {
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

	s.notify(watchers, Change{})
	return true
}

// Set updates a single dot-separated path in a map-based snapshot.
func (s *Store) Set(path string, value any) bool {
	if s == nil {
		return false
	}
	path = normalizeBindingPath(path)
	if path == "" {
		return s.Replace(value)
	}

	s.mu.Lock()
	changed := setPathValue(&s.data, path, value)
	if !changed {
		s.mu.Unlock()
		return false
	}
	watchers := s.watchersLocked()
	s.mu.Unlock()

	s.notify(watchers, Change{Paths: []string{path}})
	return true
}

// Patch applies multiple Set operations and emits one notification when needed.
func (s *Store) Patch(values map[string]any) bool {
	if s == nil || len(values) == 0 {
		return false
	}

	s.mu.Lock()
	paths := make([]string, 0, len(values))
	for path, value := range values {
		normalized := normalizeBindingPath(path)
		if normalized == "" {
			if reflect.DeepEqual(s.data, value) {
				continue
			}
			s.data = value
			paths = append(paths, "")
			continue
		}
		if setPathValue(&s.data, normalized, value) {
			paths = append(paths, normalized)
		}
	}
	if len(paths) == 0 {
		s.mu.Unlock()
		return false
	}
	watchers := s.watchersLocked()
	s.mu.Unlock()

	s.notify(watchers, Change{Paths: paths})
	return true
}

// Subscribe registers a change callback and returns an unsubscribe function.
func (s *Store) Subscribe(fn func(Change)) func() {
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

func (s *Store) watchersLocked() []func(Change) {
	if len(s.watchers) == 0 {
		return nil
	}
	out := make([]func(Change), 0, len(s.watchers))
	for _, watcher := range s.watchers {
		out = append(out, watcher)
	}
	return out
}

func (s *Store) notify(watchers []func(Change), change Change) {
	for _, watcher := range watchers {
		if watcher != nil {
			watcher(change)
		}
	}
}

func setPathValue(root *any, path string, value any) bool {
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
	parts := splitBindingPath(path)
	if len(parts) == 0 {
		if reflect.DeepEqual(*root, value) {
			return false
		}
		*root = value
		return true
	}

	for _, part := range parts[:len(parts)-1] {
		next, exists := current[part]
		if !exists || next == nil {
			child := map[string]any{}
			current[part] = child
			current = child
			continue
		}
		child, ok := next.(map[string]any)
		if !ok {
			return false
		}
		current = child
	}

	last := parts[len(parts)-1]
	if reflect.DeepEqual(current[last], value) {
		return false
	}
	current[last] = value
	return true
}

func lookupPathValue(data any, path string) (any, bool) {
	path = normalizeBindingPath(path)
	if path == "" {
		return data, data != nil
	}

	current := reflect.ValueOf(data)
	for _, part := range splitBindingPath(path) {
		next, ok := descendPathValue(current, part)
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

func descendPathValue(value reflect.Value, part string) (reflect.Value, bool) {
	value = unwrapValue(value)
	if !value.IsValid() {
		return reflect.Value{}, false
	}

	switch value.Kind() {
	case reflect.Map:
		if value.Type().Key().Kind() != reflect.String {
			return reflect.Value{}, false
		}
		next := value.MapIndex(reflect.ValueOf(part))
		if !next.IsValid() {
			return reflect.Value{}, false
		}
		return next, true
	case reflect.Struct:
		field, ok := structField(value.Type(), part)
		if !ok {
			return reflect.Value{}, false
		}
		return value.FieldByIndex(field.Index), true
	case reflect.Slice, reflect.Array:
		index, err := strconv.Atoi(part)
		if err != nil || index < 0 || index >= value.Len() {
			return reflect.Value{}, false
		}
		return value.Index(index), true
	default:
		return reflect.Value{}, false
	}
}

func unwrapValue(value reflect.Value) reflect.Value {
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

func structField(typ reflect.Type, name string) (reflect.StructField, bool) {
	lowerName := strings.ToLower(strings.TrimSpace(name))
	for index := 0; index < typ.NumField(); index++ {
		field := typ.Field(index)
		if field.PkgPath != "" {
			continue
		}
		if strings.ToLower(field.Name) == lowerName {
			return field, true
		}
		for _, tagName := range []string{"json", "state", "jsonui"} {
			tagValue := field.Tag.Get(tagName)
			if tagValue == "" {
				continue
			}
			tagNameValue := strings.TrimSpace(strings.Split(tagValue, ",")[0])
			if tagNameValue != "" && tagNameValue != "-" && strings.ToLower(tagNameValue) == lowerName {
				return field, true
			}
		}
	}
	return reflect.StructField{}, false
}

func normalizeBindingPath(path string) string {
	return strings.Trim(strings.TrimSpace(path), ".")
}

func splitBindingPath(path string) []string {
	path = normalizeBindingPath(path)
	if path == "" {
		return nil
	}
	parts := strings.Split(path, ".")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}
