package cpu

// Helpers and instruction-builders

// setNZ uses the given value to set the N and Z flags.
func (c *cpu) setNZ(value byte) {
	c.r.P = (c.r.P &^ FLAG_N) | (value & FLAG_N)
	if value == 0 {
		c.r.P |= FLAG_Z
	} else {
		c.r.P &^= FLAG_Z
	}
}

// samePage is a helper that returns true if two memory addresses
// refer to the same page.
func samePage(a1 uint16, a2 uint16) bool {
	return a1^a2&0xFF00 == 0
}

// clearFlag builds instructions that clear the flag specified by the
// given mask.
func clearFlag(flag byte) func(*cpu) {
	return func(c *cpu) {
		c.r.P &^= flag
		c.t.Tick()
	}
}

// setFlag builds instructions that set the flag specified by the
// given mask.
func setFlag(flag byte) func(*cpu) {
	return func(c *cpu) {
		c.r.P |= flag
		c.t.Tick()
	}
}

// branch builds instructions that perform branches if the status
// register masks to a given value.
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

// Individual opcodes

func adc(c *cpu, value byte) {
	if c.r.P&FLAG_D > 0 {
		adc_d(c, value)
		return
	}
	result16 := uint16(c.r.A) + uint16(value) + uint16(c.r.P&FLAG_C)
	result := byte(result16)
	c.r.P &^= (FLAG_C | FLAG_V)
	c.r.P |= uint8(result16 >> 8)
	if (c.r.A^result)&(value^result)&0x80 > 0 {
		c.r.P |= FLAG_V
	}
	c.r.A = result
	c.setNZ(result)
}

// adc_d performs decimal-mode add-with-carry.
func adc_d(c *cpu, value byte) {
	// See http://www.6502.org/tutorials/decimal_mode.html#A

	// fmt.Printf("adc_d: $%04X: A=$%02X value=$%02X carry=%d\n",
	//           c.oldPC, c.r.A, value, c.r.P & FLAG_C)

	bin := c.r.A + value + (c.r.P & FLAG_C)
	al := (c.r.A & 0x0F) + (value & 0x0F) + (c.r.P & FLAG_C)
	if al >= 0x0A {
		al = ((al + 0x06) & 0x0F) + 0x10
	}
	// fmt.Printf(" al=$%04X\n", al)
	a := uint16(c.r.A&0xF0) + uint16(value&0xF0) + uint16(al)
	if a >= 0xA0 {
		a += 0x60
	}
	// fmt.Printf(" a=$%04X\n", a)
	a_nv := int16(int8(c.r.A&0xF0)) + int16(int8(value&0xF0)) + int16(int8(al))
	c.r.P &^= (FLAG_V | FLAG_N | FLAG_Z)
	if byte(a_nv&0xFF)&FLAG_N > 0 {
		c.r.P |= FLAG_N
	}
	if a_nv < -128 || a_nv > 127 {
		c.r.P |= FLAG_V
	}
	c.r.A = byte(a & 0xFF)
	// fmt.Printf(" A=$%02X\n", c.r.A)
	c.r.P &^= FLAG_C
	if a >= 0x100 {
		c.r.P |= FLAG_C
	}

	switch c.version {
	case VERSION_6502:
		if bin == 0 {
			c.r.P |= FLAG_Z
		}
	case VERSION_65C02:
		c.t.Tick()
		c.setNZ(byte(a & 0xFF))
	default:
		panic("Unknown chip version")
	}
}

func and(c *cpu, value byte) {
	c.r.A &= value
	c.setNZ(c.r.A)
}

func asl(c *cpu, value byte) byte {
	result := value << 1
	c.r.P = (c.r.P &^ FLAG_C) | (value >> 7)
	c.setNZ(result)
	return result
}

func bit(c *cpu, value byte) {
	if t := c.r.A & value; t == 0 {
		c.r.P |= FLAG_Z
	} else {
		c.r.P &^= FLAG_Z
	}
	c.r.P = (c.r.P &^ FLAG_NV) | (value & FLAG_NV)
}

