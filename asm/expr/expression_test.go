package expr

import (
	"testing"
)

func TestExpressionString(t *testing.T) {
	tests := []struct {
		expr E
		want string
	}{
		{
			E{},
			"?",
		},
		{
			E{Op: OpLeaf, Text: "*"},
			"*",
		},
		{
			E{
				Op:    OpPlus,
				Left:  &E{Op: OpLeaf, Text: "Label"},
				Right: &E{Op: OpLeaf, Val: 42},
			},
			"(+ Label $002a)",
		},
		{
			E{
				Op:    OpMinus,
				Left:  &E{Op: OpLeaf, Text: "Label"},
				Right: &E{Op: OpLeaf, Val: 42},
			},
			"(- Label $002a)",
		},
		{
			E{
				Op:   OpMinus,
				Left: &E{Op: OpLeaf, Val: 42},
			},
			"(- $002a)",
		},
	}
	for i, tt := range tests {
		got := tt.expr.String()
		if got != tt.want {
			t.Errorf(`%d: want String(expr)="%s"; got "%s"`, i, tt.want, got)
		}
	}
}
