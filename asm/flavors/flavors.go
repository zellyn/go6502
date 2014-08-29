package flavors

import (
	"github.com/zellyn/go6502/asm/context"
	"github.com/zellyn/go6502/asm/inst"
	"github.com/zellyn/go6502/asm/lines"
)

type ParseMode int

const (
	ParseModeNormal ParseMode = iota
	ParseModeMacroSave
	ParseModeInactive
)

type F interface {
	ParseInstr(ctx context.Context, Line lines.Line, mode ParseMode) (inst.I, error)
	DefaultOrigin() uint16
	ReplaceMacroArgs(line string, args []string, kwargs map[string]string) (string, error)
	LocalMacroLabels() bool
	String() string
	InitContext(context.Context)
}
