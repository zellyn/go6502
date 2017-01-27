package inst

import (
	"errors"
	"fmt"
	"strings"

	"github.com/zellyn/go6502/asm/context"
	"github.com/zellyn/go6502/asm/expr"
	"github.com/zellyn/go6502/asm/lines"
)

type Type int

const (
	TypeUnknown Type = Type(iota)

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
	TypeDirective  // An assembler directive that was consumed
)

type Variant int

// Variants for instructions. These tell the instruction how to
// interpret the raw data that comes in on the first or second pass.
const (
	VarUnknown     = Variant(iota)
	VarBytes       // Data: expressions, but forced to one byte per
	VarMixed       // Bytes or words (LE), depending on individual expression widths
	VarWordsLe     // Data: expressions, but forced to one word per, little-endian
	VarWordsBe     // Data: expressions, but forced to one word per, big-endian
	VarBytesZero   // Data: a run of zeros
	VarAscii       // Data: from ASCII strings, high bit clear
	VarAsciiFlip   // Data: from ASCII strings, high bit clear, except last char
	VarAsciiHi     // Data: from ASCII strings, high bit set
	VarAsciiHiFlip // Data: from ASCII strings, high bit set, except last char
	VarRelative    // For branches: a one-byte relative address
	VarEquNormal   // Equ: a normal equate
	VarEquPageZero // Equ: a page-zero equate
	VarOpByte      // An op with a one-byte argument
	VarOpWord      // An op with a one-word argument
	VarOpBranch    // An op with a one-byte relative address argument
)

type I struct {
	Type         Type        // Type of instruction
	Label        string      // Text of label part
	MacroArgs    []string    // Macro args
	Command      string      // Text of command part
	TextArg      string      // Text of argument part
	Exprs        []*expr.E   // Expression(s)
	Data         []byte      // Actual bytes
	Width        uint16      // width in bytes
	Final        bool        // Do we know the actual bytes yet?
	Op           byte        // Opcode
	Value        int64       // For Equates, the value
	DeclaredLine uint16      // Line number listed in file
	Line         *lines.Line // Line object for this line
	Addr         uint16      // Current memory address
	Var          Variant     // Variant of instruction type

	ModeStr string // Mode description, for debug printing
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
		case VarBytesZero:
			return "data/bz"
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
		if i.ModeStr != "" {
			return i.Command + "/" + i.ModeStr
		}
		return i.Command
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
func (i *I) Compute(c context.Context) error {
	if i.Final {
		return nil
	}
	switch i.Type {
	case TypeOp:
		return i.computeOp(c)
	case TypeData:
		return i.computeData(c)
	}

	// Everything else is already final
	// TODO(zellyn): warn if we reach here?
	i.Final = true

	return nil
}

func (i *I) computeData(c context.Context) error {
	if len(i.Data) > 0 {
		i.Width = uint16(len(i.Data))
		i.Final = true
		return nil
	}

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
		val, err := e.Eval(c, i.Line)
		if err != nil {
			return err
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
	i.Data = data
	i.Final = true
	return nil
}

func (i *I) computeBlock(c context.Context, final bool) (bool, error) {
	val, err := i.Exprs[0].Eval(c, i.Line)
	if err == nil {
		i.Value = val
		i.Final = true
		i.Width = uint16(val)
	} else {
		if final {
			return false, i.Errorf("block storage with unknown size")
		}
	}
	return i.Final, nil
}

func (i *I) computeMustKnow(c context.Context) error {
	val, err := i.Exprs[0].Eval(c, i.Line)
	if err != nil {
		return err
	}
	i.Value = val
	switch i.Type {
	case TypeTarget:
		return errors.New("Target not implemented yet.")
	case TypeOrg:
		c.SetAddr(uint16(val))
	case TypeEqu:
		c.Set(i.Label, val)
		// Don't handle labels.
		return nil
	}
	return nil
}

func (i *I) computeOp(c context.Context) error {
	// An op with no args would be final already, so we must have an Expression.
	if len(i.Exprs) == 0 {
		panic(fmt.Sprintf("Reached computeOp for '%s' with no expressions", i.Command))
	}
	val, err := i.Exprs[0].Eval(c, i.Line)
	// A real error
	if err != nil {
		return err
	}

	// If we got here, we got an actual value.

	// It's a branch
	if i.Var == VarOpBranch {
		curr := c.GetAddr()
		offset := int32(val) - (int32(curr) + 2)
		if offset > 127 {
			return i.Errorf("%s cannot jump forward %d (max 127) from $%04x to $%04x", i.Command, offset, curr+2, val)
		}
		if offset < -128 {
			return i.Errorf("%s cannot jump back %d (max -128) from $%04x to $%04x", i.Command, offset, curr+2, val)
		}
		val = int64(offset)
	}

	i.Final = true

	switch i.Width {
	case 2:
		// TODO(zellyn): Warn if > 0xff
		i.Data = []byte{i.Op, byte(val)}
	case 3:
		i.Data = []byte{i.Op, byte(val), byte(val >> 8)}
	default:
		panic(fmt.Sprintf("computeOp reached erroneously for '%s'", i.Command))
	}
	return nil
}

func (i I) Errorf(format string, a ...interface{}) error {
	return i.Line.Errorf(format, a...)
}

func (i I) Sprintf(format string, a ...interface{}) string {
	return i.Line.Sprintf(format, a...)
}
