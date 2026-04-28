//go:build windows

package jsonui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// componentParamKind 描述组件参数当前支持的 JSON 值类型。
//
// 这里故意只支持少量稳定类型，避免模板系统快速滑向通用脚本语言。
type componentParamKind uint8

const (
	componentParamAny componentParamKind = iota
	componentParamString
	componentParamInt
	componentParamBool
	componentParamObject
	componentParamArray
)

// componentTemplateFileSpec 对应外部组件模板文件的原始结构。
//
// 模板文件只允许声明一个组件，并通过 node 输出一棵节点子树。
type componentTemplateFileSpec struct {
	Component string                            `json:"component"`
	Params    map[string]componentParamFileSpec `json:"params"`
	Node      any                               `json:"node"`
}

// componentParamFileSpec 描述模板文件中的单个参数声明。
//
// type 可选：若未填写且存在 default，则从 default 推导；
// 若两者都未提供，则表示接受任意 JSON 值。
type componentParamFileSpec struct {
	Type    string          `json:"type"`
	Default json.RawMessage `json:"default"`
}

// componentTemplate 是解析并校验后的模板缓存结果。
type componentTemplate struct {
	Name   string
	Path   string
	Params map[string]componentParam
	Node   map[string]any
	Slots  map[string]struct{}
}

// componentParam 是运行时使用的参数定义。
type componentParam struct {
	Kind         componentParamKind
	HasKind      bool
	HasDefault   bool
	DefaultValue any
}

// componentExpander 负责把主文档里的组件实例展开成普通节点树。
//
// 它只在加载阶段运行，不参与运行时绑定或绘制。
type componentExpander struct {
	registry map[string]string
	cache    map[string]*componentTemplate
}

// preprocessComponentDocument 在正式构建 widget 之前先展开组件模板。
//
// 这样后面的 loader 仍然只需要面对普通的 nodeSpec，不必感知模板语法。
func preprocessComponentDocument(text string, opts LoadOptions) (string, error) {
	root, err := decodeJSONObject([]byte(text))
	if err != nil {
		return "", err
	}

	registry, err := decodeComponentRegistry(root["components"], opts.AssetsDir)
	if err != nil {
		return "", err
	}
	if len(registry) == 0 {
		return text, nil
	}

	winsValue, ok := root["wins"]
	if !ok {
		return text, nil
	}
	wins, ok := winsValue.([]any)
	if !ok {
		return "", fmt.Errorf("components requires wins to be an array")
	}

	expander := &componentExpander{
		registry: registry,
		cache:    map[string]*componentTemplate{},
	}

	for i, item := range wins {
		window, ok := item.(map[string]any)
		if !ok {
			return "", fmt.Errorf("wins[%d] must be an object", i)
		}
		rootNode, ok := window["root"]
		if !ok {
			continue
		}
		expanded, err := expander.expandNode(rootNode, nil)
		if err != nil {
			return "", err
		}
		window["root"] = expanded
	}

	delete(root, "components")
	data, err := json.Marshal(root)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// decodeComponentRegistry 解析顶层组件注册表，并把相对路径归一到当前文档目录。
func decodeComponentRegistry(value any, assetsDir string) (map[string]string, error) {
	if value == nil {
		return nil, nil
	}
	raw, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("components must be an object")
	}
	if len(raw) == 0 {
		return nil, nil
	}

	registry := make(map[string]string, len(raw))
	for name, item := range raw {
		key := strings.TrimSpace(name)
		if key == "" {
			return nil, fmt.Errorf("component name is empty")
		}
		path, ok := item.(string)
		if !ok {
			return nil, fmt.Errorf("component %q path must be a string", key)
		}
		path = strings.TrimSpace(path)
		if path == "" {
			return nil, fmt.Errorf("component %q path is empty", key)
		}
		registry[key] = resolveComponentPath(assetsDir, path)
	}
	return registry, nil
}

