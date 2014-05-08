package macros

import (
	"github.com/zellyn/go6502/asm/flavors"
	"github.com/zellyn/go6502/asm/inst"
	"github.com/zellyn/go6502/asm/lines"
)

type M struct {
	Name  string
	Args  []string
	Lines []string
}

func (m M) LineSource(flavor flavors.F, in inst.I, macroCall int) (lines.LineSource, error) {
	var ls []string
	context := lines.Context{Filename: "macro:" + m.Name, Parent: in.Line, MacroCall: macroCall}
	for _, line := range m.Lines {
		// TODO(zellyn): implement named macro args
		subbed, err := flavor.ReplaceMacroArgs(line, in.MacroArgs, nil)
		if err != nil {
			return nil, in.Errorf("error in macro %s: %v", m.Name, err)
		}
		ls = append(ls, subbed)
	}
	return lines.NewSimpleLineSource(context, ls), nil
}
