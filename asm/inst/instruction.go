package inst

import (
	"errors"
	"fmt"
	"strings"

	"github.com/zellyn/go6502/asm/context"
	"github.com/zellyn/go6502/asm/expr"
	"github.com/zellyn/go6502/asm/lines"
	"github.com/zellyn/go6502/opcodes"
)

type Type int

const (
	TypeUnknown Type = iota

	TypeNone       // No type (eg. just a label, just a comment, empty, ignored directive)
	TypeMacroStart // Start of macro definition
	TypeMacroEnd   // End of macro definition
	TypeMacroCall  // Macro invocation
	TypeMacroExit  // Macro early exit
	TypeIfdef      // Ifdef start
	TypeIfdefElse  // Ifdef else block
	TypeIfdefEnd   // Ifdef end
	TypeInclude    // Include a file
	TypeData       // Data
	TypeBlock      // Block storage
	TypeOrg        // Where to store assembled code
	TypeTarget     // Target address to use for jumps, labels, etc.
	TypeSegment    // Segments: bss/code/data
	TypeEqu        // Equate
	TypeOp         // An actual asm opcode
	TypeEnd        // End assembly
	TypeSetting    // An on/off setting toggle
)

// Variants for instructions. These tell the instruction how to
// interpret the raw data that comes in on the first or second pass.
const (
	VarBytes       = iota // Data: expressions, but forced to one byte per
	VarMixed              // Bytes or words (LE), depending on individual expression widths
	VarWordsLe            // Data: expressions, but forced to one word per, little-endian
	VarWordsBe            // Data: expressions, but forced to one word per, big-endian
	VarAscii              // Data: from ASCII strings, high bit clear
	VarAsciiFlip          // Data: from ASCII strings, high bit clear, except last char
	VarAsciiHi            // Data: from ASCII strings, high bit set
	VarAsciiHiFlip        // Data: from ASCII strings, high bit set, except last char
	VarRelative           // For branches: a one-byte relative address
	VarEquNormal          // Equ: a normal equate
	VarEquPageZero        // Equ: a page-zero equate
)

type I struct {
	Type         Type                   // Type of instruction
	Label        string                 // Text of label part
	MacroArgs    []string               // Macro args
	Command      string                 // Text of command part
	TextArg      string                 // Text of argument part
	Exprs        []*expr.E              // Expression(s)
	Data         []byte                 // Actual bytes
	WidthKnown   bool                   // Do we know how many bytes this instruction takes yet?
	Width        uint16                 // width in bytes
	Final        bool                   // Do we know the actual bytes yet?
	Op           byte                   // Opcode
	Mode         opcodes.AddressingMode // Opcode mode
	ZeroMode     opcodes.AddressingMode // Possible ZP-option mode
	ZeroOp       byte                   // Possible ZP-option Opcode
	Value        uint16                 // For Equates, the value
	DeclaredLine uint16                 // Line number listed in file
	Line         *lines.Line            // Line object for this line
	Addr         uint16                 // Current memory address
	Var          int                    // Variant of instruction type
}

func (i I) TypeString() string {
	switch i.Type {
	case TypeNone:
		return "-"
	case TypeMacroStart:
		return "macro"
	case TypeMacroEnd:
		return "endm"
	case TypeMacroCall:
		return "call " + i.Command
	case TypeMacroExit:
		return "exitm"
	case TypeIfdef:
		return "if"
	case TypeIfdefElse:
		return "else"
	case TypeIfdefEnd:
		return "endif"
	case TypeInclude:
		return "inc"
	case TypeData:
		switch i.Var {
		case VarMixed:
			return "data"
		case VarBytes:
			return "data/b"
		case VarWordsLe:
			return "data/wle"
		case VarWordsBe:
			return "data/wbe"
		case VarAscii, VarAsciiHi, VarAsciiFlip, VarAsciiHiFlip:
			return "data/b"
		default:
			panic(fmt.Sprintf("unknown data variant: %d", i.Var))
		}
	case TypeBlock:
		return "block"
	case TypeOrg:
		return "org"
	case TypeTarget:
		return "target"
	case TypeSegment:
		return "seg"
	case TypeEqu:
		return "="
	case TypeEnd:
		return "end"
	case TypeSetting:
		return "set"
	case TypeOp:
		modeStr := "?"
		switch i.Mode {
		case opcodes.MODE_IMPLIED:
			modeStr = "imp"
		case opcodes.MODE_ABSOLUTE:
			modeStr = "abs"
		case opcodes.MODE_INDIRECT:
			modeStr = "ind"
		case opcodes.MODE_RELATIVE:
			modeStr = "rel"
		case opcodes.MODE_IMMEDIATE:
			modeStr = "imm"
		case opcodes.MODE_ABS_X:
			modeStr = "absX"
		case opcodes.MODE_ABS_Y:
			modeStr = "absY"
		case opcodes.MODE_ZP:
			modeStr = "zp"
		case opcodes.MODE_ZP_X:
			modeStr = "zpX"
		case opcodes.MODE_ZP_Y:
			modeStr = "zpY"
		case opcodes.MODE_INDIRECT_Y:
			modeStr = "indY"
		case opcodes.MODE_INDIRECT_X:
			modeStr = "indX"
		case opcodes.MODE_A:
			modeStr = "a"

		}
		return fmt.Sprintf("%s/%s", i.Command, modeStr)
	}
	return "?"
}

