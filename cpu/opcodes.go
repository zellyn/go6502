package cpu

import (
	_ "fmt"
)

func samePage(a1 uint16, a2 uint16) bool {
	return a1^a2&0xFF00 == 0
}

// Simple, one-off instructions ----------------------------------------------

func clearFlag(flag byte) func(*cpu) {
	return func(c *cpu) {
		c.r.P &^= flag
		c.t.Tick()
	}
}

func setFlag(flag byte) func(*cpu) {
	return func(c *cpu) {
		c.r.P |= flag
		c.t.Tick()
	}
}

func dex(c *cpu) {
	c.r.X--
	c.SetNZ(c.r.X)
	c.t.Tick()
}

func dey(c *cpu) {
	c.r.Y--
	c.SetNZ(c.r.Y)
	c.t.Tick()
}

func inx(c *cpu) {
	c.r.X++
	c.SetNZ(c.r.X)
	c.t.Tick()
}

func iny(c *cpu) {
	c.r.Y++
	c.SetNZ(c.r.Y)
	c.t.Tick()
}

func pha(c *cpu) {
	c.t.Tick()
	c.m.Write(0x100+uint16(c.r.SP), c.r.A)
	c.r.SP--
	c.t.Tick()
}

func php(c *cpu) {
	c.t.Tick()
	c.m.Write(0x100+uint16(c.r.SP), c.r.P)
	c.r.SP--
	c.t.Tick()
}
func pla(c *cpu) {
	c.t.Tick()
	c.r.SP++
	c.t.Tick()
	c.r.A = c.m.Read(0x100 + uint16(c.r.SP))
	c.t.Tick()
}
func plp(c *cpu) {
	c.t.Tick()
	c.r.SP++
	c.t.Tick()
	c.r.P = c.m.Read(0x100 + uint16(c.r.SP))
	c.t.Tick()
}

func nop(c *cpu) {
	c.t.Tick()
}

func tax(c *cpu) {
	c.r.X = c.r.A
	c.SetNZ(c.r.X)
	c.t.Tick()
}

func tay(c *cpu) {
	c.r.Y = c.r.A
	c.SetNZ(c.r.Y)
	c.t.Tick()
}

func tsx(c *cpu) {
	c.r.X = c.r.SP
	c.SetNZ(c.r.X)
	c.t.Tick()
}

func txa(c *cpu) {
	c.r.A = c.r.X
	c.SetNZ(c.r.A)
	c.t.Tick()
}

func txs(c *cpu) {
	c.r.SP = c.r.X
	c.SetNZ(c.r.SP)
	c.t.Tick()
}

func tya(c *cpu) {
	c.r.A = c.r.Y
	c.SetNZ(c.r.A)
	c.t.Tick()
}

func jmpAbsolute(c *cpu) {
	// T1
	c.r.PC++
	addr := uint16(c.m.Read(c.r.PC))
	c.t.Tick()
	// T2
	c.r.PC++
	addr |= (uint16(c.m.Read(c.r.PC)) << 8)
	c.r.PC = addr
	c.t.Tick()
}

func jmpIndirect(c *cpu) {
	// T1
	c.r.PC++
	iAddr := uint16(c.m.Read(c.r.PC))
	c.t.Tick()
	// T2
	c.r.PC++
	iAddr |= (uint16(c.m.Read(c.r.PC)) << 8)
	c.t.Tick()
	// T3
	addr := uint16(c.m.Read(iAddr))
	c.t.Tick()
	// T4
	if (iAddr&0xff == 0xff) && OPTION_BUG_JMP_FF {
		addr |= (uint16(c.m.Read(iAddr&0xff00)) << 8)
	} else {
		addr |= (uint16(c.m.Read(iAddr+1)) << 8)
	}
	c.r.PC = addr
	c.t.Tick()
}

func jsr(c *cpu) {
	// T1
	c.r.PC++
	addr := uint16(c.m.Read(c.r.PC)) // We actually push PC(next) - 1
	c.t.Tick()
	// T2
	c.r.PC++
	c.t.Tick()
	// T3
	c.m.Write(0x100+uint16(c.r.SP), byte(c.r.PC>>8))
	c.r.SP--
	c.t.Tick()
	// T4
	c.m.Write(0x100+uint16(c.r.SP), byte(c.r.PC&0xff))
	c.r.SP--
	c.t.Tick()
	// T5
	addr |= (uint16(c.m.Read(c.r.PC)) << 8)
	c.r.PC = addr
	c.t.Tick()
}

func rts(c *cpu) {
	// T1
	c.t.Tick()
	// T2
	c.r.SP++
	c.t.Tick()
	// T3
	addr := uint16(c.m.Read(0x100 + uint16(c.r.SP)))
	c.r.SP++
	c.t.Tick()
	// T4
	addr |= (uint16(c.m.Read(0x100+uint16(c.r.SP))) << 8)
	c.t.Tick()
	// T5
	c.r.PC = addr + 1 // Since we pushed PC(next) - 1
	c.t.Tick()
}

