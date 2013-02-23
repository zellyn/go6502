package cpu

// BUG(zellyn): 6502 should do invalid reads when doing indexed
// addressing across page boundaries. See
// http://en.wikipedia.org/wiki/MOS_Technology_6502#Bugs_and_quirks.

// BUG(zellyn): rmw instructions should write old data back on 6502,
// read twice on 65C02. See
// http://en.wikipedia.org/wiki/MOS_Technology_6502#Bugs_and_quirks.

// immediate2 performs 2-opcode, 2-cycle immediate mode instructions.
func immediate2(f func(*cpu, byte)) func(*cpu) {
	return func(c *cpu) {
		// T1
		value := c.m.Read(c.r.PC)
		c.r.PC++
		f(c, value)
		c.t.Tick()
	}
}

// absolute4r performs 3-opcode, 4-cycle absolute mode read instructions.
func absolute4r(f func(*cpu, byte)) func(*cpu) {
	return func(c *cpu) {
		// T1
		addr := uint16(c.m.Read(c.r.PC))
		c.r.PC++
		c.t.Tick()
		// T2
		addr |= (uint16(c.m.Read(c.r.PC)) << 8)
		c.r.PC++
		c.t.Tick()
		// T3
		value := c.m.Read(addr)
		f(c, value)
		c.t.Tick()
	}
}

// absolute4w performs 3-opcode, 4-cycle absolute mode write instructions.
func absolute4w(f func(*cpu) byte) func(*cpu) {
	return func(c *cpu) {
		// T1
		addr := uint16(c.m.Read(c.r.PC))
		c.r.PC++
		c.t.Tick()
		// T2
		addr |= (uint16(c.m.Read(c.r.PC)) << 8)
		c.r.PC++
		c.t.Tick()
		// T3
		c.m.Write(addr, f(c))
		c.t.Tick()
	}
}

// zp3r performs 2-opcode, 3-cycle zero page read instructions.
func zp3r(f func(*cpu, byte)) func(*cpu) {
	return func(c *cpu) {
		// T1
		addr := uint16(c.m.Read(c.r.PC))
		c.r.PC++
		c.t.Tick()
		// T2
		value := c.m.Read(addr)
		f(c, value)
		c.t.Tick()
	}
}

// zp3w performs 2-opcode, 3-cycle zero page write instructions.
func zp3w(f func(*cpu) byte) func(*cpu) {
	return func(c *cpu) {
		// T1
		addr := uint16(c.m.Read(c.r.PC))
		c.r.PC++
		c.t.Tick()
		// T2
		c.m.Write(addr, f(c))
		c.t.Tick()
	}
}

// absx4r performs 3-opcode, 4*-cycle abs,X read instructions.
func absx4r(f func(*cpu, byte)) func(*cpu) {
	return func(c *cpu) {
		// T1
		addr := uint16(c.m.Read(c.r.PC))
		c.r.PC++
		c.t.Tick()
		// T2
		addr |= (uint16(c.m.Read(c.r.PC)) << 8)
		c.r.PC++
		c.t.Tick()
		// T3
		if !samePage(addr, addr+uint16(c.r.X)) {
			c.t.Tick()
		}
		// T3(cotd.) or T4
		value := c.m.Read(addr + uint16(c.r.X))
		f(c, value)
		c.t.Tick()
	}
}

// absy4r performs 3-opcode, 4*-cycle abs,Y read instructions.
func absy4r(f func(*cpu, byte)) func(*cpu) {
	return func(c *cpu) {
		// T1
		addr := uint16(c.m.Read(c.r.PC))
		c.r.PC++
		c.t.Tick()
		// T2
		addr |= (uint16(c.m.Read(c.r.PC)) << 8)
		c.r.PC++
		c.t.Tick()
		// T3
		if !samePage(addr, addr+uint16(c.r.Y)) {
			c.t.Tick()
		}
		// T3(cotd.) or T4
		value := c.m.Read(addr + uint16(c.r.Y))
		f(c, value)
		c.t.Tick()
	}
}

// absx5w performs 3-opcode, 5-cycle abs,X write instructions.
func absx5w(f func(*cpu) byte) func(*cpu) {
	return func(c *cpu) {
		// T1
		addr := uint16(c.m.Read(c.r.PC))
		c.r.PC++
		c.t.Tick()
		// T2
		addr |= (uint16(c.m.Read(c.r.PC)) << 8)
		c.r.PC++
		c.t.Tick()
		// T3
		c.t.Tick()
		// T4
		c.m.Write(addr+uint16(c.r.X), f(c))
		c.t.Tick()
	}
}