func (i I) String() string {
	switch i.Type {
	case TypeInclude:
		return fmt.Sprintf("{inc '%s'}", i.TextArg)
	case TypeSetting:
		return fmt.Sprintf("{set %s %s}", i.Command, i.TextArg)
	}
	s := "{" + i.TypeString()
	if i.Label != "" {
		s += fmt.Sprintf(" '%s'", i.Label)
	}
	if i.TextArg != "" {
		s += fmt.Sprintf(` "%s"`, i.TextArg)
	}
	if len(i.MacroArgs) > 0 {
		ma := fmt.Sprintf("%#v", i.MacroArgs)[8:]
		s += " " + ma
	}
	if len(i.Exprs) > 0 {
		exprs := []string{}
		for _, expr := range i.Exprs {
			exprs = append(exprs, expr.String())
		}
		s += " " + strings.Join(exprs, ",")
	}
	return s + "}"
}

// Compute attempts to finalize the instruction.
func (i *I) Compute(c context.Context, final bool) (bool, error) {
	if i.Type == TypeEqu || i.Type == TypeTarget || i.Type == TypeOrg {
		return i.computeMustKnow(c, final)
	}
	if err := i.computeLabel(c, final); err != nil {
		return false, err
	}
	if i.Final {
		return true, nil
	}
	switch i.Type {
	case TypeOp:
		return i.computeOp(c, final)
	case TypeData:
		return i.computeData(c, final)
	case TypeBlock:
		return i.computeBlock(c, final)
	}

	// Everything else is zero-width
	i.WidthKnown = true
	i.Width = 0
	i.Final = true

	return true, nil
}

// computeLabel attempts to compute equates and label values.
func (i *I) computeLabel(c context.Context, final bool) error {
	if i.Label == "" {
		return nil
	}

	if i.Type == TypeEqu {
		panic("computeLabel should not be called for equates")
	}

	addr, aok := c.GetAddr()
	if !aok {
		return nil
	}

	lval, lok := c.Get(i.Label)
	if lok && addr != lval {
		return i.Errorf("Trying to set label '%s' to $%04x, but it already has value $%04x", i.Label, addr, lval)
	}
	c.Set(i.Label, addr)
	return nil
}

func (i *I) computeData(c context.Context, final bool) (bool, error) {
	if len(i.Data) > 0 {
		i.WidthKnown = true
		i.Width = uint16(len(i.Data))
		i.Final = true
		return true, nil
	}

	allFinal := true
	data := []byte{}
	var width uint16
	for _, e := range i.Exprs {
		var w uint16
		switch i.Var {
		case VarMixed:
			w = e.Width()
		case VarBytes:
			w = 1
		case VarWordsLe, VarWordsBe:
			w = 2
		}
		width += w
		val, labelMissing, err := e.CheckedEval(c, i.Line)
		if err != nil && !labelMissing {
			return false, err
		}
		if labelMissing {
			allFinal = false
			if final {
				return false, err
			}
		}
		switch i.Var {
		case VarMixed:
			switch w {
			case 1:
				data = append(data, byte(val))
			case 2:
				data = append(data, byte(val), byte(val>>8))
			}
		case VarBytes:
			data = append(data, byte(val))
		case VarWordsLe:
			data = append(data, byte(val), byte(val>>8))
		case VarWordsBe:
			data = append(data, byte(val>>8), byte(val))
		default:
			panic(fmt.Sprintf("Unknown data variant handed to computeData: %d", i.Var))
		}
	}
	i.Width = width
	i.WidthKnown = true
	if allFinal {
		i.Data = data
		i.Final = true
	}

	return i.Final, nil
}

