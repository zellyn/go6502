package cpu

// BUG(zellyn): 6502 should do invalid reads when doing indexed
// addressing across page boundaries. See
// http://en.wikipedia.org/wiki/MOS_Technology_6502#Bugs_and_quirks.

// BUG(zellyn): rmw instructions should write old data back on 6502,
// read twice on 65C02. See
// http://en.wikipedia.org/wiki/MOS_Technology_6502#Bugs_and_quirks.

// BUG(zellyn): Instructions should do many more reads. See
// http://users.telenet.be/kim1-6502/6502/hwman.html#AA and/or table
// 4.1 of "Understanding the Apple II".

// immediate2 performs 2-opcode, 2-cycle immediate mode instructions.
func immediate2(f func(*cpu, byte)) func(*cpu) {
	return func(c *cpu) {
		// T1
		value := c.m.Read(c.r.PC)
		c.r.PC++
		f(c, value)
		c.t()
	}
}

// absolute4r performs 3-opcode, 4-cycle absolute mode read instructions.
func absolute4r(f func(*cpu, byte)) func(*cpu) {
	return func(c *cpu) {
		// T1
		addr := uint16(c.m.Read(c.r.PC))
		c.r.PC++
		c.t()
		// T2
		addr |= (uint16(c.m.Read(c.r.PC)) << 8)
		c.r.PC++
		c.t()
		// T3
		value := c.m.Read(addr)
		f(c, value)
		c.t()
	}
}

// absolute4w performs 3-opcode, 4-cycle absolute mode write instructions.
func absolute4w(f func(*cpu) byte) func(*cpu) {
	return func(c *cpu) {
		// T1
		addr := uint16(c.m.Read(c.r.PC))
		c.r.PC++
		c.t()
		// T2
		addr |= (uint16(c.m.Read(c.r.PC)) << 8)
		c.r.PC++
		c.t()
		// T3
		c.m.Write(addr, f(c))
		c.t()
	}
}

// zp3r performs 2-opcode, 3-cycle zero page read instructions.
func zp3r(f func(*cpu, byte)) func(*cpu) {
	return func(c *cpu) {
		// T1
		addr := uint16(c.m.Read(c.r.PC))
		c.r.PC++
		c.t()
		// T2
		value := c.m.Read(addr)
		f(c, value)
		c.t()
	}
}

// zp3w performs 2-opcode, 3-cycle zero page write instructions.
func zp3w(f func(*cpu) byte) func(*cpu) {
	return func(c *cpu) {
		// T1
		addr := uint16(c.m.Read(c.r.PC))
		c.r.PC++
		c.t()
		// T2
		c.m.Write(addr, f(c))
		c.t()
	}
}

// absx4r performs 3-opcode, 4*-cycle abs,X read instructions.
func absx4r(f func(*cpu, byte)) func(*cpu) {
	return func(c *cpu) {
		// T1
		addr := uint16(c.m.Read(c.r.PC))
		c.r.PC++
		c.t()
		// T2
		addr |= (uint16(c.m.Read(c.r.PC)) << 8)
		addrX := addr + uint16(c.r.X)
		c.r.PC++
		c.t()
		// T3
		if !samePage(addr, addrX) {
			c.m.Read(addrX - 0x100)
			c.t()
		}
		// T3(cotd.) or T4
		value := c.m.Read(addrX)
		f(c, value)
		c.t()
	}
}

// absy4r performs 3-opcode, 4*-cycle abs,Y read instructions.
func absy4r(f func(*cpu, byte)) func(*cpu) {
	return func(c *cpu) {
		// T1
		addr := uint16(c.m.Read(c.r.PC))
		c.r.PC++
		c.t()
		// T2
		addr |= (uint16(c.m.Read(c.r.PC)) << 8)
		addrY := addr + uint16(c.r.Y)
		c.r.PC++
		c.t()
		// T3
		if !samePage(addr, addrY) {
			c.m.Read(addrY - 0x100)
			c.t()
		}
		// T3(cotd.) or T4
		value := c.m.Read(addrY)
		f(c, value)
		c.t()
	}
}

// absx5w performs 3-opcode, 5-cycle abs,X write instructions.
func absx5w(f func(*cpu) byte) func(*cpu) {
	return func(c *cpu) {
		// T1
		addr := uint16(c.m.Read(c.r.PC))
		c.r.PC++
		c.t()
		// T2
		addr |= (uint16(c.m.Read(c.r.PC)) << 8)
		addrX := addr + uint16(c.r.X)
		c.r.PC++
		c.t()
		// T3
		c.m.Read((addr & 0xFF00) | (addrX & 0x00FF))
		c.t()
		// T4
		c.m.Write(addrX, f(c))
		c.t()
	}
}

