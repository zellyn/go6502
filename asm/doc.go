/*
Package asm provides routines for assembling 6502 code. It currently
emulates the S-C Macro Assembler. The goal is to support (at least)
as65 and Merlin assembly files too.

Once those three (two ancient, one modern) are complete, adding
additional flavors should be straightforward.

TODO(zellyn): make errors return line and character position.
TODO(zellyn): scma requires .EQ and .BS to have known values. Is this universal?

*/
package asm