func (i *I) computeBlock(c context.Context, final bool) (bool, error) {
	val, err := i.Exprs[0].Eval(c, i.Line)
	if err == nil {
		i.Value = val
		i.WidthKnown = true
		i.Final = true
		i.Width = val
	} else {
		if final {
			return false, i.Errorf("block storage with unknown size")
		}
	}
	return i.Final, nil
}

func (i *I) computeMustKnow(c context.Context, final bool) (bool, error) {
	i.WidthKnown = true
	i.Width = 0
	i.Final = true
	val, err := i.Exprs[0].Eval(c, i.Line)
	if err != nil {
		return false, err
	}
	i.Value = val
	switch i.Type {
	case TypeTarget:
		return false, errors.New("Target not implemented yet.")
	case TypeOrg:
		c.SetAddr(val)
	case TypeEqu:
		c.Set(i.Label, val)
		// Don't handle labels.
		return true, nil
	}
	if err := i.computeLabel(c, final); err != nil {
		return false, err
	}
	return true, nil
}

func (i *I) computeOp(c context.Context, final bool) (bool, error) {
	// If the width is not known, we better have a ZeroPage alternative.
	if !i.WidthKnown && (i.ZeroOp == 0 || i.ZeroMode == 0) {
		if i.Line.Context != nil && i.Line.Context.Parent != nil {
			fmt.Println(i.Line.Context.Parent.Sprintf("foo"))
		}
		panic(i.Sprintf("Reached computeOp for '%s' with no ZeroPage alternative: %#v [%s]", i.Command, i, i.Line.Text()))
	}
	// An op with no args would be final already, so we must have an Expression.
	if len(i.Exprs) == 0 {
		panic(fmt.Sprintf("Reached computeOp for '%s' with no expressions", i.Command))
	}
	val, labelMissing, err := i.Exprs[0].CheckedEval(c, i.Line)
	// A real error
	if !labelMissing && err != nil {
		return false, err
	}

	if labelMissing {
		// Don't know, do care.
		if final {
			return false, err
		}

		// Already know enough.
		if i.WidthKnown {
			return false, nil
		}

		// Do we know the width, even though the value is unknown?
		if i.Exprs[0].Width() == 1 {
			i.WidthKnown = true
			i.Width = 2
			i.Op, i.Mode = i.ZeroOp, i.ZeroMode
			i.ZeroOp, i.ZeroMode = 0, 0
			return false, nil
		}

		// Okay, we have to set the width: since we don't know, go wide.
		i.WidthKnown = true
		i.Width = 3
		i.ZeroOp, i.ZeroMode = 0, 0
		return false, nil
	}

	// If we got here, we got an actual value.

	// We need to figure out the width
	if !i.WidthKnown {
		if val < 0x100 {
			i.Width = 2
			i.Op = i.ZeroOp
			i.Mode = i.ZeroMode
		} else {
			i.Width = 3
		}
	}

	// It's a branch
	if i.Mode == opcodes.MODE_RELATIVE {
		curr, ok := c.GetAddr()
		if !ok {
			if final {
				return false, i.Errorf("cannot determine current address for '%s'", i.Command)
			}
			return false, nil
		}
		// Found both current and target addresses
		offset := int32(val) - (int32(curr) + 2)
		if offset > 127 {
			return false, i.Errorf("%s cannot jump forward %d (max 127) from $%04x to $%04x", i.Command, offset, curr+2, val)
		}
		if offset < -128 {
			return false, i.Errorf("%s cannot jump back %d (max -128) from $%04x to $%04x", i.Command, offset, curr+2, val)
		}
		val = uint16(offset)
	}

	i.WidthKnown = true
	i.Final = true
	i.ZeroOp = 0
	i.ZeroMode = 0

	switch i.Width {
	case 2:
		// TODO(zellyn): Warn if > 0xff
		i.Data = []byte{i.Op, byte(val)}
	case 3:
		i.Data = []byte{i.Op, byte(val), byte(val >> 8)}
	default:
		panic(fmt.Sprintf("computeOp reached erroneously for '%s'", i.Command))
	}
	return true, nil
}

func (i I) Errorf(format string, a ...interface{}) error {
	return i.Line.Errorf(format, a...)
}

func (i I) Sprintf(format string, a ...interface{}) string {
	return i.Line.Sprintf(format, a...)
}
