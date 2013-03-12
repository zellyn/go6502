/*
Tests for the 6502 CPU emulator, comparing it with the transistor-level simulation.
*/

package tests

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/zellyn/go6502/cpu"
	"github.com/zellyn/go6502/visual"
)

type memOp struct {
	read    bool
	address uint16
	data    byte
}

func (op memOp) String() string {
	rw := "W"
	if op.read {
		rw = "R"
	}
	return fmt.Sprintf("{%s $%04X: %02X}", rw, op.address, op.data)
}

// Memory for the tests. Satisfies the cpu.Memory interface.
type memorizer struct {
	mem1   [65536]byte
	mem2   [65536]byte
	ops    []memOp
	verify bool
	errors []string
}

func (m *memorizer) Reset() {
	m.verify = false
	m.errors = m.errors[:0]
	m.ops = m.ops[:0]
}

func (m *memorizer) Record() {
	m.verify = false
}

func (m *memorizer) Verify() {
	m.verify = true
}

func (m *memorizer) checkOp(newOp memOp) {
	oldOp := m.ops[0]
	m.ops = m.ops[1:]
	if oldOp != newOp {
		// fmt.Println(newOp, "expected:", oldOp)
		m.errors = append(m.errors, fmt.Sprintf("Bad op: %v, expected %v", newOp, oldOp))
	} else {
		// fmt.Println(newOp)
	}
}

func (m *memorizer) Read(address uint16) byte {
	if !m.verify { // recording
		data := m.mem1[address]
		newOp := memOp{true, address, data}
		m.ops = append(m.ops, newOp)
		return data
	}
	data := m.mem2[address]
	m.checkOp(memOp{true, address, data})
	return data
}

func (m *memorizer) Write(address uint16, value byte) {
	newOp := memOp{false, address, value}
	if !m.verify { // recording
		m.ops = append(m.ops, newOp)
		m.mem1[address] = value
		return
	}
	m.checkOp(newOp)
	m.mem2[address] = value
}

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
	var m memorizer
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
	c := cpu.NewCPU(&m, &cc, cpu.VERSION_6502)
	c.Reset()

	m.Reset()
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
