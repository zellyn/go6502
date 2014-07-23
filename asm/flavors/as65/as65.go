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
	context.LabelerBase
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

func (a *AS65) SetWidthsOnFirstPass() bool {
	return false
}

func (a *AS65) ReplaceMacroArgs(line string, args []string, kwargs map[string]string) (string, error) {
	panic("AS65.ReplaceMacroArgs not implemented yet.")
}

func (a *AS65) IsNewParentLabel(label string) bool {
	return label != "" && label[0] != '.'
}

func (a *AS65) FixLabel(label string, macroCall int, locals map[string]bool) (string, error) {
	panic("AS65.FixLabel not implemented yet.")
}

func (a *AS65) LocalMacroLabels() bool {
	return false
}
