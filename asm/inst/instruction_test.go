package inst

import (
	"testing"

	"github.com/zellyn/go6502/asm/context"
	"github.com/zellyn/go6502/asm/expr"
)

func TestComputeLabel(t *testing.T) {
	i := I{
		Label: "L1",
	}
	c := &context.SimpleContext{}
	i.computeLabel(c, false, false)
}

func TestWidthDoesNotChange(t *testing.T) {
	i := I{
		Type:    TypeOp,
		Command: "LDA",
		Exprs: []*expr.E{
			&expr.E{
				Op:    expr.OpMinus,
				Left:  &expr.E{Op: expr.OpLeaf, Text: "L1"},
				Right: &expr.E{Op: expr.OpLeaf, Text: "L2"},
			},
		},
		MinWidth: 0x2,
		MaxWidth: 0x3,
		Final:    false,
		Op:       0xad,
		Mode:     0x2,
		ZeroMode: 0x80,
		ZeroOp:   0xa5,
	}
	c := &context.SimpleContext{}
	c.Set("L1", 0x102)
	final, err := i.Compute(c, false, false)
	if err != nil {
		t.Fatal(err)
	}
	if final {
		t.Fatal("First pass shouldn't be able to finalize expression with unknown width")
	}

	final, err = i.Compute(c, true, false)
	if err != nil {
		t.Fatal(err)
	}
	if final {
		t.Fatal("Second pass shouldn't be able to finalize expression with unknown variable.")
	}
	if !i.WidthKnown {
		t.Fatal("Second pass should have set width.")
	}
	if i.MinWidth != i.MaxWidth {
		t.Fatalf("i.WidthKnown, but i.MinWidth(%d) != i.MaxWidth(%d)", i.MinWidth, i.MaxWidth)
	}
	if i.MinWidth != 3 {
		t.Fatalf("i.MinWidth should be 3; got %d", i.MinWidth)
	}

	c.Set("L2", 0x101)

	final, err = i.Compute(c, true, true)
	if err != nil {
		t.Fatal(err)
	}
	if !final {
		t.Fatal("Third pass should be able to finalize expression.")
	}
	if !i.WidthKnown {
		t.Fatal("Third pass should left width unchanged.")
	}
	if i.MinWidth != i.MaxWidth {
		t.Fatalf("i.WidthKnown, but i.MinWidth(%d) != i.MaxWidth(%d)", i.MinWidth, i.MaxWidth)
	}
	if i.MinWidth != 3 {
		t.Fatalf("i.MinWidth should still be 3; got %d", i.MinWidth)
	}
}
