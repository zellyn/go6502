package scma

import (
	"strings"

	"github.com/zellyn/go6502/asm/context"
	"github.com/zellyn/go6502/asm/expr"
	"github.com/zellyn/go6502/asm/flavors/oldschool"
	"github.com/zellyn/go6502/asm/inst"
	"github.com/zellyn/go6502/asm/lines"
)

// 40 spaces = comment column
const commentWhitespacePrefix = "                                        "

// SCMA implements the S-C Macro Assembler-compatible assembler flavor.
// See http://www.txbobsc.com/scsc/ and http://stjarnhimlen.se/apple2/
type SCMA struct {
	oldschool.Base
}

func New() *SCMA {
	a := &SCMA{}
	a.Name = "scma"
	a.LabelChars = oldschool.Letters + oldschool.Digits + ".:"
	a.LabelColons = oldschool.ReqDisallowed
	a.ExplicitARegister = oldschool.ReqDisallowed
	a.SpacesForComment = 2
	a.MsbChars = "/"
	a.ImmediateChars = "#"
	a.CharChars = "'"
	a.MacroArgSep = ","
	a.DefaultOriginVal = 0x0800
	divZeroVal := uint16(0xffff)
	a.DivZeroVal = &divZeroVal

	a.Directives = map[string]oldschool.DirectiveInfo{
		".IN":   {inst.TypeInclude, a.ParseInclude, 0},
		".OR":   {inst.TypeOrg, a.ParseOrg, 0},
		".TA":   {inst.TypeTarget, a.ParseNotImplemented, 0},
		".TF":   {inst.TypeNone, nil, 0},
		".EN":   {inst.TypeEnd, a.ParseNoArgDir, 0},
		".EQ":   {inst.TypeEqu, a.ParseEquate, 0},
		".DA":   {inst.TypeData, a.ParseData, inst.VarMixed},
		".HS":   {inst.TypeData, a.ParseHexString, inst.VarBytes},
		".AS":   {inst.TypeData, a.ParseAscii, inst.VarBytes},
		".AT":   {inst.TypeData, a.ParseAscii, inst.VarBytes},
		".BS":   {inst.TypeBlock, a.ParseBlockStorage, 0},
		".TI":   {inst.TypeNone, nil, 0},
		".LIST": {inst.TypeNone, nil, 0},
		".PG":   {inst.TypeNone, nil, 0},
		".DO":   {inst.TypeIfdef, a.ParseDo, 0},
		".ELSE": {inst.TypeIfdefElse, a.ParseNoArgDir, 0},
		".FIN":  {inst.TypeIfdefEnd, a.ParseNoArgDir, 0},
		".MA":   {inst.TypeMacroStart, a.ParseMacroStart, 0},
		".EM":   {inst.TypeMacroEnd, a.ParseNoArgDir, 0},
		".US":   {inst.TypeNone, a.ParseNotImplemented, 0},
	}

	a.EquateDirectives = map[string]bool{
		".EQ": true,
	}

	a.Operators = map[string]expr.Operator{
		"*": expr.OpMul,
		"/": expr.OpDiv,
		"+": expr.OpPlus,
		"-": expr.OpMinus,
		"<": expr.OpLt,
		">": expr.OpGt,
		"=": expr.OpEq,
	}

	a.ExtraCommenty = func(s string) bool {
		//Comment by virtue of long whitespace prefix
		return strings.HasPrefix(s, commentWhitespacePrefix)
	}

	a.SetAsciiVariation = func(ctx context.Context, in *inst.I, lp *lines.Parse) {
		// For S-C Assembler, leading "-" flips high bit
		invert := lp.Consume("-")
		invertLast := in.Command == ".AT"
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
	a.ParseMacroCall = func(ctx context.Context, in inst.I, lp *lines.Parse) (inst.I, bool, error) {
		if in.Command == "" || in.Command[0] != '>' {
			// not a macro call
			return in, false, nil
		}

		in.Type = inst.TypeMacroCall
		in.Command = in.Command[1:]

		lp.Consume(oldschool.Whitespace)

		for {
			s, err := a.ParseMacroArg(in, lp)
			if err != nil {
				return in, true, err
			}
			in.MacroArgs = append(in.MacroArgs, s)
			if !lp.Consume(",") {
				break
			}
		}

		return in, true, nil
	}

	a.FixLabel = a.DefaultFixLabel
	a.IsNewParentLabel = a.DefaultIsNewParentLabel

	return a
}
