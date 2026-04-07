//go:build windows

package markup

import (
	"encoding/xml"
	"io"
	"regexp"
	"strings"
)

type cssRule struct {
	selectors    []cssSelector
	declarations map[string]string
	order        int
}

type cssSelector struct {
	tag         string
	id          string
	classes     []string
	specificity int
}

var cssCommentPattern = regexp.MustCompile(`(?s)/\*.*?\*/`)

func parseHTMLDocument(htmlText string) (*node, error) {
	decoder := xml.NewDecoder(strings.NewReader(htmlText))
	decoder.Strict = false
	var root *node
	stack := make([]*node, 0)
	for {
		token, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			line, col := decoder.InputPos()
			return nil, newParseError("html", position{Line: line, Column: col}, "document", err.Error())
		}
		line, col := decoder.InputPos()
		pos := position{Line: line, Column: col}
		switch typed := token.(type) {
		case xml.StartElement:
			n := &node{
				Kind:         nodeElement,
				Tag:          strings.ToLower(typed.Name.Local),
				Attrs:        make(map[string]string),
				Styles:       make(map[string]string),
				InlineStyles: make(map[string]string),
				Pos:          pos,
			}
			for _, attr := range typed.Attr {
				key := strings.ToLower(strings.TrimSpace(attr.Name.Local))
				n.Attrs[key] = attr.Value
			}
			if inline := strings.TrimSpace(n.Attrs["style"]); inline != "" {
				n.InlineStyles = parseStyleDeclarations(inline)
			}
			if len(stack) > 0 {
				parent := stack[len(stack)-1]
				parent.Children = append(parent.Children, n)
			} else if root == nil {
				root = n
			} else {
				return nil, newParseError("html", pos, n.inlineContext(), "multiple root elements are not supported")
			}
			stack = append(stack, n)
		case xml.EndElement:
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
		case xml.CharData:
			if len(stack) == 0 {
				if strings.TrimSpace(string(typed)) == "" {
					continue
				}
				return nil, newParseError("html", pos, "document", "text is not allowed outside the root element")
			}
			parent := stack[len(stack)-1]
			parent.Children = append(parent.Children, &node{Kind: nodeText, Text: string(typed), Pos: pos})
		}
	}
	if root == nil {
		return nil, newParseError("html", position{}, "document", "empty document")
	}
	return root, nil
}

func parseCSS(cssText string) ([]cssRule, error) {
	clean := cssCommentPattern.ReplaceAllString(cssText, "")
	rules := make([]cssRule, 0)
	order := 0
	for {
		clean = strings.TrimSpace(clean)
		if clean == "" {
			break
		}
		open := strings.IndexByte(clean, '{')
		close := strings.IndexByte(clean, '}')
		if open <= 0 || close <= open {
			return nil, newParseError("css", position{}, "stylesheet", "invalid rule syntax")
		}
		selectorText := strings.TrimSpace(clean[:open])
		body := clean[open+1 : close]
		clean = clean[close+1:]
		selectors, err := parseSelectors(selectorText)
		if err != nil {
			return nil, err
		}
		rules = append(rules, cssRule{
			selectors:    selectors,
			declarations: parseStyleDeclarations(body),
			order:        order,
		})
		order++
	}
	return rules, nil
}

func parseSelectors(selectorText string) ([]cssSelector, error) {
	parts := strings.Split(selectorText, ",")
	selectors := make([]cssSelector, 0, len(parts))
	for _, part := range parts {
		selector, err := parseSelector(part)
		if err != nil {
			return nil, err
		}
		selectors = append(selectors, selector)
	}
	return selectors, nil
}

func parseSelector(selectorText string) (cssSelector, error) {
	text := strings.TrimSpace(selectorText)
	if text == "" {
		return cssSelector{}, newParseError("css", position{}, "selector", "empty selector")
	}
	if strings.ContainsAny(text, " >+~") {
		return cssSelector{}, newParseError("css", position{}, text, "complex selectors are not supported")
	}
	selector := cssSelector{}
	for len(text) > 0 {
		switch text[0] {
		case '#':
			text = text[1:]
			value, rest := readSelectorIdent(text)
			if value == "" {
				return cssSelector{}, newParseError("css", position{}, selectorText, "invalid id selector")
			}
			selector.id = value
			selector.specificity += 100
			text = rest
		case '.':
			text = text[1:]
			value, rest := readSelectorIdent(text)
			if value == "" {
				return cssSelector{}, newParseError("css", position{}, selectorText, "invalid class selector")
			}
			selector.classes = append(selector.classes, value)
			selector.specificity += 10
			text = rest
		default:
			value, rest := readSelectorIdent(text)
			if value == "" {
				return cssSelector{}, newParseError("css", position{}, selectorText, "invalid tag selector")
			}
			selector.tag = strings.ToLower(value)
			selector.specificity++
			text = rest
		}
	}
	return selector, nil
}

func readSelectorIdent(text string) (string, string) {
	index := 0
	for index < len(text) {
		ch := text[index]
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '-' || ch == '_' {
			index++
			continue
		}
		break
	}
	return text[:index], text[index:]
}

func applyCSS(root *node, rules []cssRule) error {
	applyCSSNode(root, rules)
	return nil
}

func applyCSSNode(n *node, rules []cssRule) {
	if n == nil {
		return
	}
	if n.Kind == nodeElement {
		type applied struct {
			specificity int
			order       int
			value       string
		}
		resolved := make(map[string]applied)
		for _, rule := range rules {
			matched := false
			best := 0
			for _, selector := range rule.selectors {
				if selector.matches(n) {
					matched = true
					if selector.specificity > best {
						best = selector.specificity
					}
				}
			}
			if !matched {
				continue
			}
			for key, value := range rule.declarations {
				current, ok := resolved[key]
				if !ok || best > current.specificity || (best == current.specificity && rule.order >= current.order) {
					resolved[key] = applied{specificity: best, order: rule.order, value: value}
				}
			}
		}
		n.Styles = make(map[string]string, len(resolved)+len(n.InlineStyles))
		for key, value := range resolved {
			n.Styles[key] = value.value
		}
		for key, value := range n.InlineStyles {
			n.Styles[key] = value
		}
	}
	for _, child := range n.Children {
		applyCSSNode(child, rules)
	}
}

func (s cssSelector) matches(n *node) bool {
	if n == nil || n.Kind != nodeElement {
		return false
	}
	if s.tag != "" && s.tag != n.Tag {
		return false
	}
	if s.id != "" && s.id != strings.TrimSpace(n.attr("id")) {
		return false
	}
	if len(s.classes) == 0 {
		return true
	}
	classes := map[string]struct{}{}
	for _, className := range strings.Fields(strings.TrimSpace(n.attr("class"))) {
		classes[className] = struct{}{}
	}
	for _, className := range s.classes {
		if _, ok := classes[className]; !ok {
			return false
		}
	}
	return true
}

func parseStyleDeclarations(text string) map[string]string {
	decls := make(map[string]string)
	for _, part := range strings.Split(text, ";") {
		entry := strings.TrimSpace(part)
		if entry == "" {
			continue
		}
		pieces := strings.SplitN(entry, ":", 2)
		if len(pieces) != 2 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(pieces[0]))
		value := strings.TrimSpace(pieces[1])
		if key != "" {
			decls[key] = value
		}
	}
	return decls
}
