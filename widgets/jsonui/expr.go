//go:build windows

package jsonui

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode"
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

type exprVar uint8

const (
	exprVarWindowW exprVar = iota + 1
	exprVarWindowH
	exprVarParentW
	exprVarParentH
)

type exprNodeKind uint8

const (
	exprNodeLiteral exprNodeKind = iota + 1
	exprNodePercent
	exprNodeVariable
	exprNodeAdd
	exprNodeSub
	exprNodeMul
	exprNodeDiv
	exprNodeNeg
)

type scalarExprNode struct {
	kind     exprNodeKind
	value    int32
	variable exprVar
	left     *scalarExprNode
	right    *scalarExprNode
}

// ScalarExpr describes one DPI-aware logical coordinate or size expression.
type ScalarExpr struct {
	root *scalarExprNode
}

// ParseScalarExpr parses a JSON scalar expression.
//
// Supported values:
// - numbers like 100
// - strings like "100"
// - arithmetic expressions using + - * / and parentheses
// - variables: winW, winH, parentW, parentH
// - percentages like "50%" that keep the legacy axis-based window percentage semantics
func ParseScalarExpr(input any) (ScalarExpr, error) {
	switch typed := input.(type) {
	case nil:
		return ScalarExpr{}, fmt.Errorf("scalar expression is nil")
	case int:
		return literalScalarExpr(int32(typed)), nil
	case int32:
		return literalScalarExpr(typed), nil
	case int64:
		return literalScalarExpr(int32(typed)), nil
	case float64:
		if math.Trunc(typed) != typed {
			return ScalarExpr{}, fmt.Errorf("scalar expression %v must be an integer", typed)
		}
		return literalScalarExpr(int32(typed)), nil
	case json.Number:
		number, err := typed.Int64()
		if err != nil {
			return ScalarExpr{}, fmt.Errorf("invalid scalar expression %q", typed)
		}
		return literalScalarExpr(int32(number)), nil
	case string:
		return parseScalarExprString(typed)
	default:
		return ScalarExpr{}, fmt.Errorf("unsupported scalar expression type %T", input)
	}
}

func literalScalarExpr(value int32) ScalarExpr {
	return ScalarExpr{
		root: &scalarExprNode{
			kind:  exprNodeLiteral,
			value: value,
		},
	}
}

func parseScalarExprString(input string) (ScalarExpr, error) {
	text := strings.TrimSpace(input)
	if text == "" {
		return ScalarExpr{}, fmt.Errorf("scalar expression is empty")
	}

	tokens, err := tokenizeScalarExpr(text)
	if err != nil {
		return ScalarExpr{}, fmt.Errorf("invalid scalar expression %q: %w", input, err)
	}

	parser := scalarExprParser{
		input:  text,
		tokens: tokens,
	}
	root, err := parser.parse()
	if err != nil {
		return ScalarExpr{}, fmt.Errorf("invalid scalar expression %q: %w", input, err)
	}
	return ScalarExpr{root: root}, nil
}

// Resolve evaluates the expression against the provided logical dimensions.
func (e ScalarExpr) Resolve(axis ExprAxis, ctx ExprContext) int32 {
	if e.root == nil {
		return 0
	}
	return clampResolvedScalar(e.root.resolve(axis, ctx))
}

func clampResolvedScalar(value int64) int32 {
	switch {
	case value > math.MaxInt32:
		return math.MaxInt32
	case value < math.MinInt32:
		return math.MinInt32
	default:
		return int32(value)
	}
}

func (n *scalarExprNode) resolve(axis ExprAxis, ctx ExprContext) int64 {
	if n == nil {
		return 0
	}

	switch n.kind {
	case exprNodeLiteral:
		return int64(n.value)
	case exprNodePercent:
		return int64(axisWindowSize(axis, ctx)) * int64(n.value) / 100
	case exprNodeVariable:
		return int64(resolveExprVariable(n.variable, ctx))
	case exprNodeAdd:
		return n.left.resolve(axis, ctx) + n.right.resolve(axis, ctx)
	case exprNodeSub:
		return n.left.resolve(axis, ctx) - n.right.resolve(axis, ctx)
	case exprNodeMul:
		return n.left.resolve(axis, ctx) * n.right.resolve(axis, ctx)
	case exprNodeDiv:
		denominator := n.right.resolve(axis, ctx)
		if denominator == 0 {
			return 0
		}
		return n.left.resolve(axis, ctx) / denominator
	case exprNodeNeg:
		return -n.left.resolve(axis, ctx)
	default:
		return 0
	}
}