// Note that BRK skips the next instruction:
// http://en.wikipedia.org/wiki/Interrupts_in_65xx_processors#Using_BRK_and_COP
func brk(c *cpu) {
	// T1
	c.r.PC++
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
	c.r.SP--
	c.r.P |= FLAG_I // Disable interrupts
	c.t.Tick()
	// T5
	addr := uint16(c.m.Read(IRQ_VECTOR))
	c.t.Tick()
	// T6
	addr |= (uint16(c.m.Read(IRQ_VECTOR+1)) << 8)
	c.r.PC = addr
	c.t.Tick()
}

func cmp(c *cpu, value byte) {
	v := c.r.A - value
	c.r.P &^= FLAG_C
	if c.r.A >= value {
		c.r.P |= FLAG_C
	}
	c.setNZ(v)
}

func cpx(c *cpu, value byte) {
	v := c.r.X - value
	c.r.P &^= FLAG_C
	if c.r.X >= value {
		c.r.P |= FLAG_C
	}
	c.setNZ(v)
}

func cpy(c *cpu, value byte) {
	v := c.r.Y - value
	c.r.P &^= FLAG_C
	if c.r.Y >= value {
		c.r.P |= FLAG_C
	}
	c.setNZ(v)
}

func dec(c *cpu, value byte) byte {
	result := value - 1
	c.setNZ(result)
	return result
}

func dex(c *cpu) {
	c.r.X--
	c.setNZ(c.r.X)
	c.t.Tick()
}

func dey(c *cpu) {
	c.r.Y--
	c.setNZ(c.r.Y)
	c.t.Tick()
}

func eor(c *cpu, value byte) {
	c.r.A ^= value
	c.setNZ(c.r.A)
}

func inc(c *cpu, value byte) byte {
	result := value + 1
	c.setNZ(result)
	return result
}

func inx(c *cpu) {
	c.r.X++
	c.setNZ(c.r.X)
	c.t.Tick()
}

func iny(c *cpu) {
	c.r.Y++
	c.setNZ(c.r.Y)
	c.t.Tick()
}

func jmpAbsolute(c *cpu) {
	// T1
	addr := uint16(c.m.Read(c.r.PC))
	c.r.PC++
	c.t.Tick()
	// T2
	addr |= (uint16(c.m.Read(c.r.PC)) << 8)
	c.r.PC++
	c.r.PC = addr
	c.t.Tick()
}

func jmpIndirect(c *cpu) {
	// T1
	iAddr := uint16(c.m.Read(c.r.PC))
	c.r.PC++
	c.t.Tick()
	// T2
	iAddr |= (uint16(c.m.Read(c.r.PC)) << 8)
	c.r.PC++
	c.t.Tick()
	// T3
	addr := uint16(c.m.Read(iAddr))
	c.t.Tick()
	// T4
	// 6502 jumps to (xxFF,xx00) instead of (xxFF,xxFF+1).
	// See http://en.wikipedia.org/wiki/MOS_Technology_6502#Bugs_and_quirks
	switch c.version {
	case VERSION_6502:
		if iAddr&0xff == 0xff {
			addr |= (uint16(c.m.Read(iAddr&0xff00)) << 8)
		} else {
			addr |= (uint16(c.m.Read(iAddr+1)) << 8)
		}
	case VERSION_65C02:
		addr |= (uint16(c.m.Read(iAddr+1)) << 8)
	default:
		panic("Unknown chip version")
	}
	c.r.PC = addr
	c.t.Tick()
}