// absy5w performs 3-opcode, 5-cycle abs,Y write instructions.
func absy5w(f func(*cpu) byte) func(*cpu) {
	return func(c *cpu) {
		// T1
		addr := uint16(c.m.Read(c.r.PC))
		c.r.PC++
		c.t.Tick()
		// T2
		addr |= (uint16(c.m.Read(c.r.PC)) << 8)
		c.r.PC++
		c.t.Tick()
		// T3
		c.t.Tick()
		// T4
		c.m.Write(addr+uint16(c.r.Y), f(c))
		c.t.Tick()
	}
}

// zpx4r performs 2-opcode, 4-cycle zp,X read instructions.
func zpx4r(f func(*cpu, byte)) func(*cpu) {
	return func(c *cpu) {
		// T1
		addr := c.m.Read(c.r.PC)
		c.r.PC++
		c.t.Tick()
		// T2
		c.t.Tick()
		// T3
		addr += c.r.X
		value := c.m.Read(uint16(addr))
		f(c, value)
		c.t.Tick()
	}
}

// zpx4w performs 2-opcode, 4-cycle zp,X write instructions.
func zpx4w(f func(*cpu) byte) func(*cpu) {
	return func(c *cpu) {
		// T1
		addr := c.m.Read(c.r.PC)
		c.r.PC++
		c.t.Tick()
		// T2
		c.t.Tick()
		// T3
		addr += c.r.X
		c.m.Write(uint16(addr), f(c))
		c.t.Tick()
	}
}

// zpy4r performs 2-opcode, 4-cycle zp,Y instructions.
func zpy4r(f func(*cpu, byte)) func(*cpu) {
	return func(c *cpu) {
		// T1
		addr := c.m.Read(c.r.PC)
		c.r.PC++
		c.t.Tick()
		// T2
		c.t.Tick()
		// T3
		addr += c.r.Y
		value := c.m.Read(uint16(addr))
		f(c, value)
		c.t.Tick()
	}
}

// zpy4w performs 2-opcode, 4-cycle zp,Y write instructions.
func zpy4w(f func(*cpu) byte) func(*cpu) {
	return func(c *cpu) {
		// T1
		addr := c.m.Read(c.r.PC)
		c.r.PC++
		c.t.Tick()
		// T2
		c.t.Tick()
		// T3
		addr += c.r.Y
		c.m.Write(uint16(addr), f(c))
		c.t.Tick()
	}
}

// zpiy5r performs 2-opcode, 5*-cycle zero-page indirect Y read instructions.
func zpiy5r(f func(*cpu, byte)) func(*cpu) {
	return func(c *cpu) {
		// T1
		iAddr := c.m.Read(c.r.PC)
		c.r.PC++
		c.t.Tick()
		// T2
		addr := uint16(uint16(c.m.Read(uint16(iAddr))))
		c.t.Tick()
		// T3
		addr |= (uint16(c.m.Read(uint16(iAddr+1))) << 8)
		c.t.Tick()
		// T4
		if !samePage(addr, addr+uint16(c.r.Y)) {
			c.t.Tick()
		}
		// T4(cotd.) or T5
		value := c.m.Read(addr + uint16(c.r.Y))
		f(c, value)
		c.t.Tick()
	}
}

// zpiy6w performs 2-opcode, 6-cycle zero-page indirect Y write instructions.
func zpiy6w(f func(*cpu) byte) func(*cpu) {
	return func(c *cpu) {
		// T1
		iAddr := c.m.Read(c.r.PC)
		c.r.PC++
		c.t.Tick()
		// T2
		addr := uint16(uint16(c.m.Read(uint16(iAddr))))
		c.t.Tick()
		// T3
		addr |= (uint16(c.m.Read(uint16(iAddr+1))) << 8)
		c.t.Tick()
		// T4
		c.t.Tick()
		// T5
		c.m.Write(addr+uint16(c.r.Y), f(c))
		c.t.Tick()
	}
}

