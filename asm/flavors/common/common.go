package common

import (
	"fmt"

	"github.com/zellyn/go6502/asm/context"
	"github.com/zellyn/go6502/asm/inst"
	"github.com/zellyn/go6502/opcodes"
)

// DecodeOp contains the common code that decodes an Opcode, once we
// have fully parsed it.
func DecodeOp(c context.Context, in inst.I, summary opcodes.OpSummary, indirect bool, xy rune) (inst.I, error) {
	i := inst.I{}
	ex := in.Exprs[0]

	// At this point we've parsed the Opcode. Let's see if it makes sense.
	if indirect {
		switch xy {
		case 'x':
			op, ok := summary.OpForMode(opcodes.MODE_INDIRECT_X)
			if !ok {
				return i, fmt.Errorf("%s doesn't support indexed indirect (addr,X) mode", in.Command)
			}
			in.Op = op.Byte
			in.WidthKnown = true
			in.MinWidth = 2
			in.MaxWidth = 2
			in.Mode = opcodes.MODE_INDIRECT_X
			return in, nil
		case 'y':
			op, ok := summary.OpForMode(opcodes.MODE_INDIRECT_Y)
			if !ok {
				return i, fmt.Errorf("%s doesn't support indirect indexed (addr),Y mode", in.Command)
			}
			in.WidthKnown = true
			in.MinWidth = 2
			in.MaxWidth = 2
			in.Mode = opcodes.MODE_INDIRECT_Y
			in.Op = op.Byte
			return in, nil
		default:
			op, ok := summary.OpForMode(opcodes.MODE_INDIRECT)
			if !ok {
				return i, fmt.Errorf("%s doesn't support indirect (addr) mode", in.Command)
			}
			in.Op = op.Byte
			in.WidthKnown = true
			in.MinWidth = 3
			in.MaxWidth = 3
			in.Mode = opcodes.MODE_INDIRECT
			return in, nil
		}
	}

	// Branch
	if summary.Modes == opcodes.MODE_RELATIVE {
		op, ok := summary.OpForMode(opcodes.MODE_RELATIVE)
		if !ok {
			panic(fmt.Sprintf("opcode error: %s has no MODE_RELATIVE opcode", in.Command))
		}
		in.Op = op.Byte
		in.WidthKnown = true
		in.MinWidth = 2
		in.MaxWidth = 2
		in.Mode = opcodes.MODE_RELATIVE
		return in, nil
	}

	// No ,X or ,Y, and width is forced to 1-byte: immediate mode.
	if xy == '-' && ex.Width() == 1 && summary.AnyModes(opcodes.MODE_IMMEDIATE) {
		op, ok := summary.OpForMode(opcodes.MODE_IMMEDIATE)
		if !ok {
			panic(fmt.Sprintf("opcode error: %s has no MODE_IMMEDIATE opcode", in.Command))
		}
		in.Op = op.Byte
		in.WidthKnown = true
		in.MinWidth = 2
		in.MaxWidth = 2
		in.Mode = opcodes.MODE_IMMEDIATE
		return in, nil
	}

	var zp, wide opcodes.AddressingMode
	var zpS, wideS string
	switch xy {
	case 'x':
		zp, wide = opcodes.MODE_ZP_X, opcodes.MODE_ABS_X
		zpS, wideS = "ZeroPage,X", "Absolute,X"
	case 'y':
		zp, wide = opcodes.MODE_ZP_Y, opcodes.MODE_ABS_Y
		zpS, wideS = "ZeroPage,Y", "Absolute,Y"
	default:
		zp, wide = opcodes.MODE_ZP, opcodes.MODE_ABSOLUTE
		zpS, wideS = "ZeroPage", "Absolute"
	}

	opWide, wideOk := summary.OpForMode(wide)
	opZp, zpOk := summary.OpForMode(zp)

	if !summary.AnyModes(zp | wide) {
		return i, fmt.Errorf("%s opcode doesn't support %s or %s modes.", zpS, wideS)
	}

	if !summary.AnyModes(zp) {
		if !wideOk {
			panic(fmt.Sprintf("opcode error: %s has no %s opcode", in.Command, wideS))
		}
		in.Op = opWide.Byte
		in.WidthKnown = true
		in.MinWidth = 3
		in.MaxWidth = 3
		in.Mode = wide
		return in, nil
	}
	if !summary.AnyModes(wide) {
		if !zpOk {
			panic(fmt.Sprintf("opcode error: %s has no %s opcode", in.Command, zpS))
		}
		in.Op = opZp.Byte
		in.WidthKnown = true
		in.MinWidth = 2
		in.MaxWidth = 2
		in.Mode = zp
		return in, nil
	}

	// Okay, we don't know whether it's wide or narrow: store enough info for either.
	if !zpOk {
		panic(fmt.Sprintf("opcode error: %s has no %s opcode", in.Command, zpS))
	}
	if !wideOk {
		panic(fmt.Sprintf("opcode error: %s has no %s opcode", in.Command, wideS))
	}
	in.Op = opWide.Byte
	in.ZeroOp = opZp.Byte
	in.Mode = wide
	in.ZeroMode = zp
	in.MinWidth = 2
	in.MaxWidth = 3
	return in, nil
}