func jsr(c *cpu) {
	// T1
	addr := uint16(c.m.Read(c.r.PC)) // We actually push PC(next) - 1
	c.r.PC++
	c.t.Tick()
	// T2
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

func lda(c *cpu, value byte) {
	c.r.A = value
	c.setNZ(value)
}

func ldx(c *cpu, value byte) {
	c.r.X = value
	c.setNZ(value)
}

func ldy(c *cpu, value byte) {
	c.r.Y = value
	c.setNZ(value)
}

func lsr(c *cpu, value byte) byte {
	result := (value >> 1)
	c.r.P = (c.r.P &^ FLAG_C) | (value & FLAG_C)
	c.setNZ(result)
	return result
}

func ora(c *cpu, value byte) {
	c.r.A |= value
	c.setNZ(c.r.A)
}

func nop(c *cpu) {
	c.t.Tick()
}

func pha(c *cpu) {
	c.t.Tick()
	c.m.Write(0x100+uint16(c.r.SP), c.r.A)
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

func php(c *cpu) {
	c.t.Tick()
	c.m.Write(0x100+uint16(c.r.SP), c.r.P)
	c.r.SP--
	c.t.Tick()
}
func plp(c *cpu) {
	c.t.Tick()
	c.r.SP++
	c.t.Tick()
	c.r.P = c.m.Read(0x100+uint16(c.r.SP)) | FLAG_UNUSED | FLAG_B
	c.t.Tick()
}

func rol(c *cpu, value byte) byte {
	result := value<<1 | (c.r.P & FLAG_C)
	c.r.P = (c.r.P &^ FLAG_C) | (value >> 7)
	c.setNZ(result)
	return result
}

func ror(c *cpu, value byte) byte {
	result := (value >> 1) | (c.r.P << 7)
	c.r.P = (c.r.P &^ FLAG_C) | (value & FLAG_C)
	c.setNZ(result)
	return result
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
	c.r.P = c.m.Read(0x100+uint16(c.r.SP)) | FLAG_UNUSED
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

func sbc(c *cpu, value byte) {
	if c.r.P&FLAG_D > 0 {
		sbc_d(c, value)
		return
	} else {
		c.r.A = sbc_bin(c, value)
	}
}

// sbc_bin performs binary-mode subtract with carry. Broken out into a
// separate routine so sbc_d can call it to determine flag values.
func sbc_bin(c *cpu, value byte) byte {
	// Same as adc, except we take the ones complement of value
	value = ^value
	result16 := uint16(c.r.A) + uint16(value) + uint16(c.r.P&FLAG_C)
	result := byte(result16)
	c.r.P &^= (FLAG_C | FLAG_V)
	c.r.P |= uint8(result16 >> 8)
	if (c.r.A^result)&(value^result)&0x80 > 0 {
		c.r.P |= FLAG_V
	}
	c.setNZ(result)
	return result
}

// sdc_d performs decimal-mode subtract-with-carry.
func sbc_d(c *cpu, value byte) {
	// See http://www.6502.org/tutorials/decimal_mode.html#A

	carry := c.r.P & FLAG_C

	// fmt.Printf("sbc_d: $%04X: A=$%02X value=$%02X carry=%d\n",
	//            c.oldPC, c.r.A, value, carry)

	// Compute normal sbc, and set all flags accordingly
	sbc_bin(c, value)

	switch c.version {
	case VERSION_6502:
		al := int16(c.r.A&0x0F) - int16(value&0x0F) + int16(carry) - 1
		if al < 0 {
			al = ((al - 0x06) & 0x0F) - 0x10
		}
		// fmt.Printf(" al=$%04X\n", al)
		a := int16(c.r.A&0xF0) - int16(value&0xF0) + al
		if a < 0 {
			a = a - 0x60
		}
		// fmt.Printf(" a=$%04X\n", a)
		c.r.A = byte(a)
	case VERSION_65C02:
		al := int16(c.r.A&0x0F) - int16(value&0x0F) + int16(carry) - 1
		a := int16(c.r.A) - int16(value) + int16(carry) - 1
		// fmt.Printf(" al=$%04X\n", al)
		if a < 0 {
			a = a - 0x60
		}
		if al < 0 {
			a = a - 0x06
		}
		// fmt.Printf(" a=$%04X ($%02X)\n", a, byte(a))
		c.r.A = byte(a)
		c.t.Tick()
		c.setNZ(c.r.A)
	default:
		panic("Unknown chip version")
	}
}

func sta(c *cpu) byte {
	return c.r.A
}

func stx(c *cpu) byte {
	return c.r.X
}

func sty(c *cpu) byte {
	return c.r.Y
}

func tax(c *cpu) {
	c.r.X = c.r.A
	c.setNZ(c.r.X)
	c.t.Tick()
}

func tay(c *cpu) {
	c.r.Y = c.r.A
	c.setNZ(c.r.Y)
	c.t.Tick()
}

func tsx(c *cpu) {
	c.r.X = c.r.SP
	c.setNZ(c.r.X)
	c.t.Tick()
}

func txa(c *cpu) {
	c.r.A = c.r.X
	c.setNZ(c.r.A)
	c.t.Tick()
}

func txs(c *cpu) {
	c.r.SP = c.r.X
	c.t.Tick()
}

func tya(c *cpu) {
	c.r.A = c.r.Y
	c.setNZ(c.r.A)
	c.t.Tick()
}
