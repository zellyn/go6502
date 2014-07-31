package asm

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/zellyn/go6502/asm/flavors"
	"github.com/zellyn/go6502/asm/inst"
	"github.com/zellyn/go6502/asm/lines"
	"github.com/zellyn/go6502/asm/macros"
	"github.com/zellyn/go6502/asm/membuf"
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
func (a *Assembler) Load(filename string, prefix int) error {
	a.initPass()
	context := lines.Context{Filename: filename}
	ls, err := lines.NewFileLineSource(filename, context, a.Opener, prefix)
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
		// if line.Parse != nil {
		// 	fmt.Fprintf("PLUGH: %s\n", line.Text())
		// }
		if done {
			lineSources = lineSources[1:]
			continue
		}
		in, parseErr := a.Flavor.ParseInstr(line)
		if len(ifdefs) > 0 && !ifdefs[0] && in.Type != inst.TypeIfdefElse && in.Type != inst.TypeIfdefEnd {
			// we're in an inactive ifdef branch
			continue
		}

		if err != nil {
			return err
		}

		if err := in.FixLabels(a.Flavor); err != nil {
			return err
		}

		if _, err := a.passInst(&in, a.Flavor.SetWidthsOnFirstPass(), false); err != nil {
			return err
		}

		switch in.Type {
		case inst.TypeUnknown:
			if parseErr != nil {
				return parseErr
			}
			return line.Errorf("unknown instruction: %s", line.Parse.Text())
		case inst.TypeMacroStart:

			if err := a.readMacro(in, lineSources[0]); err != nil {
				return err
			}
			continue // no need to append
		case inst.TypeMacroCall:
			macroCall++
			m, ok := a.Macros[in.Command]
			if !ok {
				return in.Errorf(`unknown macro: "%s"`, in.Command)
			}
			subLs, err := m.LineSource(a.Flavor, in, macroCall, prefix)
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
			filename = filepath.Join(filepath.Dir(in.Line.Context.Filename), in.TextArg)
			subContext := lines.Context{Filename: filename, Parent: in.Line}
			subLs, err := lines.NewFileLineSource(filename, subContext, a.Opener, prefix)
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

func (a *Assembler) Assemble(filename string) error {
	return a.AssembleWithPrefix(filename, 0)
}

func (a *Assembler) AssembleWithPrefix(filename string, prefix int) error {
	a.Reset()
	if err := a.Load(filename, prefix); err != nil {
		return err
	}

	// Setwidth pass if necessary.
	if !a.Flavor.SetWidthsOnFirstPass() {
		if _, err := a.Pass(true, false); err != nil {
			return err
		}
	}

	// Final pass.
	if _, err := a.Pass(true, true); err != nil {
		return err
	}
	return nil
}

func (a *Assembler) readMacro(in inst.I, ls lines.LineSource) error {
	m := macros.M{
		Name: in.TextArg,
		Args: in.MacroArgs,
	}
	if a.Flavor.LocalMacroLabels() {
		m.Locals = make(map[string]bool)
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
		if a.Flavor.LocalMacroLabels() && in2.Label != "" {
			m.Locals[in2.Label] = true
		}
		if err == nil && in2.Type == inst.TypeMacroEnd {
			m.Lines = append(m.Lines, line.Parse.Text())
			a.Macros[m.Name] = m
			a.Flavor.AddMacroName(m.Name)
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

	// Update address
	if a.Flavor.AddrKnown() {
		addr, _ := a.Flavor.GetAddr()
		in.Addr = addr
		in.AddrKnown = true

		if in.WidthKnown {
			a.Flavor.SetAddr(addr + in.MinWidth)
		} else {
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

// RawBytes returns the raw bytes, sequentially in the order of the
// lines of the file. Intended for testing, the return value gives no
// indication of address changes.
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

func (a *Assembler) Membuf() (*membuf.Membuf, error) {
	m := &membuf.Membuf{}
	for _, in := range a.Insts {
		if !in.Final {
			return nil, in.Errorf("cannot finalize value: %s", in)
		}
		if !in.AddrKnown {
			return nil, in.Errorf("address unknown: %s", in)
		}
		if in.MinWidth > 0 {
			m.Write(int(in.Addr), in.Data)
		}
	}
	return m, nil
}

func (a *Assembler) GenerateListing(w io.Writer, width int) error {
	for _, in := range a.Insts {
		if !in.Final {
			return in.Errorf("cannot finalize value: %s", in)
		}
		if !in.AddrKnown {
			return in.Errorf("address unknown: %s", in)
		}

		for i := 0; i < len(in.Data) || i < width; i++ {
			if i%width == 0 {
				s := fmt.Sprintf("%04x:", int(in.Addr)+i)
				if i > 0 {
					s = "\n" + s
				}
				if _, err := fmt.Fprint(w, s); err != nil {
					return err
				}
			}

			s := "   "
			if i < len(in.Data) {
				s = fmt.Sprintf(" %02x", in.Data[i])
			}
			if _, err := fmt.Fprint(w, s); err != nil {
				return err
			}

			if i == width-1 {
				if _, err := fmt.Fprint(w, "    "+in.Line.Text()); err != nil {
					return err
				}
			}

		}
		if _, err := fmt.Fprint(w, "\n"); err != nil {
			return err
		}
	}
	return nil
}
