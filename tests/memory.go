package tests

import (
	"fmt"
)

type MemOp struct {
	read    bool
	address uint16
	data    byte
}

func (op MemOp) String() string {
	rw := "W"
	if op.read {
		rw = "R"
	}
	return fmt.Sprintf("{%s $%04X: %02X}", rw, op.address, op.data)
}

type Mode int

const (
	MODE_SAVE_LAST Mode = iota // Just hold the last action
	MODE_RECORD                // Record all memory actions
	MODE_VERIFY                // Verify that the recorded actions match
)

// Memory for the tests. Satisfies the cpu.Memory interface.
type Memorizer struct {
	mem1   [65536]byte
	mem2   [65536]byte
	ops    []MemOp
	mode   Mode
	errors []string
	last   MemOp
}

func (m *Memorizer) Reset(mode Mode) {
	m.mode = mode
	m.errors = m.errors[:0]
	m.ops = m.ops[:0]
}

func (m *Memorizer) Record() {
	m.mode = MODE_RECORD
}

func (m *Memorizer) Verify() {
	m.mode = MODE_VERIFY
}

func (m *Memorizer) SaveLast() {
	m.mode = MODE_SAVE_LAST
}

func (m *Memorizer) checkOp(newOp MemOp) {
	oldOp := m.ops[0]
	m.ops = m.ops[1:]
	if oldOp != newOp {
		// fmt.Println(newOp, "expected:", oldOp)
		m.errors = append(m.errors, fmt.Sprintf("Bad op: %v, expected %v", newOp, oldOp))
	} else {
		// fmt.Println(newOp)
	}
}

func (m *Memorizer) Read(address uint16) byte {
	switch m.mode {
	case MODE_RECORD:
		data := m.mem1[address]
		m.ops = append(m.ops, MemOp{true, address, data})
		return data
	case MODE_VERIFY:
		data := m.mem2[address]
		m.checkOp(MemOp{true, address, data})
		return data
	case MODE_SAVE_LAST:
		data := m.mem1[address]
		m.last = MemOp{true, address, data}
		return data
	}
	panic("Unknown MODE")
}

func (m *Memorizer) Write(address uint16, value byte) {
	newOp := MemOp{false, address, value}
	switch m.mode {
	case MODE_RECORD:
		m.ops = append(m.ops, newOp)
		m.mem1[address] = value
		return
	case MODE_VERIFY:
		m.checkOp(newOp)
		m.mem2[address] = value
		return
	case MODE_SAVE_LAST:
		m.last = newOp
		m.mem1[address] = value
		return
	}
	panic("Unknown MODE")
}