func rti(c *cpu) {
	// T1
	c.t.Tick()
	// T2
	c.r.SP++
	c.t.Tick()
	// T3
	c.r.P = c.m.Read(0x100 + uint16(c.r.SP))
	c.r.SP++
	// T4
	addr := uint16(c.m.Read(0x100 + uint16(c.r.SP)))
	c.r.SP++
	c.t.Tick()
	// T5
	addr |= (uint16(c.m.Read(0x100+uint16(c.r.SP))) << 8)
	c.r.PC = addr
	c.t.Tick()
}

// Note that BRK skips the next instruction:
// http://en.wikipedia.org/wiki/Interrupts_in_65xx_processors#Using_BRK_and_COP
func brk(c *cpu) {
	// T1
	c.r.PC++
	c.r.SP--
	c.t.Tick()
	// T2
	c.m.Write(0x100+uint16(c.r.SP), byte(c.r.PC>>8))
	c.r.SP--
	c.t.Tick()
	// T3
	c.m.Write(0x100+uint16(c.r.SP), byte(c.r.PC&0xff))
	c.r.SP--
	c.t.Tick()
	// T4
	c.m.Write(0x100+uint16(c.r.SP), c.r.P|FLAG_B) // Set B flag
	c.r.P |= FLAG_I                               // Disable interrupts
	c.t.Tick()
	// T5
	addr := uint16(c.m.Read(IRQ_VECTOR))
	c.t.Tick()
	// T6
	addr |= (uint16(c.m.Read(IRQ_VECTOR+1)) << 8)
	c.r.PC = addr
	c.t.Tick()
}

func branch(mask, value byte) func(*cpu) {
	return func(c *cpu) {
		offset := c.m.Read(c.r.PC)
		c.r.PC++
		c.t.Tick()
		oldPC := c.r.PC
		if c.r.P&mask == value {
			c.t.Tick()
			c.r.PC = c.r.PC + uint16(offset)
			if offset >= 128 {
				c.r.PC = c.r.PC - 256
			}
			if !samePage(c.r.PC, oldPC) {
				c.t.Tick()
			}
		}
	}
}

func immediate(f func(*cpu, byte)) func(*cpu) {
	return func(c *cpu) {
		// T1
		value := c.m.Read(c.r.PC)
		c.r.PC++
		f(c, value)
		c.t.Tick()
	}
}

func lda(c *cpu, value byte) {
	c.r.A = value
	c.SetNZ(value)
}

func ldx(c *cpu, value byte) {
	c.r.X = value
	c.SetNZ(value)
}

func ldy(c *cpu, value byte) {
	c.r.Y = value
	c.SetNZ(value)
}

func ora(c *cpu, value byte) {
	c.r.A |= value
	c.SetNZ(c.r.A)
}

func and(c *cpu, value byte) {
	c.r.A &= value
	c.SetNZ(c.r.A)
}

func eor(c *cpu, value byte) {
	c.r.A ^= value
	c.SetNZ(c.r.A)
}

func cmp(c *cpu, value byte) {
	v := c.r.A - value
	c.SetNZ(v)
}

func cpx(c *cpu, value byte) {
	v := c.r.X - value
	c.SetNZ(v)
}

func cpy(c *cpu, value byte) {
	v := c.r.Y - value
	c.SetNZ(v)
}

var opcodes = map[byte]func(*cpu){
	0x18: clearFlag(FLAG_C), // CLC
	0xD8: clearFlag(FLAG_D), // CLD
	0x58: clearFlag(FLAG_I), // CLI
	0xB8: clearFlag(FLAG_V), // CLV
	0x38: setFlag(FLAG_C),   // SEC
	0xF8: setFlag(FLAG_D),   // SED
	0x78: setFlag(FLAG_I),   // SEI
	0xEA: nop,
	0xAA: tax,
	0xA8: tay,
	0xBA: tsx,
	0x8A: txa,
	0x9A: txs,
	0x98: tya,

	0xCA: dex,
	0x88: dey,
	0xE8: inx,
	0xC8: iny,
	0x48: pha,
	0x08: php,
	0x68: pla,
	0x28: plp,

	0x4C: jmpAbsolute,
	0x6C: jmpIndirect,
	0x20: jsr,
	0x60: rts,
	0x40: rti,
	0x00: brk,

	0x90: branch(FLAG_C, 0),      // BCC
	0xB0: branch(FLAG_C, FLAG_C), // BCS
	0xF0: branch(FLAG_Z, FLAG_Z), // BEQ
	0x30: branch(FLAG_N, FLAG_N), // BMI
	0xD0: branch(FLAG_Z, 0),      // BNE
	0x10: branch(FLAG_N, 0),      // BPL
	0x50: branch(FLAG_V, 0),      // BVC
	0x70: branch(FLAG_V, FLAG_V), // BVS

	0x09: immediate(ora),
	0x29: immediate(and),
	0x49: immediate(eor),
	// 0x69: immediate(adc),
	0xC0: immediate(cpy),
	0xC9: immediate(cmp),
	0xA0: immediate(ldy),
	0xA2: immediate(ldx),
	0xA9: immediate(lda),
	0xE0: immediate(cpx),
	// 0xE9: immediate(sbc),
}
