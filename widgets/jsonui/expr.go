//go:build windows

package jsonui

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// ExprAxis identifies which coordinate or size dimension is being resolved.
type ExprAxis uint8

const (
	AxisX ExprAxis = iota + 1
	AxisY
	AxisWidth
	AxisHeight
)

// ExprContext provides logical window and parent dimensions for expression resolution.
type ExprContext struct {
	WindowW int32
	WindowH int32
	ParentW int32
	ParentH int32
}

type exprBase uint8

const (
	exprLiteral exprBase = iota + 1
	exprPercent
	exprWindowW
	exprWindowH
	exprParentW
	exprParentH
)

// ScalarExpr describes one DPI-aware logical coordinate or size expression.
type ScalarExpr struct {
	base   exprBase
	value  int32
	offset int32
}

// ParseScalarExpr parses a JSON scalar expression.
//
// Supported values:
// - numbers like 100
// - strings like "100"
// - percentages like "50%"
// - percentages minus offsets like "50%-100"
// - window expressions like "winW-100" and "winH-100"
// - parent expressions like "parentW-100" and "parentH-100"
func ParseScalarExpr(input any) (ScalarExpr, error) {
	switch typed := input.(type) {
	case nil:
		return ScalarExpr{}, fmt.Errorf("scalar expression is nil")
	case int:
		return ScalarExpr{base: exprLiteral, value: int32(typed)}, nil
	case int32:
		return ScalarExpr{base: exprLiteral, value: typed}, nil
	case int64:
		return ScalarExpr{base: exprLiteral, value: int32(typed)}, nil
	case float64:
		if math.Trunc(typed) != typed {
			return ScalarExpr{}, fmt.Errorf("scalar expression %v must be an integer", typed)
		}
		return ScalarExpr{base: exprLiteral, value: int32(typed)}, nil
	case json.Number:
		number, err := typed.Int64()
		if err != nil {
			return ScalarExpr{}, fmt.Errorf("invalid scalar expression %q", typed)
		}
		return ScalarExpr{base: exprLiteral, value: int32(number)}, nil
	case string:
		return parseScalarExprString(typed)
	default:
		return ScalarExpr{}, fmt.Errorf("unsupported scalar expression type %T", input)
	}
}

func parseScalarExprString(input string) (ScalarExpr, error) {
	text := strings.TrimSpace(input)
	if text == "" {
		return ScalarExpr{}, fmt.Errorf("scalar expression is empty")
	}
	if number, err := strconv.Atoi(text); err == nil {
		return ScalarExpr{base: exprLiteral, value: int32(number)}, nil
	}

	offset := int32(0)
	head := text
	if dash := strings.LastIndex(text, "-"); dash > 0 {
		parsed, err := strconv.Atoi(strings.TrimSpace(text[dash+1:]))
		if err != nil {
			return ScalarExpr{}, fmt.Errorf("invalid scalar expression %q", input)
		}
		head = strings.TrimSpace(text[:dash])
		offset = int32(parsed)
	}

	switch {
	case strings.HasSuffix(head, "%"):
		percentText := strings.TrimSpace(strings.TrimSuffix(head, "%"))
		percent, err := strconv.Atoi(percentText)
		if err != nil {
			return ScalarExpr{}, fmt.Errorf("invalid scalar expression %q", input)
		}
		return ScalarExpr{base: exprPercent, value: int32(percent), offset: offset}, nil
	case head == "winW":
		return ScalarExpr{base: exprWindowW, offset: offset}, nil
	case head == "winH":
		return ScalarExpr{base: exprWindowH, offset: offset}, nil
	case head == "parentW":
		return ScalarExpr{base: exprParentW, offset: offset}, nil
	case head == "parentH":
		return ScalarExpr{base: exprParentH, offset: offset}, nil
	default:
		return ScalarExpr{}, fmt.Errorf("invalid scalar expression %q", input)
	}
}

// Resolve evaluates the expression against the provided logical dimensions.
func (e ScalarExpr) Resolve(axis ExprAxis, ctx ExprContext) int32 {
	var base int32

	switch e.base {
	case exprLiteral:
		base = e.value
	case exprPercent:
		base = axisWindowSize(axis, ctx) * e.value / 100
	case exprWindowW:
		base = ctx.WindowW
	case exprWindowH:
		base = ctx.WindowH
	case exprParentW:
		base = ctx.ParentW
	case exprParentH:
		base = ctx.ParentH
	default:
		return 0
	}

	return base - e.offset
}

func axisWindowSize(axis ExprAxis, ctx ExprContext) int32 {
	switch axis {
	case AxisY, AxisHeight:
		return ctx.WindowH
	default:
		return ctx.WindowW
	}
}
