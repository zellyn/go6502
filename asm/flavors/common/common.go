package common

import (
	"fmt"

	"github.com/zellyn/go6502/asm/context"
	"github.com/zellyn/go6502/asm/inst"
	"github.com/zellyn/go6502/opcodes"
)

const xyzzy = false

// DecodeOp contains the common code that decodes an Opcode, once we
// have fully parsed it.
func DecodeOp(c context.Context, in inst.I, summary opcodes.OpSummary, indirect bool, xy rune, forceWide bool) (inst.I, error) {
	ex := in.Exprs[0]
	val, err := ex.Eval(c, in.Line)
	valKnown := err == nil

	// At this point we've parsed the Opcode. Let's see if it makes sense.
	if indirect {
		switch xy {
		case 'x':
			op, ok := summary.OpForMode(opcodes.MODE_INDIRECT_X)
			if !ok {
				return in, in.Errorf("%s doesn't support indexed indirect (addr,X) mode", in.Command)
			}
			if forceWide {
				return in, in.Errorf("%s (addr,X) doesn't have a wide variant", in.Command)
			}
			in.Op = op.Byte
			in.Width = 2
			in.Var = inst.VarOpByte
			if valKnown {
				in.Final = true
				in.Data = []byte{in.Op, byte(val)}
			}
			return in, nil
		case 'y':
			op, ok := summary.OpForMode(opcodes.MODE_INDIRECT_Y)
			if !ok {
				return in, in.Errorf("%s doesn't support indirect indexed (addr),Y mode", in.Command)
			}
			if forceWide {
				return in, fmt.Errorf("%s (addr),Y doesn't have a wide variant", in.Command)
			}
			in.Width = 2
			in.Op = op.Byte
			in.Var = inst.VarOpByte
			if valKnown {
				in.Final = true
				in.Data = []byte{in.Op, byte(val)}
			}
			return in, nil
		default:
			op, ok := summary.OpForMode(opcodes.MODE_INDIRECT)
			if !ok {
				return in, in.Errorf("%s doesn't support indirect (addr) mode", in.Command)
			}
			in.Op = op.Byte
			in.Width = 3
			in.Var = inst.VarOpWord
			if valKnown {
				in.Final = true
				in.Data = []byte{in.Op, byte(val), byte(val >> 8)}
			}
			return in, nil
		}
	}

	// Branch
	if summary.Modes == opcodes.MODE_RELATIVE {
		op, ok := summary.OpForMode(opcodes.MODE_RELATIVE)
		if !ok {
			panic(fmt.Sprintf("opcode error: %s has no MODE_RELATIVE opcode", in.Command))
		}
		if forceWide {
			return in, fmt.Errorf("%s doesn't have a wide variant", in.Command)
		}

		in.Op = op.Byte
		in.Width = 2
		in.Var = inst.VarOpBranch
		if valKnown {
			b, err := RelativeAddr(c, in, val)
			if err != nil {
				return in, err
			}
			in.Data = []byte{in.Op, b}
			in.Final = true
		}
		return in, nil
	}

	// No ,X or ,Y, and width is forced to 1-byte: immediate mode.
	if xy == '-' && ex.Width() == 1 && summary.AnyModes(opcodes.MODE_IMMEDIATE) && !forceWide {
		op, ok := summary.OpForMode(opcodes.MODE_IMMEDIATE)
		if !ok {
			panic(fmt.Sprintf("opcode error: %s has no MODE_IMMEDIATE opcode", in.Command))
		}
		in.Op = op.Byte
		in.Width = 2
		in.Var = inst.VarOpByte
		if valKnown {
			in.Data = []byte{in.Op, byte(val)}
			in.Final = true
		}
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
		return in, in.Errorf("%s opcode doesn't support %s or %s modes.", zpS, wideS)
	}

	if !zpOk {
		in.Op = opWide.Byte
		in.Width = 3
		in.Var = inst.VarOpWord
		if valKnown {
			in.Data = []byte{in.Op, byte(val), byte(val >> 8)}
			in.Final = true
		}
		return in, nil
	}
	if !wideOk {
		if forceWide {
			return in, fmt.Errorf("%s doesn't have a wide variant", in.Command)
		}
		in.Op = opZp.Byte
		in.Width = 2
		in.Var = inst.VarOpByte
		if valKnown {
			in.Data = []byte{in.Op, byte(val)}
			in.Final = true
		}
		return in, nil
	}

	if forceWide {
		in.Op = opWide.Byte
		in.Width = 3
		in.Var = inst.VarOpWord
		if valKnown {
			in.Data = []byte{in.Op, byte(val), byte(val >> 8)}
			in.Final = true
		}
		return in, nil
	}

	if valKnown {
		if val < 0x100 {
			in.Op = opZp.Byte
			in.Data = []byte{in.Op, byte(val)}
			in.Width = 2
			in.Var = inst.VarOpByte
			in.Final = true
			return in, nil
		}
		in.Op = opWide.Byte
		in.Data = []byte{in.Op, byte(val), byte(val >> 8)}
		in.Width = 3
		in.Var = inst.VarOpWord
		in.Final = true
		return in, nil
	}

	if in.Exprs[0].Width() == 1 {
		in.Op = opZp.Byte
		in.Width = 2
		in.Var = inst.VarOpByte
		return in, nil
	}

	in.Op = opWide.Byte
	in.Width = 3
	in.Var = inst.VarOpWord
	return in, nil
}

func RelativeAddr(c context.Context, in inst.I, val uint16) (byte, error) {
	curr := c.GetAddr()
	offset := int32(val) - (int32(curr) + 2)
	if offset > 127 {
		return 0, in.Errorf("%s cannot jump forward %d (max 127) from $%04x to $%04x", in.Command, offset, curr+2, val)
	}
	if offset < -128 {
		return 0, in.Errorf("%s cannot jump back %d (max -128) from $%04x to $%04x", in.Command, offset, curr+2, val)
	}
	return byte(offset), nil
}
