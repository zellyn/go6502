/*
Package asm provides routines for decompiling 6502 assembly language.
*/
package asm

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"

	"github.com/zellyn/go6502/opcodes"
)

// Symbols are symbol tables used to convert addresses to names
type Symbols map[int]string

// name returns a symbol for an address. If lookback is 1 or greater,
// it can return strings like "FOO+1" for the addresses succeeding
// defined symbols.
func (s Symbols) name(addr int, lookback int) string {
	if n, ok := s[addr]; ok {
		return n
	}
	for i := 1; i <= lookback; i++ {
		if n, ok := s[addr-i]; ok {
			return fmt.Sprintf("%s+%d", n, i)
		}
	}
	return ""
}

// bytesString takes three bytes and a length, returning the formatted
// hex bytes for an instrction of the given length.
func bytesString(byte0, byte1, byte2 byte, length int) string {
	switch length {
	case 1:
		return fmt.Sprintf("%02X", byte0)
	case 2:
		return fmt.Sprintf("%02X %02X", byte0, byte1)
	case 3:
		return fmt.Sprintf("%02X %02X %02X", byte0, byte1, byte2)
	}
	panic("Length must be 1, 2, or 3")
}

func (s *Symbols) addr4(addr uint16, lookback int) string {
	if n := s.name(int(addr), lookback); n != "" {
		return n
	}
	return fmt.Sprintf("$%04X", addr)
}

func (s *Symbols) addr2(addr byte, lookback int) string {
	if n := s.name(int(addr), lookback); n != "" {
		return n
	}
	return fmt.Sprintf("$%02X", addr)
}

// addrString returns the address part of a 6502 assembly language
// instruction.
func addrString(pc uint16, byte1, byte2 byte, length int, mode opcodes.AddressingMode, s Symbols, lookback int) string {
	addr16 := uint16(byte1) + uint16(byte2)<<8
	addrRel := uint16(int32(pc+2) + int32(int8(byte1)))
	switch mode {
	case opcodes.MODE_IMPLIED, opcodes.MODE_A:
		return ""
	case opcodes.MODE_ABSOLUTE:
		return fmt.Sprintf("%s", s.addr4(addr16, lookback))
	case opcodes.MODE_INDIRECT:
		return fmt.Sprintf("(%s)", s.addr4(addr16, lookback))
	case opcodes.MODE_RELATIVE:
		return fmt.Sprintf("%s", s.addr4(addrRel, lookback))
	case opcodes.MODE_IMMEDIATE:
		return fmt.Sprintf("#$%02X", byte1)
	case opcodes.MODE_ABS_X:
		return fmt.Sprintf("%s,X", s.addr4(addr16, lookback))
	case opcodes.MODE_ABS_Y:
		return fmt.Sprintf("%s,Y", s.addr4(addr16, lookback))
	case opcodes.MODE_ZP:
		return fmt.Sprintf("%s", s.addr2(byte1, lookback))
	case opcodes.MODE_ZP_X:
		return fmt.Sprintf("%s,X", s.addr2(byte1, lookback))
	case opcodes.MODE_ZP_Y:
		return fmt.Sprintf("%s,Y", s.addr2(byte1, lookback))
	case opcodes.MODE_INDIRECT_Y:
		return fmt.Sprintf("(%s),Y", s.addr2(byte1, lookback))
	case opcodes.MODE_INDIRECT_X:
		return fmt.Sprintf("(%s,X)", s.addr2(byte1, lookback))
	}
	panic(fmt.Sprintf("Unknown op mode: %d", mode))
}

// Disasm disassembles a single (up to three byte) 6502
// instruction. It returns the formatted bytes, the formatted
// instruction and address, and the length. If it cannot find the
// instruction, it returns a 1-byte "???" instruction.
func Disasm(pc uint16, byte0, byte1, byte2 byte, s Symbols, lookback int) (string, string, int) {
	op, ok := opcodes.Opcodes[byte0]
	if !ok {
		op = opcodes.NoOp
	}
	length := opcodes.ModeLengths[op.Mode]
	bytes := bytesString(byte0, byte1, byte2, length)
	addr := addrString(pc, byte1, byte2, length, op.Mode, s, lookback)
	return bytes, op.Name + " " + addr, length
}

// DisasmBlock disassembles an entire block, writing out the disassembly.
func DisasmBlock(block []byte, startAddr uint16, w io.Writer, s Symbols, lookback int, printLabels bool) {
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
		bytes, op, length := Disasm(addr, byte0, byte1, byte2, s, lookback)
		if printLabels {
			label, ok := s[int(addr)]
			if ok {
				fmt.Fprintf(w, "$%04X:          %s:\n", addr, label)
			}
		}
		fmt.Fprintf(w, "$%04X: %-8s %s\n", addr, bytes, op)
		i += length - 1
	}
}

var symRe = regexp.MustCompile(`(?i)^([0-9a-z_]+) *(?:.eq|equ|epz|=) *\$([0-9a-f]+)\b`)

func ReadSymbols(filename string) (Symbols, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(file)
	s := make(Symbols)
	lineNum := 0
	for scanner.Scan() {
		line := scanner.Text()
		lineNum++
		groups := symRe.FindStringSubmatch(line)
		if groups == nil {
			continue
		}
		i, err := strconv.ParseInt(groups[2], 16, 32)
		if err != nil {
			return nil, fmt.Errorf("%d: %s: %s", lineNum, line, err)
		}
		if i > 0xffff {
			return nil, fmt.Errorf("%d: %s: Value %d out of range", lineNum, line, i)
		}
		s[int(i)] = groups[1]
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return s, nil
}
