//go:build windows

package jsonui

import "testing"

func TestParseScalarExprResolve(t *testing.T) {
	tests := []struct {
		name  string
		input any
		ctx   ExprContext
		axis  ExprAxis
		want  int32
	}{
		{
			name:  "plain number stays logical",
			input: 100,
			ctx: ExprContext{
				WindowW: 300,
				WindowH: 200,
				ParentW: 300,
				ParentH: 200,
			},
			axis: AxisX,
			want: 100,
		},
		{
			name:  "percent of window width",
			input: "50%",
			ctx: ExprContext{
				WindowW: 300,
				WindowH: 200,
				ParentW: 240,
				ParentH: 180,
			},
			axis: AxisX,
			want: 150,
		},
		{
			name:  "percent minus offset",
			input: "50%-100",
			ctx: ExprContext{
				WindowW: 300,
				WindowH: 200,
				ParentW: 300,
				ParentH: 200,
			},
			axis: AxisX,
			want: 50,
		},
		{
			name:  "window width minus offset",
			input: "winW-100",
			ctx: ExprContext{
				WindowW: 300,
				WindowH: 200,
				ParentW: 160,
				ParentH: 120,
			},
			axis: AxisWidth,
			want: 200,
		},
		{
			name:  "window height minus offset",
			input: "winH-100",
			ctx: ExprContext{
				WindowW: 300,
				WindowH: 220,
				ParentW: 160,
				ParentH: 120,
			},
			axis: AxisHeight,
			want: 120,
		},
		{
			name:  "parent width minus offset",
			input: "parentW-24",
			ctx: ExprContext{
				WindowW: 500,
				WindowH: 320,
				ParentW: 280,
				ParentH: 200,
			},
			axis: AxisWidth,
			want: 256,
		},
		{
			name:  "parent height minus offset",
			input: "parentH-18",
			ctx: ExprContext{
				WindowW: 500,
				WindowH: 320,
				ParentW: 280,
				ParentH: 210,
			},
			axis: AxisHeight,
			want: 192,
		},
		{
			name:  "percent height uses axis base",
			input: "25%",
			ctx: ExprContext{
				WindowW: 500,
				WindowH: 320,
				ParentW: 280,
				ParentH: 200,
			},
			axis: AxisHeight,
			want: 80,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := ParseScalarExpr(tt.input)
			if err != nil {
				t.Fatalf("ParseScalarExpr(%v) returned error: %v", tt.input, err)
			}
			got := expr.Resolve(tt.axis, tt.ctx)
			if got != tt.want {
				t.Fatalf("Resolve(%v) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseScalarExprRejectsInvalidInput(t *testing.T) {
	tests := []any{
		"",
		"abc",
		"50%+12",
		"winW+12",
		"parentX-10",
		[]int{1, 2, 3},
	}

	for _, input := range tests {
		t.Run("reject", func(t *testing.T) {
			if _, err := ParseScalarExpr(input); err == nil {
				t.Fatalf("ParseScalarExpr(%v) unexpectedly succeeded", input)
			}
		})
	}
}
