package asm

import (
	"fmt"

	"github.com/zellyn/go6502/asm/context"
	"github.com/zellyn/go6502/asm/inst"
	"github.com/zellyn/go6502/asm/lines"
)

type Flavor interface {
	ParseInstr(Line lines.Line) (inst.I, error)
	DefaultOrigin() (uint16, error)
	context.Context
}

type Assembler struct {
	Flavor    Flavor
	Opener    lines.Opener
	Insts     []*inst.I
	LastLabel string
}

func NewAssembler(flavor Flavor, opener lines.Opener) *Assembler {
	return &Assembler{Flavor: flavor, Opener: opener}
}

// Load loads a new assembler file, deleting any previous data.
func (a *Assembler) Load(filename string) error {
	context := lines.Context{Filename: filename}
	ls, err := lines.NewFileLineSource(filename, context, a.Opener)
	if err != nil {
		return err
	}
	lineSources := []lines.LineSource{ls}
	ifdefs := []bool{}
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
		if in.Label != "" && in.Label[0] != '.' {
			a.Flavor.SetLastLabel(in.Label)
		}
		in.FixLabels(a.Flavor.GetLastLabel())
		if err != nil {
			return err
		}

		if _, err := a.passInst(&in, false, false); err != nil {
			return err
		}

		switch in.Type {
		case inst.TypeUnknown:
			return fmt.Errorf("Unknown instruction: %s", line)
		case inst.TypeMacroStart:
			return fmt.Errorf("Macro start not (yet) implemented: %s", line)
		case inst.TypeMacroCall:
			return fmt.Errorf("Macro call not (yet) implemented: %s", line)
		case inst.TypeIfdef:
			if len(in.Exprs) == 0 {
				panic(fmt.Sprintf("Ifdef got parsed with no expression: %s", line))
			}
			val, err := in.Exprs[0].Eval(a.Flavor)
			if err != nil {
				return fmt.Errorf("Cannot eval ifdef condition: %v", err)
			}
			ifdefs = append([]bool{val != 0}, ifdefs...)

		case inst.TypeIfdefElse:
			if len(ifdefs) == 0 {
				return fmt.Errorf("Ifdef else branch encountered outside ifdef: %s", line)
			}
			ifdefs[0] = !ifdefs[0]
		case inst.TypeIfdefEnd:
			if len(ifdefs) == 0 {
				return fmt.Errorf("Ifdef end encountered outside ifdef: %s", line)
			}
			ifdefs = ifdefs[1:]
		case inst.TypeInclude:
			subContext := lines.Context{Filename: in.TextArg, Parent: in.Line}
			subLs, err := lines.NewFileLineSource(in.TextArg, subContext, a.Opener)
			if err != nil {
				return fmt.Errorf("error including file: %v", err)
			}
			lineSources = append([]lines.LineSource{subLs}, lineSources...)
			continue // no need to append
		case inst.TypeTarget:
			return fmt.Errorf("Target not (yet) implemented: %s", line)
		case inst.TypeSegment:
			return fmt.Errorf("Segment not (yet) implemented: %s", line)
		case inst.TypeEnd:
			return nil
		default:
		}
		a.Insts = append(a.Insts, &in)
	}
	return nil
}

// Clear out stuff that may be hanging around from the previous pass, set origin to default, etc.
func (a *Assembler) initPass() {
	a.Flavor.SetLastLabel("") // No last label (yet)
	a.Flavor.RemoveChanged()  // Remove any variables whose value ever changed.
	if org, err := a.Flavor.DefaultOrigin(); err == nil {
		a.Flavor.SetAddr(org)
	} else {
		a.Flavor.ClearAddr()
	}
}

// passInst performs a pass on a single instruction. Depending on
// whether the instruction width can be determined, it updates or
// clears the current address. If setWidth is true, it forces the
// instruction to decide its final width. If final is true, and the
// instruction cannot be finalized, it returns an error.
func (a *Assembler) passInst(in *inst.I, setWidth, final bool) (isFinal bool, err error) {
	fmt.Printf("PLUGH: in.Compute(a.Flavor, true, true) on %s\n", in)
	isFinal, err = in.Compute(a.Flavor, setWidth, final)
	fmt.Printf("PLUGH: isFinal=%v, in.Final=%v, in.WidthKnown=%v, in.MinWidth=%v\n", isFinal, in.Final, in.WidthKnown, in.MinWidth)
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
		a.Flavor.ClearAddr()
	}

	return isFinal, nil
}

// Pass performs an assembly pass. If setWidth is true, it causes all
// instructions to set their final width. If final is true, it returns
// an error for any instruction that cannot be finalized.
func (a *Assembler) Pass(setWidth, final bool) (isFinal bool, err error) {
	fmt.Printf("PLUGH: Pass(%v, %v): %d instructions\n", setWidth, final, len(a.Insts))
	setWidth = setWidth || final // final ⊢ setWidth

	a.initPass()

	isFinal = true
	for _, in := range a.Insts {
		instFinal, err := a.passInst(in, setWidth, final)
		if err != nil {
			return false, err
		}
		if final && !instFinal {
			return false, fmt.Errorf("Cannot finalize instruction: %s", in)
		}
		fmt.Printf("PLUGH: instFinal=%v, in.Final=%v, in.WidthKnown=%v, in.MinWidth=%v\n", instFinal, in.Final, in.WidthKnown, in.MinWidth)
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
			return []byte{}, fmt.Errorf("cannot finalize value: %s", in)
		}
		result = append(result, in.Data...)
	}
	return result, nil
}

func (a *Assembler) Reset() {
	a.Insts = nil
	a.LastLabel = ""
}
