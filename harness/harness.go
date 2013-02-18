package main

import (
	"fmt"
	"io/ioutil"

	"github.com/zellyn/go6502/cpu"
)

type K64 [65536]byte

func (m *K64) Read(address uint16) byte {
	return m[address]
}

func (m *K64) Write(address uint16, value byte) {
	m[address] = value
}

type CycleCount uint64

func (c *CycleCount) Tick() {
	*c += 1
}

func main() {
	fmt.Println("Hello, world.")
	bytes, err := ioutil.ReadFile("6502_functional_test.bin")
	if err != nil {
		panic("Cannot read file")
	}
	var m K64
	var cc CycleCount
	OFFSET := 0xa
	copy(m[OFFSET:len(bytes)+OFFSET], bytes)
	c := cpu.NewCPU(&m, &cc)
	c.Reset()
	c.SetPC(0x1000)
	for {
		oldPC := c.PC()
		err := c.Step()
		if err != nil {
			fmt.Println(err)
			break
		}
		if c.PC() == oldPC {
			fmt.Printf("Stuck at 0x%X: 0x%X\n", oldPC, m[oldPC])
			break
		}
	}
	fmt.Println("Goodbye, world.")
}
