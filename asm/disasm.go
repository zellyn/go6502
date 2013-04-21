/* 
Package asm provides routines for decompiling 6502 assembly language.
*/
package asm

import (
	"fmt"
	"io"

	"github.com/zellyn/go6502/opcodes"
)

// bytesString takes three bytes and a length, returning the formatted
// hex bytes for an instrction of the given length.
func bytesString(byte0, byte1, byte2 byte, length int) string {
	switch length {
	case 1:
		return fmt.Sprintf("%02X      ", byte0)
	case 2:
		return fmt.Sprintf("%02X %02X   ", byte0, byte1)
	case 3:
		return fmt.Sprintf("%02X %02X %02X", byte0, byte1, byte2)
	}
	panic("Length must be 1, 2, or 3")
}

// addrString returns the address part of a 6502 assembly language
// instruction.
func addrString(pc uint16, byte1, byte2 byte, length int, mode opcodes.AddressingMode) string {
	addr16 := uint16(byte1) + uint16(byte2)<<8
	addrRel := uint16(int32(pc+2) + int32(int8(byte1)))
	switch mode {
	case opcodes.MODE_IMPLIED:
		return "       "
	case opcodes.MODE_ABSOLUTE:
		return fmt.Sprintf("$%04X  ", addr16)
	case opcodes.MODE_INDIRECT:
		return fmt.Sprintf("($%04X)", addr16)
	case opcodes.MODE_RELATIVE:
		return fmt.Sprintf("$%04X  ", addrRel)
	case opcodes.MODE_IMMEDIATE:
		return fmt.Sprintf("#$%02X   ", byte1)
	case opcodes.MODE_ABS_X:
		return fmt.Sprintf("$%04X,X", addr16)
	case opcodes.MODE_ABS_Y:
		return fmt.Sprintf("$%04X,Y", addr16)
	case opcodes.MODE_ZP:
		return fmt.Sprintf("$%02X    ", byte1)
	case opcodes.MODE_ZP_X:
		return fmt.Sprintf("$%02X,X  ", byte1)
	case opcodes.MODE_ZP_Y:
		return fmt.Sprintf("$%02X,Y  ", byte1)
	case opcodes.MODE_INDIRECT_Y:
		return fmt.Sprintf("($%02X),Y", byte1)
	case opcodes.MODE_INDIRECT_X:
		return fmt.Sprintf("($%02X,X)", byte1)
	case opcodes.MODE_A:
		return "       "
	}
	panic(fmt.Sprintf("Unknown op mode: %d", mode))
}

// Disasm disassembles a single (up to three byte) 6502
// instruction. It returns the formatted bytes, the formatted
// instruction and address, and the length. If it cannot find the
// instruction, it returns a 1-byte "???" instruction.
func Disasm(pc uint16, byte0, byte1, byte2 byte) (string, string, int) {
	op, ok := opcodes.Opcodes[byte0]
	if !ok {
		op = opcodes.NoOp
	}
	length := opcodes.ModeLengths[op.Mode]
	bytes := bytesString(byte0, byte1, byte2, length)
	addr := addrString(pc, byte1, byte2, length, op.Mode)
	return bytes, op.Name + " " + addr, length
}

// DisasmBlock disassembles an entire block, writing out the disassembly.
func DisasmBlock(block []byte, startAddr uint16, w io.Writer) {
	l := len(block)
	for i := 0; i < l; i++ {
		byte0 := block[i]
		byte1 := byte(0xFF)
		byte2 := byte(0xFF)
		if i+1 < l {
			byte1 = block[i+1]
		}
		if i+2 < l {
			byte2 = block[i+2]
		}
		addr := uint16(i) + startAddr
		bytes, op, length := Disasm(addr, byte0, byte1, byte2)
		fmt.Fprintf(w, "$%04X: %s %s\n", addr, bytes, op)
		i += length - 1
	}
}
