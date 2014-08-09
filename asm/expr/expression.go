package expr

import (
	"errors"
	"fmt"

	"github.com/zellyn/go6502/asm/context"
	"github.com/zellyn/go6502/asm/lines"
)

type UnknownLabelError struct {
	Err error
}

func (e UnknownLabelError) Error() string {
	return e.Err.Error()
}

type Operator int

const (
	OpUnknown Operator = iota
	OpLeaf             // No op - this is a leaf node
	OpPlus             // add
	OpMinus            // subtract / negate (with Right==nil)
	OpMul              // multiply
	OpDiv              // divide
	OpLsb              // Low byte
	OpMsb              // High byte
	OpLt               // Less than
	OpGt               // Greater than
	OpEq               // Equal to
	OpByte             // A single byte of storage
	OpAnd
	OpOr
	OpXor
)

var OpStrings = map[Operator]string{
	OpPlus:  "+",
	OpMinus: "-",
	OpMul:   "*",
	OpDiv:   "/",
	OpLsb:   "lsb",
	OpMsb:   "msb",
	OpByte:  "byte",
	OpLt:    "<",
	OpGt:    ">",
	OpEq:    "=",
	OpAnd:   "&",
	OpOr:    "|",
	OpXor:   "^",
}

type E struct {
	Left  *E
	Right *E
	Op    Operator
	Text  string
	Val   uint16
}

func (e E) String() string {
	switch e.Op {
	case OpLeaf:
		if e.Text != "" {
			return e.Text
		}
		return fmt.Sprintf("$%04x", e.Val)

	case OpPlus, OpMinus, OpMul, OpDiv, OpLsb, OpMsb, OpByte, OpLt, OpGt, OpEq, OpAnd, OpOr, OpXor:
		if e.Right != nil {
			return fmt.Sprintf("(%s %s %s)", OpStrings[e.Op], *e.Left, *e.Right)
		}
		return fmt.Sprintf("(%s %s)", OpStrings[e.Op], *e.Left)
	case OpUnknown:
		return "?"
	default:
		panic(fmt.Sprintf("Unprintable op type: %v", e.Op))
	}
}

func (e *E) Equal(o *E) bool {
	if e == nil || o == nil {
		return e == nil && o == nil
	}
	if e.Op != o.Op {
		return false
	}
	if e.Text != o.Text {
		return false
	}
	if e.Val != o.Val {
		return false
	}
	return e.Left.Equal(o.Left) && e.Right.Equal(o.Right)
}

// Width returns the width in bytes of an expression. It'll be two for anything except
// expressions that start with Lsb or Msb operators.
func (e *E) Width() uint16 {
	switch e.Op {
	case OpLsb, OpMsb, OpByte:
		return 1
	}
	return 2
}

func (e *E) Eval(ctx context.Context, ln *lines.Line) (uint16, error) {
	if e == nil {
		return 0, errors.New("cannot Eval() nil expression")
	}
	switch e.Op {
	case OpLeaf:
		if e.Text == "" {
			return e.Val, nil
		}
		if val, ok := ctx.Get(e.Text); ok {
			return val, nil
		}
		return 0, UnknownLabelError{Err: ln.Errorf("unknown label: %s", e.Text)}
	case OpMinus:
		l, err := e.Left.Eval(ctx, ln)
		if err != nil {
			return 0, err
		}
		if e.Right == nil {
			return -l, nil
		}
		r, err := e.Right.Eval(ctx, ln)
		if err != nil {
			return 0, err
		}
		return l - r, nil
	case OpMsb, OpLsb:
		l, err := e.Left.Eval(ctx, ln)
		if err != nil {
			return 0, err
		}
		if e.Op == OpMsb {
			return l >> 8, nil
		}
		return l & 0xff, nil
	case OpByte:
		return e.Val, nil
	case OpPlus, OpMul, OpDiv, OpLt, OpGt, OpEq, OpAnd, OpOr, OpXor:
		l, err := e.Left.Eval(ctx, ln)
		if err != nil {
			return 0, err
		}
		r, err := e.Right.Eval(ctx, ln)
		if err != nil {
			return 0, err
		}
		switch e.Op {
		case OpPlus:
			return l + r, nil
		case OpMul:
			return l * r, nil
		case OpLt:
			if l < r {
				return 1, nil
			}
			return 0, nil
		case OpGt:
			if l > r {
				return 1, nil
			}
			return 0, nil
		case OpEq:
			if l == r {
				return 1, nil
			}
			return 0, nil
		case OpDiv:
			if r == 0 {
				return ctx.Zero()
			}
			return l / r, nil
		case OpAnd:
			return l & r, nil
		case OpOr:
			return l | r, nil
		case OpXor:
			return l ^ r, nil
		}
		panic(fmt.Sprintf("bad code - missing switch case: %d", e.Op))
	}
	panic(fmt.Sprintf("unknown operator type: %d", e.Op))
}

// CheckedEval calls Eval, but also turns UnknownLabelErrors into labelMissing booleans.
func (e *E) CheckedEval(ctx context.Context, ln *lines.Line) (val uint16, labelMissing bool, err error) {
	val, err = e.Eval(ctx, ln)
	switch err.(type) {
	case nil:
		return val, false, nil
	case UnknownLabelError:
		return val, true, err
	}
	return val, false, err
}

// FixLabels attempts to turn .1 into LAST_LABEL.1, etc.
func (e *E) FixLabels(labeler context.Labeler, macroCall int, locals map[string]bool, ln *lines.Line) error {
	newL, err := labeler.FixLabel(e.Text, macroCall, locals)
	if err != nil {
		return ln.Errorf("%v", err)
	}
	e.Text = newL

	if e.Left != nil {
		if err := e.Left.FixLabels(labeler, macroCall, locals, ln); err != nil {
			return err
		}
	}
	if e.Right != nil {
		return e.Right.FixLabels(labeler, macroCall, locals, ln)
	}

	return nil
}
