package merlin

import (
	"errors"
	"fmt"
	"strings"

	"github.com/zellyn/go6502/asm/expr"
	"github.com/zellyn/go6502/asm/flavors/oldschool"
	"github.com/zellyn/go6502/asm/inst"
	"github.com/zellyn/go6502/asm/lines"
)

// Merlin implements the Merlin-compatible assembler flavor.
// See http://en.wikipedia.org/wiki/Merlin_(assembler) and
// http://www.apple-iigs.info/doc/fichiers/merlin816.pdf‎
type Merlin struct {
	oldschool.Base
}

const whitespace = " \t"

func New() *Merlin {
	m := &Merlin{}
	m.LabelChars = oldschool.Letters + oldschool.Digits + ":"
	m.LabelColons = oldschool.ReqDisallowed
	m.ExplicitARegister = oldschool.ReqOptional
	m.StringEndOptional = false
	m.CommentChar = ';'
	m.BinaryChar = '%'
	m.LsbChars = "<"
	m.MsbChars = ">/"
	m.ImmediateChars = "#"
	m.HexCommas = oldschool.ReqOptional

	m.Directives = map[string]oldschool.DirectiveInfo{
		"ORG":    {inst.TypeOrg, m.ParseAddress, 0},
		"OBJ":    {inst.TypeNone, nil, 0},
		"ENDASM": {inst.TypeEnd, m.ParseNoArgDir, 0},
		"=":      {inst.TypeEqu, m.ParseEquate, inst.EquNormal},
		"HEX":    {inst.TypeData, m.ParseHexString, inst.DataBytes},
		"DFB":    {inst.TypeData, m.ParseData, inst.DataBytes},
		"DA":     {inst.TypeData, m.ParseData, inst.DataWordsLe},
		"DDB":    {inst.TypeData, m.ParseData, inst.DataWordsBe},
		"ASC":    {inst.TypeData, m.ParseAscii, inst.DataAscii},
		"DCI":    {inst.TypeData, m.ParseAscii, inst.DataAsciiFlip},
		".DO":    {inst.TypeIfdef, m.ParseDo, 0},
		".ELSE":  {inst.TypeIfdefElse, m.ParseNoArgDir, 0},
		".FIN":   {inst.TypeIfdefEnd, m.ParseNoArgDir, 0},
		".MA":    {inst.TypeMacroStart, m.ParseMacroStart, 0},
		".EM":    {inst.TypeMacroEnd, m.ParseNoArgDir, 0},
		".US":    {inst.TypeNone, m.ParseNotImplemented, 0},
		"PAGE":   {inst.TypeNone, nil, 0}, // New page
		"TTL":    {inst.TypeNone, nil, 0}, // Title
		"SAV":    {inst.TypeNone, nil, 0}, // Save
		"DSK":    {inst.TypeNone, nil, 0}, // Assemble to disk
		"PUT":    {inst.TypeInclude, m.ParseInclude, 0},
		"USE":    {inst.TypeInclude, m.ParseInclude, 0},
	}
	m.Operators = map[string]expr.Operator{
		"*": expr.OpMul,
		"/": expr.OpDiv,
		"+": expr.OpPlus,
		"-": expr.OpMinus,
		"<": expr.OpLt,
		">": expr.OpGt,
		"=": expr.OpEq,
	}

	m.SetOnOffDefaults(map[string]bool{
		"LST":   true,  // Display listing: not used
		"XC":    false, // Extended commands: not implemented yet
		"EXP":   false, // How to print macro calls
		"LSTDO": false, // List conditional code?
		"TR":    false, // truncate listing to 3 bytes?
		"CYC":   false, // print cycle times?
	})

	m.SetAsciiVariation = func(in *inst.I, lp *lines.Parse) {
		if in.Command != "ASC" && in.Command != "DCI" {
			panic(fmt.Sprintf("Unimplemented/unknown ascii directive: '%s'", in.Command))
		}
		invert := lp.Peek() < '\''
		invertLast := in.Command == "DCI"
		switch {
		case !invert && !invertLast:
			in.Var = inst.DataAscii
		case !invert && invertLast:
			in.Var = inst.DataAsciiFlip
		case invert && !invertLast:
			in.Var = inst.DataAsciiHi
		case invert && invertLast:
			in.Var = inst.DataAsciiHiFlip
		}
	}

	return m
}

func (m *Merlin) Zero() (uint16, error) {
	return 0, errors.New("Division by zero.")
}

func (m *Merlin) DefaultOrigin() (uint16, error) {
	return 0x8000, nil
}

func (m *Merlin) SetWidthsOnFirstPass() bool {
	// TODO(zellyn): figure this out
	return true
}

func (m *Merlin) ParseInclude(in inst.I, lp *lines.Parse) (inst.I, error) {
	lp.IgnoreRun(whitespace)
	lp.AcceptUntil(";")
	filename := strings.TrimSpace(lp.Emit())
	prefix := "T."
	if len(filename) > 0 && filename[0] < '@' {
		prefix = ""
		filename = strings.TrimSpace(filename[1:])
	}
	if filename == "" {
		return inst.I{}, in.Errorf("%s expects filename", in.Command)
	}
	in.TextArg = prefix + filename
	in.WidthKnown = true
	in.MinWidth = 0
	in.MaxWidth = 0
	in.Final = true
	return in, nil
}

func (m *Merlin) ReplaceMacroArgs(line string, args []string, kwargs map[string]string) (string, error) {
	panic("Merlin.ReplaceMacroArgs not implemented yet.")
}

func (m *Merlin) IsNewParentLabel(label string) bool {
	return label != "" && label[0] != ':'
}

func (m *Merlin) FixLabel(label string, macroCount int) (string, error) {
	switch {
	case label == "":
		return label, nil
	case label[0] == ':':
		if last := m.LastLabel(); last == "" {
			return "", fmt.Errorf("sublabel '%s' without previous label", label)
		} else {
			return fmt.Sprintf("%s/%s", last, label), nil
		}
	}
	return label, nil
}