func resolveExprVariable(variable exprVar, ctx ExprContext) int32 {
	switch variable {
	case exprVarWindowW:
		return ctx.WindowW
	case exprVarWindowH:
		return ctx.WindowH
	case exprVarParentW:
		return ctx.ParentW
	case exprVarParentH:
		return ctx.ParentH
	default:
		return 0
	}
}

func axisWindowSize(axis ExprAxis, ctx ExprContext) int32 {
	switch axis {
	case AxisY, AxisHeight:
		return ctx.WindowH
	default:
		return ctx.WindowW
	}
}

type scalarExprTokenKind uint8

const (
	scalarExprTokenEOF scalarExprTokenKind = iota
	scalarExprTokenNumber
	scalarExprTokenIdent
	scalarExprTokenPlus
	scalarExprTokenMinus
	scalarExprTokenStar
	scalarExprTokenSlash
	scalarExprTokenLParen
	scalarExprTokenRParen
	scalarExprTokenPercent
)

type scalarExprToken struct {
	kind   scalarExprTokenKind
	text   string
	number int32
}

func tokenizeScalarExpr(input string) ([]scalarExprToken, error) {
	tokens := make([]scalarExprToken, 0, len(input)/2+1)
	for index := 0; index < len(input); {
		ch := rune(input[index])
		if unicode.IsSpace(ch) {
			index++
			continue
		}

		switch ch {
		case '+':
			tokens = append(tokens, scalarExprToken{kind: scalarExprTokenPlus, text: "+"})
			index++
		case '-':
			tokens = append(tokens, scalarExprToken{kind: scalarExprTokenMinus, text: "-"})
			index++
		case '*':
			tokens = append(tokens, scalarExprToken{kind: scalarExprTokenStar, text: "*"})
			index++
		case '/':
			tokens = append(tokens, scalarExprToken{kind: scalarExprTokenSlash, text: "/"})
			index++
		case '(':
			tokens = append(tokens, scalarExprToken{kind: scalarExprTokenLParen, text: "("})
			index++
		case ')':
			tokens = append(tokens, scalarExprToken{kind: scalarExprTokenRParen, text: ")"})
			index++
		case '%':
			tokens = append(tokens, scalarExprToken{kind: scalarExprTokenPercent, text: "%"})
			index++
		default:
			switch {
			case ch >= '0' && ch <= '9':
				start := index
				for index < len(input) {
					digit := input[index]
					if digit < '0' || digit > '9' {
						break
					}
					index++
				}
				number, err := strconv.Atoi(input[start:index])
				if err != nil {
					return nil, err
				}
				tokens = append(tokens, scalarExprToken{
					kind:   scalarExprTokenNumber,
					text:   input[start:index],
					number: int32(number),
				})
			case unicode.IsLetter(ch):
				start := index
				for index < len(input) {
					part := rune(input[index])
					if !unicode.IsLetter(part) && !unicode.IsDigit(part) {
						break
					}
					index++
				}
				tokens = append(tokens, scalarExprToken{
					kind: scalarExprTokenIdent,
					text: input[start:index],
				})
			default:
				return nil, fmt.Errorf("unexpected character %q", ch)
			}
		}
	}

	tokens = append(tokens, scalarExprToken{kind: scalarExprTokenEOF})
	return tokens, nil
}

type scalarExprParser struct {
	input  string
	tokens []scalarExprToken
	pos    int
}

func (p *scalarExprParser) parse() (*scalarExprNode, error) {
	root, err := p.parseAddSub()
	if err != nil {
		return nil, err
	}
	if token := p.current(); token.kind != scalarExprTokenEOF {
		return nil, fmt.Errorf("unexpected token %q", token.text)
	}
	return root, nil
}

