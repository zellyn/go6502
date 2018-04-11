/*
Package opcodes stores metadata/information about 6502 opcodes. It's
used for disassembly, and (in the future) for building the opcode
function tables.
*/
package opcodes

// Opcode addressing modes.
type AddressingMode uint

const MODE_UNKNOWN = 0

const (
	MODE_IMPLIED AddressingMode = 1 << iota
	MODE_ABSOLUTE
	MODE_INDIRECT
	MODE_RELATIVE
	MODE_IMMEDIATE
	MODE_ABS_X
	MODE_ABS_Y
	MODE_ZP
	MODE_ZP_X
	MODE_ZP_Y
	MODE_INDIRECT_Y
	MODE_INDIRECT_X
	MODE_A
)

// Logical OR of the three indirect modes
const MODE_INDIRECT_ANY AddressingMode = MODE_INDIRECT | MODE_INDIRECT_X | MODE_INDIRECT_Y

// Opcode read/write semantics: does the opcode read, write, or
// rmw. Useful to distinguish between instructions further than just
// addressing mode.
type ReadWrite int

const (
	RW_X   ReadWrite = 0 // Don't care
	RW_R   ReadWrite = 1
	RW_W   ReadWrite = 2
	RW_RMW ReadWrite = 3
)

// Lengths of instructions for each addressing mode.
var ModeLengths = map[AddressingMode]int{
	MODE_IMPLIED:    1,
	MODE_ABSOLUTE:   3,
	MODE_INDIRECT:   3,
	MODE_RELATIVE:   2,
	MODE_IMMEDIATE:  2,
	MODE_ABS_X:      3,
	MODE_ABS_Y:      3,
	MODE_ZP:         2,
	MODE_ZP_X:       2,
	MODE_ZP_Y:       2,
	MODE_INDIRECT_Y: 2,
	MODE_INDIRECT_X: 2,
	MODE_A:          1,
}

// Opcode stores information about instructions.
type Opcode struct {
	Name string
	Mode AddressingMode
	RW   ReadWrite
}

// Fake NoOp instruction used when disassembling.
var NoOp = Opcode{"???", MODE_IMPLIED, RW_X}

