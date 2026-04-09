//go:build windows

package jsonui

import (
	"strings"

	"github.com/AzureIvory/winui/widgets"
)

type windowBinding struct {
	paths []string
	apply func(*bindingContext)
}

type bindingContext struct {
	window *Window
	scene  *widgets.Scene
	data   DataSource
}

func (c *bindingContext) Lookup(path string) (any, bool) {
	if c == nil || c.data == nil {
		return nil, false
	}
	return c.data.Get(path)
}

func (b windowBinding) matches(paths []string) bool {
	if len(b.paths) == 0 || len(paths) == 0 {
		return true
	}
	for _, changed := range paths {
		changed = normalizeBindingPath(changed)
		if changed == "" {
			return true
		}
		for _, path := range b.paths {
			if bindingPathsOverlap(path, changed) {
				return true
			}
		}
	}
	return false
}

func bindingPathsOverlap(left string, right string) bool {
	left = normalizeBindingPath(left)
	right = normalizeBindingPath(right)
	if left == "" || right == "" {
		return true
	}
	return left == right ||
		strings.HasPrefix(left, right+".") ||
		strings.HasPrefix(right, left+".")
}
