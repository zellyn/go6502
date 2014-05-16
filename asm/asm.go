package asm

import (
	"fmt"

	"github.com/zellyn/go6502/asm/flavors"
	"github.com/zellyn/go6502/asm/inst"
	"github.com/zellyn/go6502/asm/lines"
	"github.com/zellyn/go6502/asm/macros"
)

type Assembler struct {
	Flavor    flavors.F
	Opener    lines.Opener
	Insts     []*inst.I
	LastLabel string
	Macros    map[string]macros.M
}

func NewAssembler(flavor flavors.F, opener lines.Opener) *Assembler {
	return &Assembler{
		Flavor: flavor,
		Opener: opener,
		Macros: make(map[string]macros.M),
	}
}

// Load loads a new assembler file, deleting any previous data.
func (a *Assembler) Load(filename string) error {
	a.initPass()
	context := lines.Context{Filename: filename}
	ls, err := lines.NewFileLineSource(filename, context, a.Opener)
	if err != nil {
		return err
	}
	lineSources := []lines.LineSource{ls}
	ifdefs := []bool{}
	macroCall := 0
	for len(lineSources) > 0 {
		line, done, err := lineSources[0].Next()
		if err != nil {
			return err
		}
		if done {
			lineSources = lineSources[1:]
			continue
		}
		in, err := a.Flavor.ParseInstr(line)
		if len(ifdefs) > 0 && !ifdefs[0] && in.Type != inst.TypeIfdefElse && in.Type != inst.TypeIfdefEnd {
			// we're in an inactive ifdef branch
			continue
		}
		if err := in.FixLabels(a.Flavor); err != nil {
			return err
		}

		if _, err := a.passInst(&in, a.Flavor.SetWidthsOnFirstPass(), false); err != nil {
			return err
		}

		switch in.Type {
		case inst.TypeUnknown:
			return line.Errorf("unknown instruction: %s", line.Parse.Text())
		case inst.TypeMacroStart:
			if err := a.readMacro(in, lineSources[0]); err != nil {
				return err
			}
			continue // no need to append
			return in.Errorf("macro start not (yet) implemented: %s", line)
		case inst.TypeMacroCall:
			macroCall++
			m, ok := a.Macros[in.Command]
			if !ok {
				return in.Errorf(`unknown macro: "%s"`, in.Command)
			}
			subLs, err := m.LineSource(a.Flavor, in, macroCall)
			if err != nil {
				return in.Errorf(`error calling macro "%s": %v`, m.Name, err)
			}
			lineSources = append([]lines.LineSource{subLs}, lineSources...)
		case inst.TypeIfdef:
			if len(in.Exprs) == 0 {
				panic(fmt.Sprintf("Ifdef got parsed with no expression: %s", line))
			}
			val, err := in.Exprs[0].Eval(a.Flavor, in.Line)
			if err != nil {
				return in.Errorf("cannot eval ifdef condition: %v", err)
			}
			ifdefs = append([]bool{val != 0}, ifdefs...)

		case inst.TypeIfdefElse:
			if len(ifdefs) == 0 {
				return in.Errorf("ifdef else branch encountered outside ifdef: %s", line)
			}
			ifdefs[0] = !ifdefs[0]
		case inst.TypeIfdefEnd:
			if len(ifdefs) == 0 {
				return in.Errorf("ifdef end encountered outside ifdef: %s", line)
			}
			ifdefs = ifdefs[1:]
		case inst.TypeInclude:
			subContext := lines.Context{Filename: in.TextArg, Parent: in.Line}
			subLs, err := lines.NewFileLineSource(in.TextArg, subContext, a.Opener)
			if err != nil {
				return in.Errorf("error including file: %v", err)
			}
			lineSources = append([]lines.LineSource{subLs}, lineSources...)
		case inst.TypeTarget:
			return in.Errorf("target not (yet) implemented: %s", line)
		case inst.TypeSegment:
			return in.Errorf("segment not (yet) implemented: %s", line)
		case inst.TypeEnd:
			return nil
		default:
		}
		a.Insts = append(a.Insts, &in)
	}
	return nil
}

