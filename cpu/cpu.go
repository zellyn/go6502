/*
  6502 implementation.

  TODO(zellyn): Provide configurable options
  TODO(zellyn): Implement IRQ, NMI
  TODO(zellyn): Does BRK on 65C02 CLD?
*/

package cpu

import (
	"fmt"
)

// See http://en.wikipedia.org/wiki/MOS_Technology_6502#Bugs_and_quirks
const (
	VERSION_6502 = iota
	VERSION_65C02
)

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
	// TODO(zellyn): Add signaling of interrupts
}

type Memory interface {
	Read(uint16) byte
	Write(uint16, byte)
}

type Ticker interface {
	Tick()
}

const (
	STACK_BASE   = 0x100
	IRQ_VECTOR   = 0xFFFE
	NMI_VECTOR   = 0xFFFA
	RESET_VECTOR = 0xFFFC
)

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
	version int
}

func NewCPU(memory Memory, ticker Ticker, version int) Cpu {
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

func (c *cpu) readWord(address uint16) uint16 {
	return uint16(c.m.Read(address)) + (uint16(c.m.Read(address+1)) << 8)
}

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

func (c *cpu) Step() error {
	c.oldPC = c.r.PC
	i := c.m.Read(c.r.PC)
	c.r.PC++
	c.t.Tick()

	if op, ok := Opcodes[i]; ok {
		op.function(c)
		return nil
	}

	return fmt.Errorf("Unknown opcode at location $%04X: $%02X", c.r.PC, i)
}

func (c *cpu) SetPC(address uint16) {
	c.r.PC = address
}

func (c *cpu) SetNZ(value byte) {
	c.r.P = (c.r.P &^ FLAG_N) | (value & FLAG_N)
	if value == 0 {
		c.r.P |= FLAG_Z
	} else {
		c.r.P &^= FLAG_Z
	}
}