// absy5w performs 3-opcode, 5-cycle abs,Y write instructions.
func absy5w(f func(*cpu) byte) func(*cpu) {
	return func(c *cpu) {
		// T1
		addr := uint16(c.m.Read(c.r.PC))
		c.r.PC++
		c.t()
		// T2
		addr |= (uint16(c.m.Read(c.r.PC)) << 8)
		addrY := addr + uint16(c.r.Y)
		c.r.PC++
		c.t()
		// T3
		c.m.Read((addr & 0xFF00) | (addrY & 0x00FF))
		c.t()
		// T4
		c.m.Write(addrY, f(c))
		c.t()
	}
}

// zpx4r performs 2-opcode, 4-cycle zp,X read instructions.
func zpx4r(f func(*cpu, byte)) func(*cpu) {
	return func(c *cpu) {
		// T1
		addr := c.m.Read(c.r.PC)
		addrX := uint16(addr + c.r.X)
		c.r.PC++
		c.t()
		// T2
		c.m.Read(uint16(addr))
		c.t()
		// T3
		value := c.m.Read(addrX)
		f(c, value)
		c.t()
	}
}

// zpx4w performs 2-opcode, 4-cycle zp,X write instructions.
func zpx4w(f func(*cpu) byte) func(*cpu) {
	return func(c *cpu) {
		// T1
		addr := c.m.Read(c.r.PC)
		addrX := uint16(addr + c.r.X)
		c.r.PC++
		c.t()
		// T2
		c.m.Read(uint16(addr))
		c.t()
		// T3
		c.m.Write(uint16(addrX), f(c))
		c.t()
	}
}

// zpy4r performs 2-opcode, 4-cycle zp,Y instructions.
func zpy4r(f func(*cpu, byte)) func(*cpu) {
	return func(c *cpu) {
		// T1
		addr := c.m.Read(c.r.PC)
		addrY := uint16(addr + c.r.Y)
		c.r.PC++
		c.t()
		// T2
		c.m.Read(uint16(addr))
		c.t()
		// T3
		value := c.m.Read(uint16(addrY))
		f(c, value)
		c.t()
	}
}

// zpy4w performs 2-opcode, 4-cycle zp,Y write instructions.
func zpy4w(f func(*cpu) byte) func(*cpu) {
	return func(c *cpu) {
		// T1
		addr := c.m.Read(c.r.PC)
		addrY := uint16(addr + c.r.Y)
		c.r.PC++
		c.t()
		// T2
		c.m.Read(uint16(addr))
		c.t()
		// T3
		c.m.Write(addrY, f(c))
		c.t()
	}
}

// zpiy5r performs 2-opcode, 5*-cycle zero-page indirect Y read instructions.
func zpiy5r(f func(*cpu, byte)) func(*cpu) {
	return func(c *cpu) {
		// T1
		iAddr := c.m.Read(c.r.PC)
		c.r.PC++
		c.t()
		// T2
		addr := uint16(c.m.Read(uint16(iAddr)))
		c.t()
		// T3
		addr |= (uint16(c.m.Read(uint16(iAddr+1))) << 8)
		addrY := addr + uint16(c.r.Y)
		c.t()
		// T4
		if !samePage(addr, addrY) {
			c.m.Read((addr & 0xFF00) | (addrY & 0x00FF))
			c.t()
		}
		// T4(cotd.) or T5
		value := c.m.Read(addr + uint16(c.r.Y))
		f(c, value)
		c.t()
	}
}

// zpiy6w performs 2-opcode, 6-cycle zero-page indirect Y write instructions.
func zpiy6w(f func(*cpu) byte) func(*cpu) {
	return func(c *cpu) {
		// T1
		iAddr := c.m.Read(c.r.PC)
		c.r.PC++
		c.t()
		// T2
		addr := uint16(uint16(c.m.Read(uint16(iAddr))))
		c.t()
		// T3
		addr |= (uint16(c.m.Read(uint16(iAddr+1))) << 8)
		addrY := addr + uint16(c.r.Y)
		c.t()
		// T4
		c.m.Read((addr & 0xFF00) | (addrY & 0x00FF))
		c.t()
		// T5
		c.m.Write(addr+uint16(c.r.Y), f(c))
		c.t()
	}
}

