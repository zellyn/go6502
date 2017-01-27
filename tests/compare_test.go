/*
Tests for the 6502 CPU emulator, comparing it with the transistor-level simulation.
*/

package tests

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/zellyn/go6502/asm"
	"github.com/zellyn/go6502/asm/flavors/scma"
	"github.com/zellyn/go6502/asm/lines"
	"github.com/zellyn/go6502/cpu"
	"github.com/zellyn/go6502/opcodes"
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

func getTestProgram() ([]byte, error) {
	o := lines.NewTestOpener()
	a := asm.NewAssembler(scma.New(opcodes.SetSweet16), o)
	o["FILE"] = `
    cld
    ldx #$ff
    txs
    lda #$ff
    sta $70
    lda #$0e
    sta $71
    ldy #0
    adc ($70),Y
    ldy #1
    clc
    adc ($70),Y
    sec
    adc ($70),Y
    ldx #0
    inc $3412,X
    ldx $ff
    clc
    inc $3412,X
    sec
    inc $3412,X
    jmp 0
`
	if err := a.Load("FILE", 0); err != nil {
		return nil, err
	}
	if err := a.Pass2(); err != nil {
		return nil, err
	}
	bb, err := a.RawBytes()
	if err != nil {
		return nil, fmt.Errorf("a.RawBytes() failed: %v", err)
	}
	return bb, nil
}

// Test a custom test program against the instruction- and gate-level
// CPU emulations, making sure they have the same memory access
// patterns.
func TestCustomTestCompare(t *testing.T) {
	bytes, err := getTestProgram()
	if err != nil {
		t.Fatalf("Failed to assemble test program: %v", err)
	}
	var m Memorizer
	START := 0x6000
	copy(m.mem1[START:], bytes)
	copy(m.mem2[START:], bytes)
	// Set the RESET vector to jump to the tests
	m.mem1[0xFFFC] = byte(START % 256)
	m.mem1[0xFFFD] = byte(START / 256)
	m.mem2[0xFFFC] = byte(START % 256)
	m.mem2[0xFFFD] = byte(START / 256)

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
			pc := c.PC()
			fmt.Printf("%04X\n", pc)
			if pc == 0x00 {
				return
			}
		}
	}
}