// zpxi6r performs 2-opcode, 6-cycle zero-page X indirect read instructions.
func zpxi6r(f func(*cpu, byte)) func(*cpu) {
	return func(c *cpu) {
		// T1
		iAddr := c.m.Read(c.r.PC)
		c.r.PC++
		c.t.Tick()
		// T2
		c.t.Tick()
		// T3
		addr := uint16(uint16(c.m.Read(uint16(iAddr + c.r.X))))
		c.t.Tick()
		// T4
		addr |= (uint16(c.m.Read(uint16(iAddr+c.r.X+1))) << 8)
		c.t.Tick()
		// T5
		value := c.m.Read(addr)
		f(c, value)
		c.t.Tick()
	}
}

// zpxi6w performs 2-opcode, 6-cycle zero-page X indirect write instructions.
func zpxi6w(f func(*cpu) byte) func(*cpu) {
	return func(c *cpu) {
		// T1
		iAddr := c.m.Read(c.r.PC)
		c.r.PC++
		c.t.Tick()
		// T2
		c.t.Tick()
		// T3
		addr := uint16(uint16(c.m.Read(uint16(iAddr + c.r.X))))
		c.t.Tick()
		// T4
		addr |= (uint16(c.m.Read(uint16(iAddr+c.r.X+1))) << 8)
		c.t.Tick()
		// T5
		c.m.Write(addr, f(c))
		c.t.Tick()
	}
}

// acc2rmw performs 1-opcode, 2-cycle, accumulator rmw instructions.
func acc2rmw(f func(*cpu, byte) byte) func(*cpu) {
	return func(c *cpu) {
		// T1
		c.r.A = f(c, c.r.A)
		c.t.Tick()
	}
}

// zp5rmw performs 2-opcode, 5-cycle, zp rmw instructions.
func zp5rmw(f func(*cpu, byte) byte) func(*cpu) {
	return func(c *cpu) {
		// T1
		addr := uint16(c.m.Read(c.r.PC))
		c.r.PC++
		c.t.Tick()
		// T2
		value := c.m.Read(addr)
		c.t.Tick()
		// T3
		c.t.Tick()
		// T4
		c.m.Write(addr, f(c, value))
		c.t.Tick()
	}
}

// abs6rmw performs 3-opcode, 6-cycle, abs rmw instructions.
func abs6rmw(f func(*cpu, byte) byte) func(*cpu) {
	return func(c *cpu) {
		// T1
		addr := uint16(c.m.Read(c.r.PC))
		c.r.PC++
		c.t.Tick()
		// T2
		addr |= (uint16(c.m.Read(c.r.PC)) << 8)
		c.r.PC++
		c.t.Tick()
		// T3
		value := c.m.Read(addr)
		c.t.Tick()
		// T4
		c.m.Write(addr, value) // Spurious write...
		c.t.Tick()
		// T5
		c.m.Write(addr, f(c, value))
		c.t.Tick()
	}
}

// zpx6rmw performs 2-opcode, 6-cycle, zp,X rmw instructions.
func zpx6rmw(f func(*cpu, byte) byte) func(*cpu) {
	return func(c *cpu) {
		// T1
		addr8 := c.m.Read(c.r.PC)
		c.r.PC++
		c.t.Tick()
		// T2
		c.t.Tick()
		// T3
		addr := uint16(addr8 + c.r.X)
		value := c.m.Read(addr)
		c.t.Tick()
		// T4
		c.m.Write(addr, value)
		c.t.Tick()
		// T5
		c.m.Write(addr, f(c, value))
		c.t.Tick()
	}
}

// absx7rmw performs 3-opcode, 7-cycle, abs,X rmw instructions.
func absx7rmw(f func(*cpu, byte) byte) func(*cpu) {
	return func(c *cpu) {
		// T1
		addr := uint16(c.m.Read(c.r.PC))
		c.r.PC++
		c.t.Tick()
		// T2
		addr |= (uint16(c.m.Read(c.r.PC)) << 8)
		c.r.PC++
		c.t.Tick()
		// T3
		c.t.Tick()
		// T4
		value := c.m.Read(addr + uint16(c.r.X))
		c.t.Tick()
		// T5
		c.m.Write(addr+uint16(c.r.X), value) // Spurious write
		c.t.Tick()
		// T6
		c.m.Write(addr+uint16(c.r.X), f(c, value))
		c.t.Tick()
	}
}