// expandNode 只在“节点位置”做递归展开。
//
// 它不会把普通 style/layout 对象误判成组件实例，从而保持 JSON 语义稳定。
func (e *componentExpander) expandNode(value any, stack []string) (map[string]any, error) {
	node, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("node must be an object")
	}

	if componentName, isInstance, err := componentInstanceName(node); err != nil {
		return nil, err
	} else if isInstance {
		args, err := decodeComponentArgs(node["args"])
		if err != nil {
			return nil, fmt.Errorf("component %q args: %w", componentName, err)
		}
		slots, err := decodeComponentSlots(node["slots"])
		if err != nil {
			return nil, fmt.Errorf("component %q slots: %w", componentName, err)
		}
		return e.instantiate(componentName, args, slots, stack)
	}

	childrenValue, ok := node["children"]
	if !ok {
		return node, nil
	}
	children, ok := childrenValue.([]any)
	if !ok {
		return nil, fmt.Errorf("node children must be an array")
	}
	for i, child := range children {
		expanded, err := e.expandNode(child, stack)
		if err != nil {
			return nil, fmt.Errorf("children[%d]: %w", i, err)
		}
		children[i] = expanded
	}
	node["children"] = children
	return node, nil
}

// instantiate 加载模板、校验参数、替换占位符，并继续展开模板内部的子组件。
func (e *componentExpander) instantiate(name string, args map[string]any, slots map[string][]map[string]any, stack []string) (map[string]any, error) {
	template, err := e.loadTemplate(name)
	if err != nil {
		return nil, err
	}
	if cycle, ok := appendComponentStack(stack, name); !ok {
		return nil, fmt.Errorf("component cycle detected: %s", strings.Join(cycle, " -> "))
	}

	values, err := template.resolveArgs(args)
	if err != nil {
		return nil, fmt.Errorf("component %q: %w", name, err)
	}
	if err := template.validateSlots(slots); err != nil {
		return nil, fmt.Errorf("component %q: %w", name, err)
	}

	resolved, err := materializeComponentNode(template.Node, values, slots)
	if err != nil {
		return nil, fmt.Errorf("component %q: %w", name, err)
	}
	return e.expandNode(resolved, append(stack, name))
}

// loadTemplate 从磁盘读取并缓存模板，避免同一组件被重复解析。
func (e *componentExpander) loadTemplate(name string) (*componentTemplate, error) {
	path, ok := e.registry[name]
	if !ok {
		return nil, fmt.Errorf("component %q is not registered", name)
	}
	if cached := e.cache[path]; cached != nil {
		return cached, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("component %q: %w", name, err)
	}

	spec, err := decodeComponentTemplateFile(data)
	if err != nil {
		return nil, fmt.Errorf("component %q: %w", name, err)
	}
	if strings.TrimSpace(spec.Component) == "" {
		return nil, fmt.Errorf("component %q: template component name is empty", name)
	}
	if strings.TrimSpace(spec.Component) != name {
		return nil, fmt.Errorf("component %q: template declares component %q", name, spec.Component)
	}

	node, ok := spec.Node.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("component %q: node must be an object", name)
	}

	params := make(map[string]componentParam, len(spec.Params))
	for paramName, paramSpec := range spec.Params {
		normalized := strings.TrimSpace(paramName)
		if normalized == "" {
			return nil, fmt.Errorf("component %q: param name is empty", name)
		}
		param, err := buildComponentParam(paramSpec)
		if err != nil {
			return nil, fmt.Errorf("component %q param %q: %w", name, normalized, err)
		}
		params[normalized] = param
	}

	template := &componentTemplate{
		Name:   name,
		Path:   path,
		Params: params,
		Node:   cloneJSONObject(node),
	}
	slots, err := collectComponentSlots(template.Node)
	if err != nil {
		return nil, fmt.Errorf("component %q: %w", name, err)
	}
	template.Slots = slots
	e.cache[path] = template
	return template, nil
}

