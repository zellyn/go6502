package as65

import (
	"errors"

	"github.com/zellyn/go6502/asm/context"
	"github.com/zellyn/go6502/asm/inst"
	"github.com/zellyn/go6502/asm/lines"
)

// AS65 implements the AS65-compatible assembler flavor.
// See http://www.kingswood-consulting.co.uk/assemblers/

type AS65 struct {
	context.SimpleContext
}

func New() *AS65 {
	return &AS65{}
}

// Parse an entire instruction, or return an appropriate error.
func (a *AS65) ParseInstr(line lines.Line) (inst.I, error) {
	return inst.I{}, nil
}

func (a *AS65) Zero() (uint16, error) {
	return 0, errors.New("Division by zero.")
}

func (a *AS65) DefaultOrigin() (uint16, error) {
	return 0, nil
}
