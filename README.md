go6502
======

A 6502 emulator, simulator, and assembler(s), written in Go.

[![Build Status](https://travis-ci.org/zellyn/go6502.svg?branch=master)](https://travis-ci.org/zellyn/go6502)

This repository should probably be split up. It contains:

## cpu

The actual 6502 CPU emulation.

TODOs:
- [ ] Implement 65C02 variant
- [ ] Implement undocumented instructions
- [ ] Profile and speed up

## visual

A go transliteration of an old version of
https://github.com/mist64/perfect6502, the gate-level simulation of
the 6502.

TODOs:
- [ ] Incorporate recent speedups/simplifications made to perfect6502
- [ ] Profile and speed up
- [ ] Write a ridiculous one-goroutine-per-transistor simulation

## asm

A 6502 assembler, more-or-less compatible with several flavors of
oldschool (and soon, modern) assemblers:

Oldschool:
- [SCMA](http://www.txbobsc.com/scsc/scassembler/SCMacroAssembler20.html)
- [Merlin](https://en.wikipedia.org/wiki/Merlin_(assembler))
- "Redbook" (A and B) the flavor used in some Apple source listings.

Modern:
- (in-progress) [as65](http://www.kingswood-consulting.co.uk/assemblers/)
- (todo) [acme](https://sourceforge.net/projects/acme-crossass/)
