package as65

import (
	"errors"

	"github.com/zellyn/go6502/asm/context"
	"github.com/zellyn/go6502/asm/flavors"
	"github.com/zellyn/go6502/asm/inst"
	"github.com/zellyn/go6502/asm/lines"
)

// AS65 implements the AS65-compatible assembler flavor.
// See http://www.kingswood-consulting.co.uk/assemblers/

type AS65 struct{}

func New() *AS65 {
	return &AS65{}
}

// Parse an entire instruction, or return an appropriate error.
func (a *AS65) ParseInstr(ctx context.Context, line lines.Line, mode flavors.ParseMode) (inst.I, error) {
	return inst.I{}, nil
}

func (a *AS65) Zero() (uint16, error) {
	return 0, errors.New("Division by zero.")
}

func (a *AS65) DefaultOrigin() uint16 {
	return 0
}

func (a *AS65) ReplaceMacroArgs(line string, args []string, kwargs map[string]string) (string, error) {
	panic("AS65.ReplaceMacroArgs not implemented yet.")
}

func (a *AS65) IsNewParentLabel(label string) bool {
	return label != "" && label[0] != '.'
}

func (a *AS65) LocalMacroLabels() bool {
	return false
}

func (a *AS65) String() string {
	return "as65"
}

func (a *AS65) InitContext(ctx context.Context) {
}
