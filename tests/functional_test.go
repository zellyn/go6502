/*
Tests for the 6502 CPU emulator.
*/
package tests

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"testing"

	"github.com/zellyn/go6502/asm"
	"github.com/zellyn/go6502/cpu"
	"github.com/zellyn/go6502/visual"
)

// Memory for the tests. Satisfies the cpu.Memory interface.
type K64 [65536]byte

func (m *K64) Read(address uint16) byte {
	return m[address]
}
func (m *K64) Write(address uint16, value byte) {
	m[address] = value
}

// Cycle counter for the tests. Satisfies the cpu.Ticker interface.
type CycleCount uint64

func (c *CycleCount) Tick() {
	*c += 1
}

func randomize(k *K64) {
	for i := 0; i < 65536; i++ {
		k[i] = byte(rand.Int())
	}
}

// status prints out the current CPU instruction and register status.
func status(c cpu.Cpu, m *[65536]byte, cc uint64) string {
	bytes, text, _ := asm.Disasm(c.PC(), m[c.PC()], m[c.PC()+1], m[c.PC()+2])
	return fmt.Sprintf("$%04X: %s  %s  A=$%02X X=$%02X Y=$%02X SP=$%02X P=$%08b - %d\n",
		c.PC(), bytes, text, c.A(), c.X(), c.Y(), c.SP(), c.P(), cc)
}

// Run Klaus Dormann's amazing comprehensive test against the
// instruction-level CPU emulation.
func TestFunctionalTestInstructions(t *testing.T) {
	unused := map[byte]bool{}
	for k := range cpu.Opcodes {
		unused[k] = true
	}
	bytes, err := ioutil.ReadFile("6502_functional_test.bin")
	if err != nil {
		panic("Cannot read file")
	}
	var m K64
	randomize(&m)
	var cc CycleCount
	OFFSET := 0xa
	copy(m[OFFSET:len(bytes)+OFFSET], bytes)
	c := cpu.NewCPU(&m, &cc, cpu.VERSION_6502)
	c.Reset()
	c.SetPC(0x1000)
	for {
		unused[m[c.PC()]] = false
		oldPC := c.PC()
		// status(c, m, cc)
		err := c.Step()
		if err != nil {
			t.Error(err)
			break
		}
		if c.PC() == oldPC {
			if c.PC() != 0x3CC5 {
				t.Errorf("Stuck at $%04X: 0x%02X", oldPC, m[oldPC])
			}
			return
		}
	}
	for k, v := range unused {
		if v {
			t.Errorf("Unused instruction: 0x%2X", k)
		}
	}
}

// Run the first few thousand steps of Klaus Dormann's amazing
// comprehensive test against the gate-level CPU emulation.
func TestFunctionalTestGates(t *testing.T) {
	expected := []uint16{
		0,
		4327,
		4288,
		4253,
		4221,
		4194,
		4171,
		4152,
		4137,
		4105,
		4374,
		4368,
	}
	bytes, err := ioutil.ReadFile("6502_functional_test.bin")
	if err != nil {
		panic("Cannot read file")
	}
	var m K64
	randomize(&m)
	OFFSET := 0xa
	copy(m[OFFSET:len(bytes)+OFFSET], bytes)
	// Set the RESET vector to jump to the tests
	m[0xFFFC] = 0x00
	m[0xFFFD] = 0x10
	c := visual.NewCPU(&m)
	c.Reset()
	count := uint64(0)
	for {
		count++
		if count%1000 == 0 {
			pc := c.PC()
			if pc == 0x3CC5 {
				return
			}
			if count/1000 < uint64(len(expected)) {
				expectedPC := expected[count/1000]
				if pc != expectedPC {
					t.Errorf("Expected PC=%d at step %d, got %d", expectedPC, count, pc)
				}
			} else {
				return
			}
		}
		// status(c, m, cc)
		c.Step()
	}
}

// Run Bruce Clark's decimal test in 6502 mode.
func TestDecimalMode6502(t *testing.T) {
	bytes, err := ioutil.ReadFile("decimal_mode.bin")
	if err != nil {
		panic("Cannot read file")
	}
	var m K64
	randomize(&m)
	var cc CycleCount
	OFFSET := 0x1000
	copy(m[OFFSET:len(bytes)+OFFSET], bytes)
	m[1] = 0 // 6502
	c := cpu.NewCPU(&m, &cc, cpu.VERSION_6502)
	c.Reset()
	c.SetPC(0x1000)
	for {
		oldPC := c.PC()
		// status(c, m, cc)
		err := c.Step()
		if err != nil {
			t.Error(err)
			break
		}
		if c.PC() == oldPC {
			if c.PC() != 0x1037 {
				t.Errorf("Stuck at 0x%X: 0x%X\n", oldPC, m[oldPC])
			}
			break
		}
	}
	error := m[0]
	if error > 0 {
		t.Errorf("Decimal mode test failed: error=%d", error)
	}
}

// Run Bruce Clark's decimal test in 65C02 mode.
func TestDecimalMode65C02(t *testing.T) {
	bytes, err := ioutil.ReadFile("decimal_mode.bin")
	if err != nil {
		panic("Cannot read file")
	}
	var m K64
	randomize(&m)
	var cc CycleCount
	OFFSET := 0x1000
	copy(m[OFFSET:len(bytes)+OFFSET], bytes)
	m[1] = 1 // 65C02
	c := cpu.NewCPU(&m, &cc, cpu.VERSION_65C02)
	c.Reset()
	c.SetPC(0x1000)
	for {
		oldPC := c.PC()
		// status(c, m, cc)
		err := c.Step()
		if err != nil {
			t.Error(err)
			break
		}
		if c.PC() == oldPC {
			if c.PC() != 0x1037 {
				t.Errorf("Stuck at 0x%X: 0x%X\n", oldPC, m[oldPC])
			}
			break
		}
	}
	error := m[0]
	if error > 0 {
		fmt.Printf("N1=$%02X N2=$%02X DA=$%02X DNVZC=$%02X - AR=$%02X NF=$%02X VF=$%02X ZF=$%02X CF=$%02X\n",
			m[0x08], m[0x0B], m[0x04], m[0x05], m[0x02], m[0x0D], m[0x0E], m[0x0F], m[0x03])
		t.Errorf("Decimal mode test failed: error=%d", error)
	}
}
