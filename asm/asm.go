package asm

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/zellyn/go6502/asm/context"
	"github.com/zellyn/go6502/asm/flavors"
	"github.com/zellyn/go6502/asm/inst"
	"github.com/zellyn/go6502/asm/lines"
	"github.com/zellyn/go6502/asm/macros"
	"github.com/zellyn/go6502/asm/membuf"
)

type Assembler struct {
	Flavor flavors.F
	Opener lines.Opener
	Insts  []*inst.I
	Macros map[string]macros.M
	Ctx    *context.SimpleContext
}

func NewAssembler(flavor flavors.F, opener lines.Opener) *Assembler {
	ctx := &context.SimpleContext{}
	flavor.InitContext(ctx)
	return &Assembler{
		Flavor: flavor,
		Opener: opener,
		Macros: make(map[string]macros.M),
		Ctx:    ctx,
	}
}

type ifdef struct {
	active bool
	in     inst.I
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
	ifdefs := []ifdef{}
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
		inactive := len(ifdefs) > 0 && !ifdefs[0].active
		mode := flavors.ParseModeNormal
		if inactive {
			mode = flavors.ParseModeInactive
		}
		in, parseErr := a.Flavor.ParseInstr(a.Ctx, line, mode)
		if inactive && in.Type != inst.TypeIfdefElse && in.Type != inst.TypeIfdefEnd {
			// we're still in an inactive ifdef branch
			continue
		}
		if parseErr != nil {
			return parseErr
		}

		switch in.Type {
		case inst.TypeOrg:
			a.Ctx.SetAddr(in.Addr)
		case inst.TypeTarget:
			return in.Errorf("target not implemented yet")
		}

		// Update address
		addr := a.Ctx.GetAddr()
		in.Addr = addr
		a.Ctx.SetAddr(addr + in.Width)

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
			a.Ctx.PushMacroCall(m.Name, macroCall, m.Locals)
		case inst.TypeMacroEnd:
			// If we reached here, it's in a macro call, not a definition.
			if !a.Ctx.PopMacroCall() {
				return in.Errorf("unexpected end of macro")
			}
		case inst.TypeIfdef:
			if len(in.Exprs) == 0 {
				panic(fmt.Sprintf("Ifdef got parsed with no expression: %s", line))
			}
			val, err := in.Exprs[0].Eval(a.Ctx, in.Line)
			if err != nil {
				return in.Errorf("cannot eval ifdef condition: %v", err)
			}
			ifdefs = append([]ifdef{{val != 0, in}}, ifdefs...)

		case inst.TypeIfdefElse:
			if len(ifdefs) == 0 {
				return in.Errorf("ifdef else branch encountered outside ifdef: %s", line)
			}
			ifdefs[0].active = !ifdefs[0].active
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
	if len(ifdefs) > 0 {
		return ifdefs[0].in.Errorf("Ifdef not closed before end of file")
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

	// Final pass.
	if err := a.Pass2(); err != nil {
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
		in2, err := a.Flavor.ParseInstr(a.Ctx, line, flavors.ParseModeMacroSave)
		if a.Flavor.LocalMacroLabels() && in2.Label != "" {
			m.Locals[in2.Label] = true
		}
		m.Lines = append(m.Lines, line.Parse.Text())
		if err == nil && in2.Type == inst.TypeMacroEnd {
			a.Macros[m.Name] = m
			a.Ctx.AddMacroName(m.Name)
			return nil
		}
	}
}

// Clear out stuff that may be hanging around from the previous pass, set origin to default, etc.
func (a *Assembler) initPass() {
	a.Ctx.SetLastLabel("") // No last label (yet)
	a.Ctx.RemoveChanged()  // Remove any variables whose value ever changed.
	a.Ctx.SetAddr(a.Flavor.DefaultOrigin())
}

// Pass2 performs the second assembly pass. It returns an error for
// any instruction that cannot be finalized.
func (a *Assembler) Pass2() error {
	a.initPass()

	for _, in := range a.Insts {
		switch in.Type {
		case inst.TypeOrg:
			a.Ctx.SetAddr(in.Addr)
			continue
		case inst.TypeEqu:
			val, err := in.Exprs[0].Eval(a.Ctx, in.Line)
			if err != nil {
				return err
			}
			in.Value = val
			a.Ctx.Set(in.Label, val)
			continue
		}

		err := in.Compute(a.Ctx)
		if err != nil {
			return err
		}

		// Update address
		addr := a.Ctx.GetAddr()
		in.Addr = addr
		a.Ctx.SetAddr(addr + in.Width)
	}

	return nil
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
		data := in.Data
		for len(data) < int(in.Width) {
			data = append(data, 0x00)
		}
		result = append(result, data...)
	}
	return result, nil
}

func (a *Assembler) Reset() {
	a.Insts = nil
}

func (a *Assembler) Membuf() (*membuf.Membuf, error) {
	m := &membuf.Membuf{}
	for _, in := range a.Insts {
		if !in.Final {
			return nil, in.Errorf("cannot finalize value: %s", in)
		}
		if in.Width > 0 {
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
