package merlin

import (
	"errors"

	"github.com/zellyn/go6502/asm/context"
	"github.com/zellyn/go6502/asm/inst"
	"github.com/zellyn/go6502/asm/lines"
)

// Merlin implements the Merlin-compatible assembler flavor.
// See http://en.wikipedia.org/wiki/Merlin_(assembler) and
// http://www.apple-iigs.info/doc/fichiers/merlin816.pdf‎

type Merlin struct {
	context.SimpleContext
}

func New() *Merlin {
	return &Merlin{}
}

// Parse an entire instruction, or return an appropriate error.
func (a *Merlin) ParseInstr(line lines.Line) (inst.I, error) {
	return inst.I{}, nil
}

func (a *Merlin) Zero() (uint16, error) {
	return 0, errors.New("Division by zero.")
}

func (a *Merlin) DefaultOrigin() (uint16, error) {
	return 0x8000, nil
}

func (a *Merlin) SetWidthsOnFirstPass() bool {
	panic("don't know yet")
}

func (a *Merlin) ReplaceMacroArgs(line string, args []string, kwargs map[string]string) (string, error) {
	panic("Merlin.ReplaceMacroArgs not implemented yet.")
}
