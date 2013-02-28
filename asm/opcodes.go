package asm

// Opcode addressing modes.
const (
	MODE_IMPLIED = iota
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

// Lengths of instructions for each addressing mode.
var ModeLengths = map[int]int{
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
	Mode int
}

// Fake NoOp instruction used when disassembling.
var NoOp = Opcode{"???", MODE_IMPLIED}

// The list of Opcodes.
var Opcodes = map[byte]Opcode{
	// BUG(zellyn): Add 65C02 instructions.

	// Flag set and clear
	0x18: {"CLC", MODE_IMPLIED}, // CLC
	0xD8: {"CLD", MODE_IMPLIED}, // CLD
	0x58: {"CLI", MODE_IMPLIED}, // CLI
	0xB8: {"CLV", MODE_IMPLIED}, // CLV
	0x38: {"SEC", MODE_IMPLIED},   // SEC
	0xF8: {"SED", MODE_IMPLIED},   // SED
	0x78: {"SEI", MODE_IMPLIED},   // SEI

	// Very simple 1-opcode instructions
	0xEA: {"NOP", MODE_IMPLIED},
	0xAA: {"TAX", MODE_IMPLIED},
	0xA8: {"TAY", MODE_IMPLIED},
	0xBA: {"TSX", MODE_IMPLIED},
	0x8A: {"TXA", MODE_IMPLIED},
	0x9A: {"TXS", MODE_IMPLIED},
	0x98: {"TYA", MODE_IMPLIED},

	// Slightly more complex 1-opcode instructions
	0xCA: {"DEX", MODE_IMPLIED},
	0x88: {"DEY", MODE_IMPLIED},
	0xE8: {"INX", MODE_IMPLIED},
	0xC8: {"INY", MODE_IMPLIED},
	0x48: {"PHA", MODE_IMPLIED},
	0x08: {"PHP", MODE_IMPLIED},
	0x68: {"PLA", MODE_IMPLIED},
	0x28: {"PLP", MODE_IMPLIED},

	// Jumps, returns, etc.
	0x4C: {"JMP", MODE_ABSOLUTE},
	0x6C: {"JMP", MODE_INDIRECT},
	0x20: {"JSR", MODE_ABSOLUTE},
	0x60: {"RTS", MODE_IMPLIED},
	0x40: {"RTI", MODE_IMPLIED},
	0x00: {"BRK", MODE_IMPLIED},

	// Branches
	0x90: {"BCC", MODE_RELATIVE}, // BCC
	0xB0: {"BCS", MODE_RELATIVE}, // BCS
	0xF0: {"BEQ", MODE_RELATIVE}, // BEQ
	0x30: {"BMI", MODE_RELATIVE}, // BMI
	0xD0: {"BNE", MODE_RELATIVE}, // BNE
	0x10: {"BPL", MODE_RELATIVE}, // BPL
	0x50: {"BVC", MODE_RELATIVE}, // BVC
	0x70: {"BVS", MODE_RELATIVE}, // BVS

	// 2-opcode, 2-cycle immediate mode
	0x09: {"ORA", MODE_IMMEDIATE},
	0x29: {"AND", MODE_IMMEDIATE},
	0x49: {"EOR", MODE_IMMEDIATE},
	0x69: {"ADC", MODE_IMMEDIATE},
	0xC0: {"CPY", MODE_IMMEDIATE},
	0xC9: {"CMP", MODE_IMMEDIATE},
	0xA0: {"LDY", MODE_IMMEDIATE},
	0xA2: {"LDX", MODE_IMMEDIATE},
	0xA9: {"LDA", MODE_IMMEDIATE},
	0xE0: {"CPX", MODE_IMMEDIATE},
	0xE9: {"SBC", MODE_IMMEDIATE},

	// 3-opcode, 4-cycle absolute mode
	0x8D: {"STA", MODE_ABSOLUTE},
	0x8E: {"STX", MODE_ABSOLUTE},
	0x8C: {"STY", MODE_ABSOLUTE},
	0x6D: {"ADC", MODE_ABSOLUTE},
	0x2D: {"AND", MODE_ABSOLUTE},
	0x2C: {"BIT", MODE_ABSOLUTE},
	0xCD: {"CMP", MODE_ABSOLUTE},
	0xEC: {"CPX", MODE_ABSOLUTE},
	0xCC: {"CPY", MODE_ABSOLUTE},
	0x4D: {"EOR", MODE_ABSOLUTE},
	0xAD: {"LDA", MODE_ABSOLUTE},
	0xAE: {"LDX", MODE_ABSOLUTE},
	0xAC: {"LDY", MODE_ABSOLUTE},
	0x0D: {"ORA", MODE_ABSOLUTE},
	0xED: {"SBC", MODE_ABSOLUTE},

	// 2-opcode, 3-cycle zero page
	0x05: {"ORA", MODE_ZP},
	0x24: {"BIT", MODE_ZP},
	0x25: {"AND", MODE_ZP},
	0x45: {"EOR", MODE_ZP},
	0x65: {"ADC", MODE_ZP},
	0x84: {"STY", MODE_ZP},
	0x85: {"STA", MODE_ZP},
	0x86: {"STX", MODE_ZP},
	0xA4: {"LDY", MODE_ZP},
	0xA5: {"LDA", MODE_ZP},
	0xA6: {"LDX", MODE_ZP},
	0xC4: {"CPY", MODE_ZP},
	0xC5: {"CMP", MODE_ZP},
	0xE4: {"CPX", MODE_ZP},
	0xE5: {"SBC", MODE_ZP},

	// 3-opcode, 4*-cycle abs,X/Y
	0x1D: {"ORA", MODE_ABS_X},
	0x19: {"ORA", MODE_ABS_X},
	0x39: {"AND", MODE_ABS_X},
	0x3D: {"AND", MODE_ABS_X},
	0x59: {"EOR", MODE_ABS_X},
	0x5D: {"EOR", MODE_ABS_X},
	0x79: {"ADC", MODE_ABS_X},
	0x7D: {"ADC", MODE_ABS_X},
	0xBD: {"LDA", MODE_ABS_X},
	0xB9: {"LDA", MODE_ABS_X},
	0xD9: {"CMP", MODE_ABS_X},
	0xDD: {"CMP", MODE_ABS_X},
	0xF9: {"SBC", MODE_ABS_X},
	0xFD: {"SBC", MODE_ABS_X},
	0xBE: {"LDX", MODE_ABS_X},
	0xBC: {"LDY", MODE_ABS_X},

	// 3-opcode, 5-cycle abs,X/Y
	0x99: {"STA", MODE_ABS_Y},
	0x9D: {"STA", MODE_ABS_X},

	// 2-opcode, 4-cycle zp,X/Y
	0x15: {"ORA", MODE_ZP_X},
	0x35: {"AND", MODE_ZP_X},
	0x55: {"EOR", MODE_ZP_X},
	0x75: {"ADC", MODE_ZP_X},
	0x95: {"STA", MODE_ZP_X},
	0xB5: {"LDA", MODE_ZP_X},
	0xD5: {"CMP", MODE_ZP_X},
	0xF5: {"SBC", MODE_ZP_X},
	0x96: {"STX", MODE_ZP_Y},
	0xB6: {"LDX", MODE_ZP_Y},
	0x94: {"STY", MODE_ZP_X},
	0xB4: {"LDY", MODE_ZP_X},

	// 2-opcode, 5*-cycle zero-page indirect Y
	0x11: {"ORA", MODE_INDIRECT_Y},
	0x31: {"AND", MODE_INDIRECT_Y},
	0x51: {"EOR", MODE_INDIRECT_Y},
	0x71: {"ADC", MODE_INDIRECT_Y},
	0x91: {"STA", MODE_INDIRECT_Y},
	0xB1: {"LDA", MODE_INDIRECT_Y},
	0xD1: {"CMP", MODE_INDIRECT_Y},
	0xF1: {"SBC", MODE_INDIRECT_Y},

	// 2-opcode, 6-cycle zero-page X indirect
	0x01: {"ORA", MODE_INDIRECT_X},
	0x21: {"AND", MODE_INDIRECT_X},
	0x41: {"EOR", MODE_INDIRECT_X},
	0x61: {"ADC", MODE_INDIRECT_X},
	0x81: {"STA", MODE_INDIRECT_X},
	0xA1: {"LDA", MODE_INDIRECT_X},
	0xC1: {"CMP", MODE_INDIRECT_X},
	0xE1: {"SBC", MODE_INDIRECT_X},

	// 1-opcode, 2-cycle, accumulator rmw
	0x0A: {"ASL", MODE_A},
	0x2A: {"ROL", MODE_A},
	0x4A: {"LSR", MODE_A},
	0x6A: {"ROR", MODE_A},

	// 2-opcode, 5-cycle, zp rmw
	0x06: {"ASL", MODE_ZP},
	0x26: {"ROL", MODE_ZP},
	0x46: {"LSR", MODE_ZP},
	0x66: {"ROR", MODE_ZP},
	0xC6: {"DEC", MODE_ZP},
	0xE6: {"INC", MODE_ZP},

	// 3-opcode, 6-cycle, abs rmw
	0x0E: {"ASL", MODE_ABSOLUTE},
	0x2E: {"ROL", MODE_ABSOLUTE},
	0x4E: {"LSR", MODE_ABSOLUTE},
	0x6E: {"ROR", MODE_ABSOLUTE},
	0xCE: {"DEC", MODE_ABSOLUTE},
	0xEE: {"INC", MODE_ABSOLUTE},

	// 2-opcode, 6-cycle, zp,X rmw
	0x16: {"ASL", MODE_ZP_X},
	0x36: {"ROL", MODE_ZP_X},
	0x56: {"LSR", MODE_ZP_X},
	0x76: {"ROR", MODE_ZP_X},
	0xD6: {"DEC", MODE_ZP_X},
	0xF6: {"INC", MODE_ZP_X},

	// 3-opcode, 7-cycle, abs,X rmw
	0x1E: {"ASL", MODE_ABS_X},
	0x3E: {"ROL", MODE_ABS_X},
	0x5E: {"LSR", MODE_ABS_X},
	0x7E: {"ROR", MODE_ABS_X},
	0xDE: {"DEC", MODE_ABS_X},
	0xFE: {"INC", MODE_ABS_X},
}