// resolveArgs 完成参数补默认值、必填校验、未知参数校验和类型校验。
func (t *componentTemplate) resolveArgs(args map[string]any) (map[string]any, error) {
	values := make(map[string]any, len(t.Params))
	for name, value := range args {
		param, ok := t.Params[name]
		if !ok {
			return nil, fmt.Errorf("unknown argument %q", name)
		}
		if err := validateComponentValueType(name, value, param); err != nil {
			return nil, err
		}
		values[name] = cloneJSONValue(value)
	}

	for name, param := range t.Params {
		if _, ok := values[name]; ok {
			continue
		}
		if param.HasDefault {
			values[name] = cloneJSONValue(param.DefaultValue)
			continue
		}
		return nil, fmt.Errorf("missing required argument %q", name)
	}
	return values, nil
}

// validateSlots 校验实例上传入的 slot 是否都在模板中声明。
//
// 当前版本不做复杂 slot 规则：未传的 slot 直接按空数组处理，
// 但传入未声明的 slot 会立即报错，避免模板和实例静默失配。
func (t *componentTemplate) validateSlots(slots map[string][]map[string]any) error {
	for name := range slots {
		if _, ok := t.Slots[name]; ok {
			continue
		}
		return fmt.Errorf("unknown slot %q", name)
	}
	return nil
}

// buildComponentParam 把模板文件中的参数定义转成运行时结构。
func buildComponentParam(spec componentParamFileSpec) (componentParam, error) {
	param := componentParam{}
	if typeName := strings.TrimSpace(spec.Type); typeName != "" {
		kind, err := parseComponentParamKind(typeName)
		if err != nil {
			return componentParam{}, err
		}
		param.Kind = kind
		param.HasKind = true
	}

	if len(spec.Default) == 0 {
		return param, nil
	}

	value, err := decodeJSONValue(spec.Default)
	if err != nil {
		return componentParam{}, fmt.Errorf("default: %w", err)
	}
	param.HasDefault = true
	param.DefaultValue = cloneJSONValue(value)

	if !param.HasKind {
		if inferred, ok := inferComponentParamKind(value); ok {
			param.Kind = inferred
			param.HasKind = true
		}
	}
	if err := validateComponentValueType("default", value, param); err != nil {
		return componentParam{}, err
	}
	return param, nil
}

// substituteComponentValue 递归替换模板节点中的 ${param} 占位符。
//
// 仅允许“整个 JSON 值就是一个占位符”，不支持字符串拼接模板。
func substituteComponentValue(value any, args map[string]any) (map[string]any, error) {
	resolved, err := substituteComponentJSON(value, args)
	if err != nil {
		return nil, err
	}
	node, ok := resolved.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("component node must resolve to an object")
	}
	return node, nil
}

// materializeComponentNode 先替换参数占位符，再把模板里的 slot 占位展开成真实子节点。
//
// 组件模板仍然只在加载前预处理；slot 展开完成后，后续 loader 继续只面对普通节点树。
func materializeComponentNode(node map[string]any, args map[string]any, slots map[string][]map[string]any) (map[string]any, error) {
	resolved, err := substituteComponentValue(node, args)
	if err != nil {
		return nil, err
	}
	return applyComponentSlotsToNode(resolved, slots)
}

// applyComponentSlotsToNode 递归展开 children 数组里的 slot 占位。
//
// v1 只支持 children 位置的 slot，因此这里只对 children 执行“原地拼接”。
func applyComponentSlotsToNode(node map[string]any, slots map[string][]map[string]any) (map[string]any, error) {
	out := make(map[string]any, len(node))
	for key, item := range node {
		if key != "children" {
			out[key] = cloneJSONValue(item)
			continue
		}

		children, ok := item.([]any)
		if !ok {
			return nil, fmt.Errorf("node children must be an array")
		}
		expandedChildren := make([]any, 0, len(children))
		for i, child := range children {
			childNode, ok := child.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("children[%d] must be an object", i)
			}
			if slotName, isSlot, err := componentSlotName(childNode); err != nil {
				return nil, fmt.Errorf("children[%d]: %w", i, err)
			} else if isSlot {
				for _, slotNode := range slots[slotName] {
					expandedChildren = append(expandedChildren, cloneJSONObject(slotNode))
				}
				continue
			}

			expandedChild, err := applyComponentSlotsToNode(childNode, slots)
			if err != nil {
				return nil, fmt.Errorf("children[%d]: %w", i, err)
			}
			expandedChildren = append(expandedChildren, expandedChild)
		}
		out[key] = expandedChildren
	}
	return out, nil
}

