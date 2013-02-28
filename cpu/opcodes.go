package cpu

// The list of Opcodes.
var Opcodes = map[byte]func(*cpu){
	// BUG(zellyn): Add 65C02 instructions.

	// Flag set and clear
	0x18: clearFlag(FLAG_C), // CLC
	0xD8: clearFlag(FLAG_D), // CLD
	0x58: clearFlag(FLAG_I), // CLI
	0xB8: clearFlag(FLAG_V), // CLV
	0x38: setFlag(FLAG_C),   // SEC
	0xF8: setFlag(FLAG_D),   // SED
	0x78: setFlag(FLAG_I),   // SEI

	// Very simple 1-opcode instructions
	0xEA: nop,
	0xAA: tax,
	0xA8: tay,
	0xBA: tsx,
	0x8A: txa,
	0x9A: txs,
	0x98: tya,

	// Slightly more complex 1-opcode instructions
	0xCA: dex,
	0x88: dey,
	0xE8: inx,
	0xC8: iny,
	0x48: pha,
	0x08: php,
	0x68: pla,
	0x28: plp,

	// Jumps, returns, etc.
	0x4C: jmpAbsolute,
	0x6C: jmpIndirect,
	0x20: jsr,
	0x60: rts,
	0x40: rti,
	0x00: brk,

	// Branches
	0x90: branch(FLAG_C, 0),      // BCC
	0xB0: branch(FLAG_C, FLAG_C), // BCS
	0xF0: branch(FLAG_Z, FLAG_Z), // BEQ
	0x30: branch(FLAG_N, FLAG_N), // BMI
	0xD0: branch(FLAG_Z, 0),      // BNE
	0x10: branch(FLAG_N, 0),      // BPL
	0x50: branch(FLAG_V, 0),      // BVC
	0x70: branch(FLAG_V, FLAG_V), // BVS

	// 2-opcode, 2-cycle immediate mode
	0x09: immediate2(ora),
	0x29: immediate2(and),
	0x49: immediate2(eor),
	0x69: immediate2(adc),
	0xC0: immediate2(cpy),
	0xC9: immediate2(cmp),
	0xA0: immediate2(ldy),
	0xA2: immediate2(ldx),
	0xA9: immediate2(lda),
	0xE0: immediate2(cpx),
	0xE9: immediate2(sbc),

	// 3-opcode, 4-cycle absolute mode
	0x8D: absolute4w(sta),
	0x8E: absolute4w(stx),
	0x8C: absolute4w(sty),
	0x6D: absolute4r(adc),
	0x2D: absolute4r(and),
	0x2C: absolute4r(bit),
	0xCD: absolute4r(cmp),
	0xEC: absolute4r(cpx),
	0xCC: absolute4r(cpy),
	0x4D: absolute4r(eor),
	0xAD: absolute4r(lda),
	0xAE: absolute4r(ldx),
	0xAC: absolute4r(ldy),
	0x0D: absolute4r(ora),
	0xED: absolute4r(sbc),

	// 2-opcode, 3-cycle zero page
	0x05: zp3r(ora),
	0x24: zp3r(bit),
	0x25: zp3r(and),
	0x45: zp3r(eor),
	0x65: zp3r(adc),
	0x84: zp3w(sty),
	0x85: zp3w(sta),
	0x86: zp3w(stx),
	0xA4: zp3r(ldy),
	0xA5: zp3r(lda),
	0xA6: zp3r(ldx),
	0xC4: zp3r(cpy),
	0xC5: zp3r(cmp),
	0xE4: zp3r(cpx),
	0xE5: zp3r(sbc),

	// 3-opcode, 4*-cycle abs,X/Y
	0x1D: absx4r(ora),
	0x19: absy4r(ora),
	0x39: absy4r(and),
	0x3D: absx4r(and),
	0x59: absy4r(eor),
	0x5D: absx4r(eor),
	0x79: absy4r(adc),
	0x7D: absx4r(adc),
	0xBD: absx4r(lda),
	0xB9: absy4r(lda),
	0xD9: absy4r(cmp),
	0xDD: absx4r(cmp),
	0xF9: absy4r(sbc),
	0xFD: absx4r(sbc),
	0xBE: absy4r(ldx),
	0xBC: absx4r(ldy),

	// 3-opcode, 5-cycle abs,X/Y
	0x99: absy5w(sta),
	0x9D: absx5w(sta),

	// 2-opcode, 4-cycle zp,X/Y
	0x15: zpx4r(ora),
	0x35: zpx4r(and),
	0x55: zpx4r(eor),
	0x75: zpx4r(adc),
	0x95: zpx4w(sta),
	0xB5: zpx4r(lda),
	0xD5: zpx4r(cmp),
	0xF5: zpx4r(sbc),
	0x96: zpy4w(stx),
	0xB6: zpy4r(ldx),
	0x94: zpx4w(sty),
	0xB4: zpx4r(ldy),

	// 2-opcode, 5*-cycle zero-page indirect Y
	0x11: zpiy5r(ora),
	0x31: zpiy5r(and),
	0x51: zpiy5r(eor),
	0x71: zpiy5r(adc),
	0x91: zpiy6w(sta),
	0xB1: zpiy5r(lda),
	0xD1: zpiy5r(cmp),
	0xF1: zpiy5r(sbc),

	// 2-opcode, 6-cycle zero-page X indirect
	0x01: zpxi6r(ora),
	0x21: zpxi6r(and),
	0x41: zpxi6r(eor),
	0x61: zpxi6r(adc),
	0x81: zpxi6w(sta),
	0xA1: zpxi6r(lda),
	0xC1: zpxi6r(cmp),
	0xE1: zpxi6r(sbc),

	// 1-opcode, 2-cycle, accumulator rmw
	0x0A: acc2rmw(asl),
	0x2A: acc2rmw(rol),
	0x4A: acc2rmw(lsr),
	0x6A: acc2rmw(ror),

	// 2-opcode, 5-cycle, zp rmw
	0x06: zp5rmw(asl),
	0x26: zp5rmw(rol),
	0x46: zp5rmw(lsr),
	0x66: zp5rmw(ror),
	0xC6: zp5rmw(dec),
	0xE6: zp5rmw(inc),

	// 3-opcode, 6-cycle, abs rmw
	0x0E: abs6rmw(asl),
	0x2E: abs6rmw(rol),
	0x4E: abs6rmw(lsr),
	0x6E: abs6rmw(ror),
	0xCE: abs6rmw(dec),
	0xEE: abs6rmw(inc),

	// 2-opcode, 6-cycle, zp,X rmw
	0x16: zpx6rmw(asl),
	0x36: zpx6rmw(rol),
	0x56: zpx6rmw(lsr),
	0x76: zpx6rmw(ror),
	0xD6: zpx6rmw(dec),
	0xF6: zpx6rmw(inc),

	// 3-opcode, 7-cycle, abs,X rmw
	0x1E: absx7rmw(asl),
	0x3E: absx7rmw(rol),
	0x5E: absx7rmw(lsr),
	0x7E: absx7rmw(ror),
	0xDE: absx7rmw(dec),
	0xFE: absx7rmw(inc),
}
