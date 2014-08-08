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
const macroNameChars = oldschool.Letters + oldschool.Digits + "_"

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
	m.CharChars = "'"
	m.InvCharChars = `"`
	m.MacroArgSep = ";"
	m.SuffixForWide = true

	m.Directives = map[string]oldschool.DirectiveInfo{
		"ORG":    {inst.TypeOrg, m.ParseAddress, 0},
		"OBJ":    {inst.TypeNone, nil, 0},
		"ENDASM": {inst.TypeEnd, m.ParseNoArgDir, 0},
		"=":      {inst.TypeEqu, m.ParseEquate, inst.EquNormal},
		"HEX":    {inst.TypeData, m.ParseHexString, inst.DataBytes},
		"DFB":    {inst.TypeData, m.ParseData, inst.DataBytes},
		"DB":     {inst.TypeData, m.ParseData, inst.DataBytes},
		"DA":     {inst.TypeData, m.ParseData, inst.DataWordsLe},
		"DDB":    {inst.TypeData, m.ParseData, inst.DataWordsBe},
		"ASC":    {inst.TypeData, m.ParseAscii, inst.DataAscii},
		"DCI":    {inst.TypeData, m.ParseAscii, inst.DataAsciiFlip},
		".DO":    {inst.TypeIfdef, m.ParseDo, 0},
		".ELSE":  {inst.TypeIfdefElse, m.ParseNoArgDir, 0},
		".FIN":   {inst.TypeIfdefEnd, m.ParseNoArgDir, 0},
		"MAC":    {inst.TypeMacroStart, m.MarkMacroStart, 0},
		"EOM":    {inst.TypeMacroEnd, m.ParseNoArgDir, 0},
		"<<<":    {inst.TypeMacroEnd, m.ParseNoArgDir, 0},
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
		".": expr.OpOr,
		"&": expr.OpAnd,
		"!": expr.OpXor,
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

	// ParseMacroCall parses a macro call. We expect in.Command to hold
	// the "command column" value, which caused isMacroCall to return
	// true, and the lp to be pointing to the following character
	// (probably whitespace).
	m.ParseMacroCall = func(in inst.I, lp *lines.Parse) (inst.I, bool, error) {
		if in.Command == "" {
			return in, false, nil
		}
		byName := m.HasMacroName(in.Command)

		// It's not a macro call.
		if in.Command != ">>>" && in.Command != "PMC" && !byName {
			return in, false, nil
		}

		in.Type = inst.TypeMacroCall
		lp.IgnoreRun(oldschool.Whitespace)
		if !byName {
			if !lp.AcceptRun(macroNameChars) {
				c := lp.Next()
				return in, true, in.Errorf("Expected macro name, got char '%c'", c)
			}
			in.Command = lp.Emit()
			if !m.HasMacroName(in.Command) {
				return in, true, in.Errorf("Unknown macro: '%s'", in.Command)
			}
			if !lp.Consume(" ./,-(") {
				c := lp.Next()
				if c == lines.Eol || c == ';' {
					return in, true, nil
				}
				return in, true, in.Errorf("Expected macro name/args separator [ ./,-(], got '%c'", c)

			}
		}
		for {
			s, err := m.ParseMacroArg(in, lp)
			if err != nil {
				return in, true, err
			}
			in.MacroArgs = append(in.MacroArgs, s)
			if !lp.Consume(";") {
				break
			}
		}

		return in, true, nil
	}

	return m
}

func (m *Merlin) Zero() (uint16, error) {
	return 0, errors.New("Division by zero.")
}

func (m *Merlin) DefaultOrigin() (uint16, error) {
	return 0x8000, nil
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
		return in, in.Errorf("%s expects filename", in.Command)
	}
	in.TextArg = prefix + filename
	in.WidthKnown = true
	in.Width = 0
	in.Final = true
	return in, nil
}

func (m *Merlin) IsNewParentLabel(label string) bool {
	return label != "" && label[0] != ':'
}

func (m *Merlin) FixLabel(label string, macroCount int, locals map[string]bool) (string, error) {
	switch {
	case label == "":
		return label, nil
	case label[0] == ':':
		if last := m.LastLabel(); last == "" {
			return "", fmt.Errorf("sublabel '%s' without previous label", label)
		} else {
			return fmt.Sprintf("%s/%s", last, label), nil
		}
	case locals[label]:
		return fmt.Sprintf("%s{%d}", label, macroCount), nil

	}
	return label, nil
}

func (m *Merlin) LocalMacroLabels() bool {
	return true
}