// zpxi6r performs 2-opcode, 6-cycle zero-page X indirect read instructions.
func zpxi6r(f func(*cpu, byte)) func(*cpu) {
	return func(c *cpu) {
		// T1
		iAddr := c.m.Read(c.r.PC)
		c.r.PC++
		c.t()
		// T2
		c.m.Read(uint16(iAddr))
		c.t()
		// T3
		addr := uint16(uint16(c.m.Read(uint16(iAddr + c.r.X))))
		c.t()
		// T4
		addr |= (uint16(c.m.Read(uint16(iAddr+c.r.X+1))) << 8)
		c.t()
		// T5
		value := c.m.Read(addr)
		f(c, value)
		c.t()
	}
}

// zpxi6w performs 2-opcode, 6-cycle zero-page X indirect write instructions.
func zpxi6w(f func(*cpu) byte) func(*cpu) {
	return func(c *cpu) {
		// T1
		iAddr := c.m.Read(c.r.PC)
		c.r.PC++
		c.t()
		// T2
		c.m.Read(uint16(iAddr))
		c.t()
		// T3
		addr := uint16(uint16(c.m.Read(uint16(iAddr + c.r.X))))
		c.t()
		// T4
		addr |= (uint16(c.m.Read(uint16(iAddr+c.r.X+1))) << 8)
		c.t()
		// T5
		c.m.Write(addr, f(c))
		c.t()
	}
}

// acc2rmw performs 1-opcode, 2-cycle, accumulator rmw instructions. eg. ASL
func acc2rmw(f func(*cpu, byte) byte) func(*cpu) {
	return func(c *cpu) {
		// T1
		c.m.Read(c.r.PC)
		c.r.A = f(c, c.r.A)
		c.t()
	}
}

// zp5rmw performs 2-opcode, 5-cycle, zp rmw instructions. eg. ASL $70
func zp5rmw(f func(*cpu, byte) byte) func(*cpu) {
	return func(c *cpu) {
		// T1
		addr := uint16(c.m.Read(c.r.PC))
		c.r.PC++
		c.t()
		// T2
		value := c.m.Read(addr)
		c.t()
		// T3
		c.m.Write(addr, value)
		c.t()
		// T4
		c.m.Write(addr, f(c, value))
		c.t()
	}
}

// abs6rmw performs 3-opcode, 6-cycle, abs rmw instructions. eg. ASL $5F72
func abs6rmw(f func(*cpu, byte) byte) func(*cpu) {
	return func(c *cpu) {
		// T1
		addr := uint16(c.m.Read(c.r.PC))
		c.r.PC++
		c.t()
		// T2
		addr |= (uint16(c.m.Read(c.r.PC)) << 8)
		c.r.PC++
		c.t()
		// T3
		value := c.m.Read(addr)
		c.t()
		// T4
		c.m.Write(addr, value) // Spurious write...
		c.t()
		// T5
		c.m.Write(addr, f(c, value))
		c.t()
	}
}

// zpx6rmw performs 2-opcode, 6-cycle, zp,X rmw instructions. eg. ASL $70,X
func zpx6rmw(f func(*cpu, byte) byte) func(*cpu) {
	return func(c *cpu) {
		// T1
		addr8 := c.m.Read(c.r.PC)
		c.r.PC++
		c.t()
		// T2
		c.m.Read(uint16(addr8))
		c.t()
		// T3
		addr := uint16(addr8 + c.r.X)
		value := c.m.Read(addr)
		c.t()
		// T4
		c.m.Write(addr, value)
		c.t()
		// T5
		c.m.Write(addr, f(c, value))
		c.t()
	}
}

// absx7rmw performs 3-opcode, 7-cycle, abs,X rmw instructions. eg. ASL $5F72,X
func absx7rmw(f func(*cpu, byte) byte) func(*cpu) {
	return func(c *cpu) {
		// T1
		addr := uint16(c.m.Read(c.r.PC))
		c.r.PC++
		c.t()
		// T2
		addr |= (uint16(c.m.Read(c.r.PC)) << 8)
		addr += uint16(c.r.X)
		c.r.PC++
		c.t()
		// T3
		c.m.Read(addr)
		c.t()
		// T4
		value := c.m.Read(addr)
		c.t()
		// T5
		c.m.Write(addr, value) // Spurious write
		c.t()
		// T6
		c.m.Write(addr, f(c, value))
		c.t()
	}
}