func substituteComponentJSON(value any, args map[string]any) (any, error) {
	switch typed := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(typed))
		for key, item := range typed {
			resolved, err := substituteComponentJSON(item, args)
			if err != nil {
				return nil, err
			}
			out[key] = resolved
		}
		return out, nil
	case []any:
		out := make([]any, len(typed))
		for i, item := range typed {
			resolved, err := substituteComponentJSON(item, args)
			if err != nil {
				return nil, err
			}
			out[i] = resolved
		}
		return out, nil
	case string:
		name, isPlaceholder, err := parseComponentPlaceholder(typed)
		if err != nil {
			return nil, err
		}
		if !isPlaceholder {
			return typed, nil
		}
		value, ok := args[name]
		if !ok {
			return nil, fmt.Errorf("placeholder %q is not declared", name)
		}
		return cloneJSONValue(value), nil
	default:
		return typed, nil
	}
}

func parseComponentPlaceholder(value string) (string, bool, error) {
	if !strings.Contains(value, "${") {
		return "", false, nil
	}
	if strings.HasPrefix(value, "${") && strings.HasSuffix(value, "}") && strings.Count(value, "${") == 1 {
		name := strings.TrimSpace(value[2 : len(value)-1])
		if name == "" {
			return "", false, fmt.Errorf("component placeholder %q is empty", value)
		}
		return name, true, nil
	}
	return "", false, fmt.Errorf("component placeholder %q must occupy the entire JSON value", value)
}

func componentInstanceName(node map[string]any) (string, bool, error) {
	value, ok := node["component"]
	if !ok {
		return "", false, nil
	}
	name, ok := value.(string)
	if !ok {
		return "", false, fmt.Errorf("component must be a string")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return "", false, fmt.Errorf("component name is empty")
	}
	if _, ok := node["type"]; ok {
		return "", false, fmt.Errorf("component instance %q must not also declare type", name)
	}
	for key := range node {
		if key == "component" || key == "args" || key == "slots" {
			continue
		}
		return "", false, fmt.Errorf("component instance %q does not support field %q", name, key)
	}
	return name, true, nil
}

func componentSlotName(node map[string]any) (string, bool, error) {
	value, ok := node["slot"]
	if !ok {
		return "", false, nil
	}
	name, ok := value.(string)
	if !ok {
		return "", false, fmt.Errorf("slot must be a string")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return "", false, fmt.Errorf("slot name is empty")
	}
	for key := range node {
		if key == "slot" {
			continue
		}
		return "", false, fmt.Errorf("slot %q does not support field %q", name, key)
	}
	return name, true, nil
}

func decodeComponentArgs(value any) (map[string]any, error) {
	if value == nil {
		return map[string]any{}, nil
	}
	args, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("args must be an object")
	}
	return args, nil
}

func decodeComponentSlots(value any) (map[string][]map[string]any, error) {
	if value == nil {
		return map[string][]map[string]any{}, nil
	}
	raw, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("slots must be an object")
	}

	slots := make(map[string][]map[string]any, len(raw))
	for name, item := range raw {
		slotName := strings.TrimSpace(name)
		if slotName == "" {
			return nil, fmt.Errorf("slot name is empty")
		}
		children, ok := item.([]any)
		if !ok {
			return nil, fmt.Errorf("slot %q must be an array", slotName)
		}
		nodes := make([]map[string]any, 0, len(children))
		for i, child := range children {
			node, ok := child.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("slot %q item %d must be an object", slotName, i)
			}
			nodes = append(nodes, cloneJSONObject(node))
		}
		slots[slotName] = nodes
	}
	return slots, nil
}

// collectComponentSlots 收集模板中显式声明过的 slot。
//
// 当前版本只允许 slot 出现在 children 数组里；如果写到别的位置，直接报错。
func collectComponentSlots(node map[string]any) (map[string]struct{}, error) {
	slots := map[string]struct{}{}
	if err := visitComponentSlots(node, false, slots); err != nil {
		return nil, err
	}
	return slots, nil
}

