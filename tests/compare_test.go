/*
Tests for the 6502 CPU emulator, comparing it with the transistor-level simulation.
*/

package tests

import (
	"io/ioutil"
	"testing"

	"github.com/zellyn/go6502/cpu"
	"github.com/zellyn/go6502/visual"
)

// Run the first few thousand steps of Klaus Dormann's comprehensive
// test against the instruction- and gate-level CPU emulations, making
// sure they have the same memory access patterns.
func TestFunctionalTestCompare(t *testing.T) {
	_ = []uint16{
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
	var m Memorizer
	OFFSET := 0xa
	copy(m.mem1[OFFSET:len(bytes)+OFFSET], bytes)
	copy(m.mem2[OFFSET:len(bytes)+OFFSET], bytes)
	// Set the RESET vector to jump to the tests
	m.mem1[0xFFFC] = 0x00
	m.mem1[0xFFFD] = 0x10
	m.mem2[0xFFFC] = 0x00
	m.mem2[0xFFFD] = 0x10

	v := visual.NewCPU(&m)
	v.Reset()
	for i := 0; i < 8; i++ {
		v.Step()
	}

	var cc CycleCount
	c := cpu.NewCPU(&m, cc.Tick, cpu.VERSION_6502)
	c.Reset()

	m.Reset(MODE_RECORD)
	v.Step()
	v.Step()
	m.Verify()
	c.Step()
	if len(m.errors) > 0 {
		t.Fatal("Errors on reset", m.errors)
	}

	for {
		m.Record()
		for i := 0; i < 1000; i++ {
			v.Step()
		}
		m.Verify()
		for len(m.ops) > 7 {
			s := status(c, &m.mem2, uint64(cc))
			c.Step()
			if len(m.errors) > 0 {
				t.Fatalf("Error at %v: %v", s, m.errors)
			}
			if cc%1000 == 0 {
				pc := c.PC()
				if cc > 20000 || pc == 0x3CC5 {
					return
				}
			}
		}
	}
}
