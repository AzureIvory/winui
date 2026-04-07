//go:build windows

package markup

import (
	"fmt"
	"strings"
)

type nodeKind uint8

const (
	nodeElement nodeKind = iota + 1
	nodeText
)

type position struct {
	Line   int
	Column int
}

type node struct {
	Kind         nodeKind
	Tag          string
	Text         string
	Attrs        map[string]string
	Styles       map[string]string
	InlineStyles map[string]string
	Children     []*node
	Pos          position
}

func (n *node) attr(name string) string {
	if n == nil || n.Attrs == nil {
		return ""
	}
	return n.Attrs[strings.ToLower(strings.TrimSpace(name))]
}

func (n *node) hasAttr(name string) bool {
	if n == nil || n.Attrs == nil {
		return false
	}
	_, ok := n.Attrs[strings.ToLower(strings.TrimSpace(name))]
	return ok
}

func (n *node) elementChildren() []*node {
	if n == nil {
		return nil
	}
	out := make([]*node, 0, len(n.Children))
	for _, child := range n.Children {
		if child != nil && child.Kind == nodeElement {
			out = append(out, child)
		}
	}
	return out
}

func (n *node) textContent() string {
	if n == nil {
		return ""
	}
	if n.Kind == nodeText {
		return n.Text
	}
	var builder strings.Builder
	for _, child := range n.Children {
		builder.WriteString(child.textContent())
	}
	return builder.String()
}

func (n *node) inlineContext() string {
	if n == nil {
		return "node"
	}
	if n.Kind == nodeText {
		return "text"
	}
	if id := strings.TrimSpace(n.attr("id")); id != "" {
		return fmt.Sprintf("<%s id=%q>", n.Tag, id)
	}
	return fmt.Sprintf("<%s>", n.Tag)
}

type parseError struct {
	stage   string
	pos     position
	context string
	msg     string
}

func (e *parseError) Error() string {
	if e == nil {
		return ""
	}
	parts := make([]string, 0, 4)
	if e.stage != "" {
		parts = append(parts, e.stage)
	}
	if e.pos.Line > 0 {
		parts = append(parts, fmt.Sprintf("line %d:%d", e.pos.Line, e.pos.Column))
	}
	if e.context != "" {
		parts = append(parts, e.context)
	}
	if e.msg != "" {
		parts = append(parts, e.msg)
	}
	return strings.Join(parts, ": ")
}

func newParseError(stage string, pos position, context string, message any, args ...any) error {
	msg := fmt.Sprint(message)
	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	return &parseError{
		stage:   stage,
		pos:     pos,
		context: context,
		msg:     msg,
	}
}
