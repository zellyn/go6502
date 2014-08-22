package merlin

import (
	"fmt"
	"strings"

	"github.com/zellyn/go6502/asm/context"
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
	m.Name = "merlin"
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
	m.LocalMacroLabelsVal = true
	m.DefaultOriginVal = 0x8000

	m.Directives = map[string]oldschool.DirectiveInfo{
		"ORG":    {inst.TypeOrg, m.ParseAddress, 0},
		"OBJ":    {inst.TypeNone, nil, 0},
		"ENDASM": {inst.TypeEnd, m.ParseNoArgDir, 0},
		"=":      {inst.TypeEqu, m.ParseEquate, inst.VarEquNormal},
		"HEX":    {inst.TypeData, m.ParseHexString, inst.VarBytes},
		"DFB":    {inst.TypeData, m.ParseData, inst.VarBytes},
		"DB":     {inst.TypeData, m.ParseData, inst.VarBytes},
		"DA":     {inst.TypeData, m.ParseData, inst.VarWordsLe},
		"DDB":    {inst.TypeData, m.ParseData, inst.VarWordsBe},
		"ASC":    {inst.TypeData, m.ParseAscii, inst.VarAscii},
		"DCI":    {inst.TypeData, m.ParseAscii, inst.VarAsciiFlip},
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

	m.InitContextFunc = func(ctx context.Context) {
		ctx.SetOnOffDefaults(map[string]bool{
			"LST":   true,  // Display listing: not used
			"XC":    false, // Extended commands: not implemented yet
			"EXP":   false, // How to print macro calls
			"LSTDO": false, // List conditional code?
			"TR":    false, // truncate listing to 3 bytes?
			"CYC":   false, // print cycle times?
		})
	}

	m.SetAsciiVariation = func(ctx context.Context, in *inst.I, lp *lines.Parse) {
		if in.Command != "ASC" && in.Command != "DCI" {
			panic(fmt.Sprintf("Unimplemented/unknown ascii directive: '%s'", in.Command))
		}
		invert := lp.Peek() < '\''
		invertLast := in.Command == "DCI"
		switch {
		case !invert && !invertLast:
			in.Var = inst.VarAscii
		case !invert && invertLast:
			in.Var = inst.VarAsciiFlip
		case invert && !invertLast:
			in.Var = inst.VarAsciiHi
		case invert && invertLast:
			in.Var = inst.VarAsciiHiFlip
		}
	}

	// ParseMacroCall parses a macro call. We expect in.Command to hold
	// the "command column" value, which caused isMacroCall to return
	// true, and the lp to be pointing to the following character
	// (probably whitespace).
	m.ParseMacroCall = func(ctx context.Context, in inst.I, lp *lines.Parse) (inst.I, bool, error) {
		if in.Command == "" {
			return in, false, nil
		}
		byName := ctx.HasMacroName(in.Command)

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
			if !ctx.HasMacroName(in.Command) {
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

	m.FixLabel = func(ctx context.Context, label string) (string, error) {
		_, macroCount, locals := ctx.GetMacroCall()
		switch {
		case label == "":
			return label, nil
		case label[0] == ':':
			if last := ctx.LastLabel(); last == "" {
				return "", fmt.Errorf("sublabel '%s' without previous label", label)
			} else {
				return fmt.Sprintf("%s/%s", last, label), nil
			}
		case locals[label]:
			return fmt.Sprintf("%s{%d}", label, macroCount), nil

		}
		return label, nil
	}

	m.IsNewParentLabel = func(label string) bool {
		return label != "" && label[0] != ':'
	}

	return m
}

func (m *Merlin) ParseInclude(ctx context.Context, in inst.I, lp *lines.Parse) (inst.I, error) {
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
