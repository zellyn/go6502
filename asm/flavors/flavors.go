package flavors

import (
	"github.com/zellyn/go6502/asm/context"
	"github.com/zellyn/go6502/asm/inst"
	"github.com/zellyn/go6502/asm/lines"
)

type F interface {
	ParseInstr(Line lines.Line) (inst.I, error)
	DefaultOrigin() (uint16, error)
	SetWidthsOnFirstPass() bool
	ReplaceMacroArgs(line string, args []string, kwargs map[string]string) (string, error)
	context.Context
}
