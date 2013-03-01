package cpu

// BUG(zellyn): Implement IRQ, NMI.

// BUG(zellyn): implement interrupts, and 6502/65C02
// decimal-mode-clearing and BRK-skipping quirks.  See
// http://en.wikipedia.org/wiki/MOS_Technology_6502#Bugs_and_quirks.

import (
	"fmt"
)

// Chip versions.
type CpuVersion int

const (
	VERSION_6502 CpuVersion = iota
	VERSION_65C02
)

// Interface for the Cpu type.
type Cpu interface {
	Reset()
	Step() error
	SetPC(uint16)
	A() byte
	X() byte
	Y() byte
	PC() uint16
	P() byte // [NV-BDIZC]
	SP() byte
	// BUG(zellyn): Add signaling of interrupts.
}

// Memory interface, for all memory access.
type Memory interface {
	Read(uint16) byte
	Write(uint16, byte)
}

// Ticker interface, for keeping track of cycles.
type Ticker interface {
	Tick()
}

// Interrupt vectors.
const (
	STACK_BASE   = 0x100
	IRQ_VECTOR   = 0xFFFE
	NMI_VECTOR   = 0xFFFA
	RESET_VECTOR = 0xFFFC
)

// Flag masks.
const (
	FLAG_C = 1 << iota
	FLAG_Z
	FLAG_I
	FLAG_D
	FLAG_B
	FLAG_UNUSED
	FLAG_V
	FLAG_N
	FLAG_NV = FLAG_N | FLAG_V
)

type registers struct {
	A  byte
	X  byte
	Y  byte
	PC uint16
	P  byte // [NV-BDIZC]
	SP byte
}

type cpu struct {
	m       Memory
	t       Ticker
	r       registers
	oldPC   uint16
	version CpuVersion
}

// Create and return a new Cpu object with the given memory, ticker, and of the given version.
func NewCPU(memory Memory, ticker Ticker, version CpuVersion) Cpu {
	c := cpu{m: memory, t: ticker, version: version}
	c.r.P |= FLAG_UNUSED // Set unused flag to 1
	return &c
}

func (c *cpu) A() byte {
	return c.r.A
}
func (c *cpu) X() byte {
	return c.r.X
}
func (c *cpu) Y() byte {
	return c.r.Y
}
func (c *cpu) PC() uint16 {
	return c.r.PC
}
func (c *cpu) P() byte {
	return c.r.P
}
func (c *cpu) SP() byte {
	return c.r.SP
}

// Helper for reading a word of memory.
func (c *cpu) readWord(address uint16) uint16 {
	return uint16(c.m.Read(address)) + (uint16(c.m.Read(address+1)) << 8)
}

// Reset performs a reset.
func (c *cpu) Reset() {
	c.r.SP = 0
	c.r.PC = c.readWord(RESET_VECTOR)
	c.r.P |= FLAG_I // Turn interrupts off
	switch c.version {
	case VERSION_6502:
		// 6502 doesn't CLD on interrupt
		c.r.P &^= FLAG_D
	case VERSION_65C02:
	default:
		panic("Unknown chip version")
	}
}

// Step takes a single step (which will last several cycles, calling Tick() on the Ticker for each).
func (c *cpu) Step() error {
	c.oldPC = c.r.PC
	i := c.m.Read(c.r.PC)
	c.r.PC++
	c.t.Tick()

	if f, ok := Opcodes[i]; ok {
		f(c)
		return nil
	}

	return fmt.Errorf("Unknown opcode at location $%04X: $%02X", c.r.PC, i)
}

// Set the program counter.
func (c *cpu) SetPC(address uint16) {
	c.r.PC = address
}