func visitComponentSlots(value any, inChildren bool, slots map[string]struct{}) error {
	switch typed := value.(type) {
	case map[string]any:
		if name, isSlot, err := componentSlotName(typed); err != nil {
			return err
		} else if isSlot {
			if !inChildren {
				return fmt.Errorf("slot %q must appear inside children", name)
			}
			slots[name] = struct{}{}
			return nil
		}
		for key, item := range typed {
			if err := visitComponentSlots(item, key == "children", slots); err != nil {
				return err
			}
		}
	case []any:
		for _, item := range typed {
			if err := visitComponentSlots(item, inChildren, slots); err != nil {
				return err
			}
		}
	}
	return nil
}

func decodeComponentTemplateFile(data []byte) (componentTemplateFileSpec, error) {
	var spec componentTemplateFileSpec
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := decoder.Decode(&spec); err != nil {
		return componentTemplateFileSpec{}, err
	}
	return spec, nil
}

func decodeJSONObject(data []byte) (map[string]any, error) {
	var value map[string]any
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		return nil, err
	}
	return value, nil
}

func decodeJSONValue(data []byte) (any, error) {
	var value any
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		return nil, err
	}
	return value, nil
}

func parseComponentParamKind(value string) (componentParamKind, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "string":
		return componentParamString, nil
	case "int":
		return componentParamInt, nil
	case "bool":
		return componentParamBool, nil
	case "object":
		return componentParamObject, nil
	case "array":
		return componentParamArray, nil
	default:
		return componentParamAny, fmt.Errorf("unsupported param type %q", value)
	}
}

func inferComponentParamKind(value any) (componentParamKind, bool) {
	switch value.(type) {
	case string:
		return componentParamString, true
	case json.Number:
		return componentParamInt, true
	case bool:
		return componentParamBool, true
	case map[string]any:
		return componentParamObject, true
	case []any:
		return componentParamArray, true
	default:
		return componentParamAny, false
	}
}

func validateComponentValueType(name string, value any, param componentParam) error {
	if !param.HasKind || param.Kind == componentParamAny {
		return nil
	}
	if matchesComponentParamKind(value, param.Kind) {
		return nil
	}
	return fmt.Errorf("argument %q must be %s", name, componentParamKindName(param.Kind))
}

func matchesComponentParamKind(value any, kind componentParamKind) bool {
	switch kind {
	case componentParamString:
		_, ok := value.(string)
		return ok
	case componentParamInt:
		number, ok := value.(json.Number)
		if !ok {
			return false
		}
		_, err := number.Int64()
		return err == nil
	case componentParamBool:
		_, ok := value.(bool)
		return ok
	case componentParamObject:
		_, ok := value.(map[string]any)
		return ok
	case componentParamArray:
		_, ok := value.([]any)
		return ok
	default:
		return true
	}
}

func componentParamKindName(kind componentParamKind) string {
	switch kind {
	case componentParamString:
		return "string"
	case componentParamInt:
		return "int"
	case componentParamBool:
		return "bool"
	case componentParamObject:
		return "object"
	case componentParamArray:
		return "array"
	default:
		return "any"
	}
}

func appendComponentStack(stack []string, next string) ([]string, bool) {
	for _, item := range stack {
		if item == next {
			return append(append([]string{}, stack...), next), false
		}
	}
	return append(append([]string{}, stack...), next), true
}

func cloneJSONObject(value map[string]any) map[string]any {
	if value == nil {
		return nil
	}
	out := make(map[string]any, len(value))
	for key, item := range value {
		out[key] = cloneJSONValue(item)
	}
	return out
}

func cloneJSONValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return cloneJSONObject(typed)
	case []any:
		out := make([]any, len(typed))
		for i, item := range typed {
			out[i] = cloneJSONValue(item)
		}
		return out
	default:
		return typed
	}
}

func resolveComponentPath(baseDir, path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	if filepath.IsAbs(path) || baseDir == "" {
		return path
	}
	return filepath.Join(baseDir, path)
}