// The list of Opcodes.
var Opcodes = map[byte]Opcode{
	// BUG(zellyn): Add 65C02 instructions.

	// Flag set and clear
	0x18: {"CLC", MODE_IMPLIED, RW_X}, // CLC
	0xD8: {"CLD", MODE_IMPLIED, RW_X}, // CLD
	0x58: {"CLI", MODE_IMPLIED, RW_X}, // CLI
	0xB8: {"CLV", MODE_IMPLIED, RW_X}, // CLV
	0x38: {"SEC", MODE_IMPLIED, RW_X}, // SEC
	0xF8: {"SED", MODE_IMPLIED, RW_X}, // SED
	0x78: {"SEI", MODE_IMPLIED, RW_X}, // SEI

	// Very simple 1-opcode instructions
	0xEA: {"NOP", MODE_IMPLIED, RW_X},
	0xAA: {"TAX", MODE_IMPLIED, RW_X},
	0xA8: {"TAY", MODE_IMPLIED, RW_X},
	0xBA: {"TSX", MODE_IMPLIED, RW_X},
	0x8A: {"TXA", MODE_IMPLIED, RW_X},
	0x9A: {"TXS", MODE_IMPLIED, RW_X},
	0x98: {"TYA", MODE_IMPLIED, RW_X},

	// Slightly more complex 1-opcode instructions
	0xCA: {"DEX", MODE_IMPLIED, RW_X},
	0x88: {"DEY", MODE_IMPLIED, RW_X},
	0xE8: {"INX", MODE_IMPLIED, RW_X},
	0xC8: {"INY", MODE_IMPLIED, RW_X},
	0x48: {"PHA", MODE_IMPLIED, RW_X},
	0x08: {"PHP", MODE_IMPLIED, RW_X},
	0x68: {"PLA", MODE_IMPLIED, RW_X},
	0x28: {"PLP", MODE_IMPLIED, RW_X},

	// Jumps, returns, etc.
	0x4C: {"JMP", MODE_ABSOLUTE, RW_X},
	0x6C: {"JMP", MODE_INDIRECT, RW_X},
	0x20: {"JSR", MODE_ABSOLUTE, RW_X},
	0x60: {"RTS", MODE_IMPLIED, RW_X},
	0x40: {"RTI", MODE_IMPLIED, RW_X},
	0x00: {"BRK", MODE_IMPLIED, RW_X},

	// Branches
	0x90: {"BCC", MODE_RELATIVE, RW_X}, // BCC
	0xB0: {"BCS", MODE_RELATIVE, RW_X}, // BCS
	0xF0: {"BEQ", MODE_RELATIVE, RW_X}, // BEQ
	0x30: {"BMI", MODE_RELATIVE, RW_X}, // BMI
	0xD0: {"BNE", MODE_RELATIVE, RW_X}, // BNE
	0x10: {"BPL", MODE_RELATIVE, RW_X}, // BPL
	0x50: {"BVC", MODE_RELATIVE, RW_X}, // BVC
	0x70: {"BVS", MODE_RELATIVE, RW_X}, // BVS

	// 2-opcode, 2-cycle immediate mode
	0x09: {"ORA", MODE_IMMEDIATE, RW_R},
	0x29: {"AND", MODE_IMMEDIATE, RW_R},
	0x49: {"EOR", MODE_IMMEDIATE, RW_R},
	0x69: {"ADC", MODE_IMMEDIATE, RW_R},
	0xC0: {"CPY", MODE_IMMEDIATE, RW_R},
	0xC9: {"CMP", MODE_IMMEDIATE, RW_R},
	0xA0: {"LDY", MODE_IMMEDIATE, RW_R},
	0xA2: {"LDX", MODE_IMMEDIATE, RW_R},
	0xA9: {"LDA", MODE_IMMEDIATE, RW_R},
	0xE0: {"CPX", MODE_IMMEDIATE, RW_R},
	0xE9: {"SBC", MODE_IMMEDIATE, RW_R},

	// 3-opcode, 4-cycle absolute mode
	0x8D: {"STA", MODE_ABSOLUTE, RW_W},
	0x8E: {"STX", MODE_ABSOLUTE, RW_W},
	0x8C: {"STY", MODE_ABSOLUTE, RW_W},
	0x6D: {"ADC", MODE_ABSOLUTE, RW_R},
	0x2D: {"AND", MODE_ABSOLUTE, RW_R},
	0x2C: {"BIT", MODE_ABSOLUTE, RW_R},
	0xCD: {"CMP", MODE_ABSOLUTE, RW_R},
	0xEC: {"CPX", MODE_ABSOLUTE, RW_R},
	0xCC: {"CPY", MODE_ABSOLUTE, RW_R},
	0x4D: {"EOR", MODE_ABSOLUTE, RW_R},
	0xAD: {"LDA", MODE_ABSOLUTE, RW_R},
	0xAE: {"LDX", MODE_ABSOLUTE, RW_R},
	0xAC: {"LDY", MODE_ABSOLUTE, RW_R},
	0x0D: {"ORA", MODE_ABSOLUTE, RW_R},
	0xED: {"SBC", MODE_ABSOLUTE, RW_R},

	// 2-opcode, 3-cycle zero page
	0x05: {"ORA", MODE_ZP, RW_R},
	0x24: {"BIT", MODE_ZP, RW_R},
	0x25: {"AND", MODE_ZP, RW_R},
	0x45: {"EOR", MODE_ZP, RW_R},
	0x65: {"ADC", MODE_ZP, RW_R},
	0x84: {"STY", MODE_ZP, RW_W},
	0x85: {"STA", MODE_ZP, RW_W},
	0x86: {"STX", MODE_ZP, RW_W},
	0xA4: {"LDY", MODE_ZP, RW_R},
	0xA5: {"LDA", MODE_ZP, RW_R},
	0xA6: {"LDX", MODE_ZP, RW_R},
	0xC4: {"CPY", MODE_ZP, RW_R},
	0xC5: {"CMP", MODE_ZP, RW_R},
	0xE4: {"CPX", MODE_ZP, RW_R},
	0xE5: {"SBC", MODE_ZP, RW_R},

	// 3-opcode, 4*-cycle abs,X/Y
	0x1D: {"ORA", MODE_ABS_X, RW_R},
	0x19: {"ORA", MODE_ABS_Y, RW_R},
	0x39: {"AND", MODE_ABS_Y, RW_R},
	0x3D: {"AND", MODE_ABS_X, RW_R},
	0x59: {"EOR", MODE_ABS_Y, RW_R},
	0x5D: {"EOR", MODE_ABS_X, RW_R},
	0x79: {"ADC", MODE_ABS_Y, RW_R},
	0x7D: {"ADC", MODE_ABS_X, RW_R},
	0xB9: {"LDA", MODE_ABS_Y, RW_R},
	0xBD: {"LDA", MODE_ABS_X, RW_R},
	0xD9: {"CMP", MODE_ABS_Y, RW_R},
	0xDD: {"CMP", MODE_ABS_X, RW_R},
	0xF9: {"SBC", MODE_ABS_Y, RW_R},
	0xFD: {"SBC", MODE_ABS_X, RW_R},
	0xBE: {"LDX", MODE_ABS_Y, RW_R},
	0xBC: {"LDY", MODE_ABS_X, RW_R},

	// 3-opcode, 5-cycle abs,X/Y
	0x99: {"STA", MODE_ABS_Y, RW_W},
	0x9D: {"STA", MODE_ABS_X, RW_W},

	// 2-opcode, 4-cycle zp,X/Y
	0x15: {"ORA", MODE_ZP_X, RW_R},
	0x35: {"AND", MODE_ZP_X, RW_R},
	0x55: {"EOR", MODE_ZP_X, RW_R},
	0x75: {"ADC", MODE_ZP_X, RW_R},
	0x95: {"STA", MODE_ZP_X, RW_W},
	0xB5: {"LDA", MODE_ZP_X, RW_R},
	0xD5: {"CMP", MODE_ZP_X, RW_R},
	0xF5: {"SBC", MODE_ZP_X, RW_R},
	0x96: {"STX", MODE_ZP_Y, RW_W},
	0xB6: {"LDX", MODE_ZP_Y, RW_R},
	0x94: {"STY", MODE_ZP_X, RW_W},
	0xB4: {"LDY", MODE_ZP_X, RW_R},

	// 2-opcode, 5*-cycle zero-page indirect Y
	0x11: {"ORA", MODE_INDIRECT_Y, RW_R},
	0x31: {"AND", MODE_INDIRECT_Y, RW_R},
	0x51: {"EOR", MODE_INDIRECT_Y, RW_R},
	0x71: {"ADC", MODE_INDIRECT_Y, RW_R},
	0x91: {"STA", MODE_INDIRECT_Y, RW_W},
	0xB1: {"LDA", MODE_INDIRECT_Y, RW_R},
	0xD1: {"CMP", MODE_INDIRECT_Y, RW_R},
	0xF1: {"SBC", MODE_INDIRECT_Y, RW_R},

	// 2-opcode, 6-cycle zero-page X indirect
	0x01: {"ORA", MODE_INDIRECT_X, RW_R},
	0x21: {"AND", MODE_INDIRECT_X, RW_R},
	0x41: {"EOR", MODE_INDIRECT_X, RW_R},
	0x61: {"ADC", MODE_INDIRECT_X, RW_R},
	0x81: {"STA", MODE_INDIRECT_X, RW_W},
	0xA1: {"LDA", MODE_INDIRECT_X, RW_R},
	0xC1: {"CMP", MODE_INDIRECT_X, RW_R},
	0xE1: {"SBC", MODE_INDIRECT_X, RW_R},

	// 1-opcode, 2-cycle, accumulator rmw
	0x0A: {"ASL", MODE_A, RW_RMW},
	0x2A: {"ROL", MODE_A, RW_RMW},
	0x4A: {"LSR", MODE_A, RW_RMW},
	0x6A: {"ROR", MODE_A, RW_RMW},

	// 2-opcode, 5-cycle, zp rmw
	0x06: {"ASL", MODE_ZP, RW_RMW},
	0x26: {"ROL", MODE_ZP, RW_RMW},
	0x46: {"LSR", MODE_ZP, RW_RMW},
	0x66: {"ROR", MODE_ZP, RW_RMW},
	0xC6: {"DEC", MODE_ZP, RW_RMW},
	0xE6: {"INC", MODE_ZP, RW_RMW},

	// 3-opcode, 6-cycle, abs rmw
	0x0E: {"ASL", MODE_ABSOLUTE, RW_RMW},
	0x2E: {"ROL", MODE_ABSOLUTE, RW_RMW},
	0x4E: {"LSR", MODE_ABSOLUTE, RW_RMW},
	0x6E: {"ROR", MODE_ABSOLUTE, RW_RMW},
	0xCE: {"DEC", MODE_ABSOLUTE, RW_RMW},
	0xEE: {"INC", MODE_ABSOLUTE, RW_RMW},

	// 2-opcode, 6-cycle, zp,X rmw
	0x16: {"ASL", MODE_ZP_X, RW_RMW},
	0x36: {"ROL", MODE_ZP_X, RW_RMW},
	0x56: {"LSR", MODE_ZP_X, RW_RMW},
	0x76: {"ROR", MODE_ZP_X, RW_RMW},
	0xD6: {"DEC", MODE_ZP_X, RW_RMW},
	0xF6: {"INC", MODE_ZP_X, RW_RMW},

	// 3-opcode, 7-cycle, abs,X rmw
	0x1E: {"ASL", MODE_ABS_X, RW_RMW},
	0x3E: {"ROL", MODE_ABS_X, RW_RMW},
	0x5E: {"LSR", MODE_ABS_X, RW_RMW},
	0x7E: {"ROR", MODE_ABS_X, RW_RMW},
	0xDE: {"DEC", MODE_ABS_X, RW_RMW},
	0xFE: {"INC", MODE_ABS_X, RW_RMW},
}

// Information for lookup of opcodes by name and mode -------------------------

type OpInfo struct {
	Mode   AddressingMode
	Length int
	Byte   byte
}

type OpSummary struct {
	Modes AddressingMode // bitmask of supported modes
	Ops   []OpInfo
}

type Set uint16

const (
	SetUnknown Set = iota
	SetSweet16
)

func ByName(sets Set) map[string]OpSummary {
	m := make(map[string]OpSummary)
	for b, oc := range Opcodes {
		info := OpInfo{oc.Mode, ModeLengths[oc.Mode], b}
		summary := m[oc.Name]
		summary.Modes |= oc.Mode
		summary.Ops = append(summary.Ops, info)
		m[oc.Name] = summary
	}

	return m
}

func (s OpSummary) AnyModes(modes AddressingMode) bool {
	return modes&s.Modes != 0
}

func (s OpSummary) OpForMode(mode AddressingMode) (OpInfo, bool) {
	for _, o := range s.Ops {
		if o.Mode == mode {
			return o, true
		}
	}
	return OpInfo{}, false
}
