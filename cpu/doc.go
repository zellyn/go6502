/*

Package cpu provides routines for emulating a 6502 or 65C02. It also
provides data about opcodes that is used by the asm package to
(dis)assemble 6502 assembly language.

BUG(zellyn): 6502 should do invalid reads when doing indexed addressing across page boundaries. See http://en.wikipedia.org/wiki/MOS_Technology_6502#Bugs_and_quirks.
BUG(zellyn): rmw instructions should write old data back on 6502, read twice on 65C02. See http://en.wikipedia.org/wiki/MOS_Technology_6502#Bugs_and_quirks.
BUG(zellyn): implement interrupts, and 6502/65C02 decimal-mode-clearing and BRK-skipping quirks.
*/
package cpu