func (a *Assembler) readMacro(in inst.I, ls lines.LineSource) error {
	m := macros.M{
		Name: in.TextArg,
		Args: in.MacroArgs,
	}
	for {
		line, done, err := ls.Next()
		if err != nil {
			return in.Errorf("error while reading macro %s: %v", m.Name)
		}
		if done {
			return in.Errorf("end of file while reading macro %s", m.Name)
		}
		in2, err := a.Flavor.ParseInstr(line)
		if err == nil && in2.Type == inst.TypeMacroEnd {
			a.Macros[m.Name] = m
			return nil
		}
		m.Lines = append(m.Lines, line.Parse.Text())
	}
}

// Clear out stuff that may be hanging around from the previous pass, set origin to default, etc.
func (a *Assembler) initPass() {
	a.Flavor.SetLastLabel("") // No last label (yet)
	a.Flavor.RemoveChanged()  // Remove any variables whose value ever changed.
	if org, err := a.Flavor.DefaultOrigin(); err == nil {
		a.Flavor.SetAddr(org)
	} else {
		a.Flavor.ClearAddr("beginning of assembly")
	}
}

// passInst performs a pass on a single instruction. Depending on
// whether the instruction width can be determined, it updates or
// clears the current address. If setWidth is true, it forces the
// instruction to decide its final width. If final is true, and the
// instruction cannot be finalized, it returns an error.
func (a *Assembler) passInst(in *inst.I, setWidth, final bool) (isFinal bool, err error) {
	// fmt.Printf("PLUGH: in.Compute(a.Flavor, true, true) on %s\n", in)
	isFinal, err = in.Compute(a.Flavor, setWidth, final)
	// fmt.Printf("PLUGH: isFinal=%v, in.Final=%v, in.WidthKnown=%v, in.MinWidth=%v\n", isFinal, in.Final, in.WidthKnown, in.MinWidth)
	if err != nil {
		return false, err
	}

	if in.WidthKnown && in.MinWidth != in.MaxWidth {
		panic(fmt.Sprintf("inst.I %s: WidthKnown=true, but MinWidth=%d, MaxWidth=%d", in, in.MinWidth, in.MaxWidth))
	}

	if in.WidthKnown && a.Flavor.AddrKnown() {
		addr, _ := a.Flavor.GetAddr()
		a.Flavor.SetAddr(addr + in.MinWidth)
	} else {
		if a.Flavor.AddrKnown() {
			a.Flavor.ClearAddr(in.Sprintf("lost known address"))
		}
	}

	return isFinal, nil
}

// Pass performs an assembly pass. If setWidth is true, it causes all
// instructions to set their final width. If final is true, it returns
// an error for any instruction that cannot be finalized.
func (a *Assembler) Pass(setWidth, final bool) (isFinal bool, err error) {
	// fmt.Printf("PLUGH: Pass(%v, %v): %d instructions\n", setWidth, final, len(a.Insts))
	setWidth = setWidth || final // final ⊢ setWidth

	a.initPass()

	isFinal = true
	for _, in := range a.Insts {
		instFinal, err := a.passInst(in, setWidth, final)
		if err != nil {
			return false, err
		}
		if final && !instFinal {
			return false, in.Errorf("cannot finalize instruction: %s", in)
		}
		// fmt.Printf("PLUGH: instFinal=%v, in.Final=%v, in.WidthKnown=%v, in.MinWidth=%v\n", instFinal, in.Final, in.WidthKnown, in.MinWidth)
		isFinal = isFinal && instFinal
	}

	return isFinal, nil
}

// RawBytes returns the raw bytes, sequentially. Intended for testing,
// the return value gives no indication of address changes.
func (a *Assembler) RawBytes() ([]byte, error) {
	result := []byte{}
	for _, in := range a.Insts {
		if !in.Final {
			return []byte{}, in.Errorf("cannot finalize value: %s", in)
		}
		result = append(result, in.Data...)
	}
	return result, nil
}

func (a *Assembler) Reset() {
	a.Insts = nil
	a.LastLabel = ""
}
