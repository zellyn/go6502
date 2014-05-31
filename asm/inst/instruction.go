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
)

// Variants for "TypeData" instructions.
const (
	DataBytes       = iota // Data: expressions, but forced to one byte per
	DataMixed              // Bytes or words (LE), depending on individual expression widths
	DataWordsLe            // Data: expressions, but forced to one word per, little-endian
	DataWordsBe            // Data: expressions, but forced to one word per, big-endian
	DataAscii              // Data: from ASCII strings, high bit clear
	DataAsciiHi            // Data: from ASCII strings, high bit set
	DataAsciiFlip          // Data: from ASCII strings, high bit clear, except last char
	DataAsciiHiFlip        // Data: from ASCII strings, high bit set, except last char
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
	MinWidth     uint16                 // minimum width in bytes
	MaxWidth     uint16                 // maximum width in bytes
	Final        bool                   // Do we know the actual bytes yet?
	Op           byte                   // Opcode
	Mode         opcodes.AddressingMode // Opcode mode
	ZeroMode     opcodes.AddressingMode // Possible ZP-option mode
	ZeroOp       byte                   // Possible ZP-option Opcode
	Value        uint16                 // For Equates, the value
	DeclaredLine uint16                 // Line number listed in file
	Line         *lines.Line            // Line object for this line
	Addr         uint16                 // Current memory address
	AddrKnown    bool                   // Whether the current memory address is known
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
		case DataMixed:
			return "data"
		case DataBytes:
			return "data/b"
		case DataWordsLe:
			return "data/wle"
		case DataWordsBe:
			return "data/wbe"
		case DataAscii, DataAsciiHi, DataAsciiFlip, DataAsciiHiFlip:
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
	}
	s := "{" + i.TypeString()
	if i.Label != "" {
		s += fmt.Sprintf(" '%s'", i.Label)
	}
	if i.TextArg != "" {
		s += fmt.Sprintf(" '%s'", i.TextArg)
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
func (i *I) Compute(c context.Context, setWidth bool, final bool) (bool, error) {
	if i.Type == TypeEqu || i.Type == TypeTarget || i.Type == TypeOrg {
		return i.computeMustKnow(c, setWidth, final)
	}
	if err := i.computeLabel(c, setWidth, final); err != nil {
		return false, err
	}
	if i.Final {
		return true, nil
	}
	switch i.Type {
	case TypeOp:
		return i.computeOp(c, setWidth, final)
	case TypeData:
		return i.computeData(c, setWidth, final)
	case TypeBlock:
		return i.computeBlock(c, setWidth, final)
	}

	// Everything else is zero-width
	i.WidthKnown = true
	i.MinWidth = 0
	i.MaxWidth = 0
	i.Final = true

	return true, nil
}

// FixLabels attempts to turn .1 into LAST_LABEL.1, etc.
func (i *I) FixLabels(labeler context.Labeler) error {
	macroCall := i.Line.GetMacroCall()
	parent := labeler.IsNewParentLabel(i.Label)
	newL, err := labeler.FixLabel(i.Label, macroCall)
	if err != nil {
		return i.Errorf("%v", err)
	}
	i.Label = newL
	if parent {
		labeler.SetLastLabel(i.Label)
	}

	for _, e := range i.Exprs {
		if err := e.FixLabels(labeler, macroCall, i.Line); err != nil {
			return err
		}
	}

	return nil
}

// computeLabel attempts to compute equates and label values.
func (i *I) computeLabel(c context.Context, setWidth bool, final bool) error {
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

func (i *I) computeData(c context.Context, setWidth bool, final bool) (bool, error) {
	if len(i.Data) > 0 {
		i.WidthKnown = true
		i.MinWidth = uint16(len(i.Data))
		i.MaxWidth = i.MinWidth
		i.Final = true
		return true, nil
	}

	allFinal := true
	data := []byte{}
	var width uint16
	for _, e := range i.Exprs {
		var w uint16
		switch i.Var {
		case DataMixed:
			w = e.Width()
		case DataBytes:
			w = 1
		case DataWordsLe, DataWordsBe:
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
		case DataMixed:
			switch w {
			case 1:
				data = append(data, byte(val))
			case 2:
				data = append(data, byte(val), byte(val>>8))
			}
		case DataBytes:
			data = append(data, byte(val))
		case DataWordsLe:
			data = append(data, byte(val), byte(val>>8))
		case DataWordsBe:
			data = append(data, byte(val>>8), byte(val))
		default:
			panic(fmt.Sprintf("Unknown data variant handed to computeData: %d", i.Var))
		}
	}
	i.MinWidth = width
	i.MaxWidth = width
	i.WidthKnown = true
	if allFinal {
		i.Data = data
		i.Final = true
	}

	return i.Final, nil
}

func (i *I) computeBlock(c context.Context, setWidth bool, final bool) (bool, error) {
	val, err := i.Exprs[0].Eval(c, i.Line)
	if err == nil {
		i.Value = val
		i.WidthKnown = true
		i.Final = true
		i.MinWidth = val
		i.MaxWidth = val
	} else {
		if setWidth || final {
			return false, i.Errorf("block storage with unknown size")
		}
	}
	return i.Final, nil
}

func (i *I) computeMustKnow(c context.Context, setWidth bool, final bool) (bool, error) {
	i.WidthKnown = true
	i.MinWidth = 0
	i.MaxWidth = 0
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
	if err := i.computeLabel(c, setWidth, final); err != nil {
		return false, err
	}
	return true, nil
}

func (i *I) computeOp(c context.Context, setWidth bool, final bool) (bool, error) {
	// If the width is not known, we better have a ZeroPage alternative.
	if !i.WidthKnown && (i.ZeroOp == 0 || i.ZeroMode == 0) {
		panic(fmt.Sprintf("Reached computeOp for '%s' with no ZeroPage alternative, i.Command"))
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
			i.MinWidth, i.MaxWidth = 2, 2
			i.Op, i.Mode = i.ZeroOp, i.ZeroMode
			i.ZeroOp, i.ZeroMode = 0, 0
			return false, nil
		}

		// Don't know the width, but don't care on this pass.
		if !setWidth {
			return false, nil
		}

		// Okay, we have to set the width: since we don't know, go wide.
		i.WidthKnown = true
		i.MinWidth, i.MaxWidth = 3, 3
		i.ZeroOp, i.ZeroMode = 0, 0
		return false, nil
	}

	// If we got here, we got an actual value.

	// We need to figure out the width
	if !i.WidthKnown {
		if val < 0x100 {
			i.MinWidth = 2
			i.MaxWidth = 2
			i.Op = i.ZeroOp
			i.Mode = i.ZeroMode
		} else {
			i.MinWidth = 3
			i.MaxWidth = 3

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

	switch i.MinWidth {
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