func (p *scalarExprParser) parseAddSub() (*scalarExprNode, error) {
	left, err := p.parseMulDiv()
	if err != nil {
		return nil, err
	}

	for {
		token := p.current()
		switch token.kind {
		case scalarExprTokenPlus, scalarExprTokenMinus:
			p.advance()
			right, err := p.parseMulDiv()
			if err != nil {
				return nil, err
			}
			kind := exprNodeAdd
			if token.kind == scalarExprTokenMinus {
				kind = exprNodeSub
			}
			left = &scalarExprNode{
				kind:  kind,
				left:  left,
				right: right,
			}
		default:
			return left, nil
		}
	}
}

func (p *scalarExprParser) parseMulDiv() (*scalarExprNode, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}

	for {
		token := p.current()
		switch token.kind {
		case scalarExprTokenStar, scalarExprTokenSlash:
			p.advance()
			right, err := p.parseUnary()
			if err != nil {
				return nil, err
			}
			kind := exprNodeMul
			if token.kind == scalarExprTokenSlash {
				kind = exprNodeDiv
			}
			left = &scalarExprNode{
				kind:  kind,
				left:  left,
				right: right,
			}
		default:
			return left, nil
		}
	}
}

func (p *scalarExprParser) parseUnary() (*scalarExprNode, error) {
	switch p.current().kind {
	case scalarExprTokenPlus:
		p.advance()
		return p.parseUnary()
	case scalarExprTokenMinus:
		p.advance()
		node, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return &scalarExprNode{
			kind: exprNodeNeg,
			left: node,
		}, nil
	default:
		return p.parsePrimary()
	}
}

func (p *scalarExprParser) parsePrimary() (*scalarExprNode, error) {
	token := p.current()
	switch token.kind {
	case scalarExprTokenNumber:
		p.advance()
		node := &scalarExprNode{
			kind:  exprNodeLiteral,
			value: token.number,
		}
		if p.current().kind == scalarExprTokenPercent {
			p.advance()
			node.kind = exprNodePercent
		}
		return node, nil
	case scalarExprTokenIdent:
		p.advance()
		variable, ok := parseExprVariable(token.text)
		if !ok {
			return nil, fmt.Errorf("unknown identifier %q", token.text)
		}
		if p.current().kind == scalarExprTokenPercent {
			return nil, fmt.Errorf("percent can only follow integer literals")
		}
		return &scalarExprNode{
			kind:     exprNodeVariable,
			variable: variable,
		}, nil
	case scalarExprTokenLParen:
		p.advance()
		node, err := p.parseAddSub()
		if err != nil {
			return nil, err
		}
		if err := p.expect(scalarExprTokenRParen); err != nil {
			return nil, err
		}
		if p.current().kind == scalarExprTokenPercent {
			return nil, fmt.Errorf("percent can only follow integer literals")
		}
		return node, nil
	default:
		if token.kind == scalarExprTokenEOF {
			return nil, fmt.Errorf("unexpected end of expression")
		}
		return nil, fmt.Errorf("unexpected token %q", token.text)
	}
}

func parseExprVariable(text string) (exprVar, bool) {
	switch text {
	case "winW":
		return exprVarWindowW, true
	case "winH":
		return exprVarWindowH, true
	case "parentW":
		return exprVarParentW, true
	case "parentH":
		return exprVarParentH, true
	default:
		return 0, false
	}
}

func (p *scalarExprParser) expect(kind scalarExprTokenKind) error {
	if p.current().kind != kind {
		token := p.current()
		if token.kind == scalarExprTokenEOF {
			return fmt.Errorf("unexpected end of expression")
		}
		return fmt.Errorf("unexpected token %q", token.text)
	}
	p.advance()
	return nil
}

func (p *scalarExprParser) current() scalarExprToken {
	if p.pos >= len(p.tokens) {
		return scalarExprToken{kind: scalarExprTokenEOF}
	}
	return p.tokens[p.pos]
}

func (p *scalarExprParser) advance() {
	if p.pos < len(p.tokens) {
		p.pos++
	}
}
