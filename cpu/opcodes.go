package cpu

import (
	_ "fmt"
)

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
	Name     string
	Mode     int
	function func(*cpu)
}

// Fake NoOp instruction used when disassembling.
var NoOp = Opcode{"???", MODE_IMPLIED, nil}

// The list of Opcodes.
var Opcodes = map[byte]Opcode{
	// BUG(zellyn): Add 65C02 instructions.

	// Flag set and clear
	0x18: {"CLC", MODE_IMPLIED, clearFlag(FLAG_C)}, // CLC
	0xD8: {"CLD", MODE_IMPLIED, clearFlag(FLAG_D)}, // CLD
	0x58: {"CLI", MODE_IMPLIED, clearFlag(FLAG_I)}, // CLI
	0xB8: {"CLV", MODE_IMPLIED, clearFlag(FLAG_V)}, // CLV
	0x38: {"SEC", MODE_IMPLIED, setFlag(FLAG_C)},   // SEC
	0xF8: {"SED", MODE_IMPLIED, setFlag(FLAG_D)},   // SED
	0x78: {"SEI", MODE_IMPLIED, setFlag(FLAG_I)},   // SEI

	// Very simple 1-opcode instructions
	0xEA: {"NOP", MODE_IMPLIED, nop},
	0xAA: {"TAX", MODE_IMPLIED, tax},
	0xA8: {"TAY", MODE_IMPLIED, tay},
	0xBA: {"TSX", MODE_IMPLIED, tsx},
	0x8A: {"TXA", MODE_IMPLIED, txa},
	0x9A: {"TXS", MODE_IMPLIED, txs},
	0x98: {"TYA", MODE_IMPLIED, tya},

	// Slightly more complex 1-opcode instructions
	0xCA: {"DEX", MODE_IMPLIED, dex},
	0x88: {"DEY", MODE_IMPLIED, dey},
	0xE8: {"INX", MODE_IMPLIED, inx},
	0xC8: {"INY", MODE_IMPLIED, iny},
	0x48: {"PHA", MODE_IMPLIED, pha},
	0x08: {"PHP", MODE_IMPLIED, php},
	0x68: {"PLA", MODE_IMPLIED, pla},
	0x28: {"PLP", MODE_IMPLIED, plp},

	// Jumps, returns, etc.
	0x4C: {"JMP", MODE_ABSOLUTE, jmpAbsolute},
	0x6C: {"JMP", MODE_INDIRECT, jmpIndirect},
	0x20: {"JSR", MODE_ABSOLUTE, jsr},
	0x60: {"RTS", MODE_IMPLIED, rts},
	0x40: {"RTI", MODE_IMPLIED, rti},
	0x00: {"BRK", MODE_IMPLIED, brk},

	// Branches
	0x90: {"BCC", MODE_RELATIVE, branch(FLAG_C, 0)},      // BCC
	0xB0: {"BCS", MODE_RELATIVE, branch(FLAG_C, FLAG_C)}, // BCS
	0xF0: {"BEQ", MODE_RELATIVE, branch(FLAG_Z, FLAG_Z)}, // BEQ
	0x30: {"BMI", MODE_RELATIVE, branch(FLAG_N, FLAG_N)}, // BMI
	0xD0: {"BNE", MODE_RELATIVE, branch(FLAG_Z, 0)},      // BNE
	0x10: {"BPL", MODE_RELATIVE, branch(FLAG_N, 0)},      // BPL
	0x50: {"BVC", MODE_RELATIVE, branch(FLAG_V, 0)},      // BVC
	0x70: {"BVS", MODE_RELATIVE, branch(FLAG_V, FLAG_V)}, // BVS

	// 2-opcode, 2-cycle immediate mode
	0x09: {"ORA", MODE_IMMEDIATE, immediate2(ora)},
	0x29: {"AND", MODE_IMMEDIATE, immediate2(and)},
	0x49: {"EOR", MODE_IMMEDIATE, immediate2(eor)},
	0x69: {"ADC", MODE_IMMEDIATE, immediate2(adc)},
	0xC0: {"CPY", MODE_IMMEDIATE, immediate2(cpy)},
	0xC9: {"CMP", MODE_IMMEDIATE, immediate2(cmp)},
	0xA0: {"LDY", MODE_IMMEDIATE, immediate2(ldy)},
	0xA2: {"LDX", MODE_IMMEDIATE, immediate2(ldx)},
	0xA9: {"LDA", MODE_IMMEDIATE, immediate2(lda)},
	0xE0: {"CPX", MODE_IMMEDIATE, immediate2(cpx)},
	0xE9: {"SBC", MODE_IMMEDIATE, immediate2(sbc)},

	// 3-opcode, 4-cycle absolute mode
	0x8D: {"STA", MODE_ABSOLUTE, absolute4w(sta)},
	0x8E: {"STX", MODE_ABSOLUTE, absolute4w(stx)},
	0x8C: {"STY", MODE_ABSOLUTE, absolute4w(sty)},
	0x6D: {"ADC", MODE_ABSOLUTE, absolute4r(adc)},
	0x2D: {"AND", MODE_ABSOLUTE, absolute4r(and)},
	0x2C: {"BIT", MODE_ABSOLUTE, absolute4r(bit)},
	0xCD: {"CMP", MODE_ABSOLUTE, absolute4r(cmp)},
	0xEC: {"CPX", MODE_ABSOLUTE, absolute4r(cpx)},
	0xCC: {"CPY", MODE_ABSOLUTE, absolute4r(cpy)},
	0x4D: {"EOR", MODE_ABSOLUTE, absolute4r(eor)},
	0xAD: {"LDA", MODE_ABSOLUTE, absolute4r(lda)},
	0xAE: {"LDX", MODE_ABSOLUTE, absolute4r(ldx)},
	0xAC: {"LDY", MODE_ABSOLUTE, absolute4r(ldy)},
	0x0D: {"ORA", MODE_ABSOLUTE, absolute4r(ora)},
	0xED: {"SBC", MODE_ABSOLUTE, absolute4r(sbc)},

	// 2-opcode, 3-cycle zero page
	0x05: {"ORA", MODE_ZP, zp3r(ora)},
	0x24: {"BIT", MODE_ZP, zp3r(bit)},
	0x25: {"AND", MODE_ZP, zp3r(and)},
	0x45: {"EOR", MODE_ZP, zp3r(eor)},
	0x65: {"ADC", MODE_ZP, zp3r(adc)},
	0x84: {"STY", MODE_ZP, zp3w(sty)},
	0x85: {"STA", MODE_ZP, zp3w(sta)},
	0x86: {"STX", MODE_ZP, zp3w(stx)},
	0xA4: {"LDY", MODE_ZP, zp3r(ldy)},
	0xA5: {"LDA", MODE_ZP, zp3r(lda)},
	0xA6: {"LDX", MODE_ZP, zp3r(ldx)},
	0xC4: {"CPY", MODE_ZP, zp3r(cpy)},
	0xC5: {"CMP", MODE_ZP, zp3r(cmp)},
	0xE4: {"CPX", MODE_ZP, zp3r(cpx)},
	0xE5: {"SBC", MODE_ZP, zp3r(sbc)},

	// 3-opcode, 4*-cycle abs,X/Y
	0x1D: {"ORA", MODE_ABS_X, absx4r(ora)},
	0x19: {"ORA", MODE_ABS_X, absy4r(ora)},
	0x39: {"AND", MODE_ABS_X, absy4r(and)},
	0x3D: {"AND", MODE_ABS_X, absx4r(and)},
	0x59: {"EOR", MODE_ABS_X, absy4r(eor)},
	0x5D: {"EOR", MODE_ABS_X, absx4r(eor)},
	0x79: {"ADC", MODE_ABS_X, absy4r(adc)},
	0x7D: {"ADC", MODE_ABS_X, absx4r(adc)},
	0xBD: {"LDA", MODE_ABS_X, absx4r(lda)},
	0xB9: {"LDA", MODE_ABS_X, absy4r(lda)},
	0xD9: {"CMP", MODE_ABS_X, absy4r(cmp)},
	0xDD: {"CMP", MODE_ABS_X, absx4r(cmp)},
	0xF9: {"SBC", MODE_ABS_X, absy4r(sbc)},
	0xFD: {"SBC", MODE_ABS_X, absx4r(sbc)},
	0xBE: {"LDX", MODE_ABS_X, absy4r(ldx)},
	0xBC: {"LDY", MODE_ABS_X, absx4r(ldy)},

	// 3-opcode, 5-cycle abs,X/Y
	0x99: {"STA", MODE_ABS_Y, absy5w(sta)},
	0x9D: {"STA", MODE_ABS_X, absx5w(sta)},

	// 2-opcode, 4-cycle zp,X/Y
	0x15: {"ORA", MODE_ZP_X, zpx4r(ora)},
	0x35: {"AND", MODE_ZP_X, zpx4r(and)},
	0x55: {"EOR", MODE_ZP_X, zpx4r(eor)},
	0x75: {"ADC", MODE_ZP_X, zpx4r(adc)},
	0x95: {"STA", MODE_ZP_X, zpx4w(sta)},
	0xB5: {"LDA", MODE_ZP_X, zpx4r(lda)},
	0xD5: {"CMP", MODE_ZP_X, zpx4r(cmp)},
	0xF5: {"SBC", MODE_ZP_X, zpx4r(sbc)},
	0x96: {"STX", MODE_ZP_Y, zpy4w(stx)},
	0xB6: {"LDX", MODE_ZP_Y, zpy4r(ldx)},
	0x94: {"STY", MODE_ZP_X, zpx4w(sty)},
	0xB4: {"LDY", MODE_ZP_X, zpx4r(ldy)},

	// 2-opcode, 5*-cycle zero-page indirect Y
	0x11: {"ORA", MODE_INDIRECT_Y, zpiy5r(ora)},
	0x31: {"AND", MODE_INDIRECT_Y, zpiy5r(and)},
	0x51: {"EOR", MODE_INDIRECT_Y, zpiy5r(eor)},
	0x71: {"ADC", MODE_INDIRECT_Y, zpiy5r(adc)},
	0x91: {"STA", MODE_INDIRECT_Y, zpiy6w(sta)},
	0xB1: {"LDA", MODE_INDIRECT_Y, zpiy5r(lda)},
	0xD1: {"CMP", MODE_INDIRECT_Y, zpiy5r(cmp)},
	0xF1: {"SBC", MODE_INDIRECT_Y, zpiy5r(sbc)},

	// 2-opcode, 6-cycle zero-page X indirect
	0x01: {"ORA", MODE_INDIRECT_X, zpxi6r(ora)},
	0x21: {"AND", MODE_INDIRECT_X, zpxi6r(and)},
	0x41: {"EOR", MODE_INDIRECT_X, zpxi6r(eor)},
	0x61: {"ADC", MODE_INDIRECT_X, zpxi6r(adc)},
	0x81: {"STA", MODE_INDIRECT_X, zpxi6w(sta)},
	0xA1: {"LDA", MODE_INDIRECT_X, zpxi6r(lda)},
	0xC1: {"CMP", MODE_INDIRECT_X, zpxi6r(cmp)},
	0xE1: {"SBC", MODE_INDIRECT_X, zpxi6r(sbc)},

	// 1-opcode, 2-cycle, accumulator rmw
	0x0A: {"ASL", MODE_A, acc2rmw(asl)},
	0x2A: {"ROL", MODE_A, acc2rmw(rol)},
	0x4A: {"LSR", MODE_A, acc2rmw(lsr)},
	0x6A: {"ROR", MODE_A, acc2rmw(ror)},

	// 2-opcode, 5-cycle, zp rmw
	0x06: {"ASL", MODE_ZP, zp5rmw(asl)},
	0x26: {"ROL", MODE_ZP, zp5rmw(rol)},
	0x46: {"LSR", MODE_ZP, zp5rmw(lsr)},
	0x66: {"ROR", MODE_ZP, zp5rmw(ror)},
	0xC6: {"DEC", MODE_ZP, zp5rmw(dec)},
	0xE6: {"INC", MODE_ZP, zp5rmw(inc)},

	// 3-opcode, 6-cycle, abs rmw
	0x0E: {"ASL", MODE_ABSOLUTE, abs6rmw(asl)},
	0x2E: {"ROL", MODE_ABSOLUTE, abs6rmw(rol)},
	0x4E: {"LSR", MODE_ABSOLUTE, abs6rmw(lsr)},
	0x6E: {"ROR", MODE_ABSOLUTE, abs6rmw(ror)},
	0xCE: {"DEC", MODE_ABSOLUTE, abs6rmw(dec)},
	0xEE: {"INC", MODE_ABSOLUTE, abs6rmw(inc)},

	// 2-opcode, 6-cycle, zp,X rmw
	0x16: {"ASL", MODE_ZP_X, zpx6rmw(asl)},
	0x36: {"ROL", MODE_ZP_X, zpx6rmw(rol)},
	0x56: {"LSR", MODE_ZP_X, zpx6rmw(lsr)},
	0x76: {"ROR", MODE_ZP_X, zpx6rmw(ror)},
	0xD6: {"DEC", MODE_ZP_X, zpx6rmw(dec)},
	0xF6: {"INC", MODE_ZP_X, zpx6rmw(inc)},

	// 3-opcode, 7-cycle, abs,X rmw
	0x1E: {"ASL", MODE_ABS_X, absx7rmw(asl)},
	0x3E: {"ROL", MODE_ABS_X, absx7rmw(rol)},
	0x5E: {"LSR", MODE_ABS_X, absx7rmw(lsr)},
	0x7E: {"ROR", MODE_ABS_X, absx7rmw(ror)},
	0xDE: {"DEC", MODE_ABS_X, absx7rmw(dec)},
	0xFE: {"INC", MODE_ABS_X, absx7rmw(inc)},
}
