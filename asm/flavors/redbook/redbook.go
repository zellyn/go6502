package redbook

import (
	"fmt"

	"github.com/zellyn/go6502/asm/expr"
	"github.com/zellyn/go6502/asm/flavors/oldschool"
	"github.com/zellyn/go6502/asm/inst"
	"github.com/zellyn/go6502/asm/lines"
)

// RedBook implements a Redbook-listing-compatible-ish assembler flavor.
type RedBook struct {
	oldschool.Base
}

func New() *RedBook {
	a := &RedBook{}

	a.LabelChars = oldschool.Letters + oldschool.Digits + "."
	a.LabelColons = oldschool.ReqOptional
	a.ExplicitARegister = oldschool.ReqRequired
	a.StringEndOptional = true

	a.Directives = map[string]oldschool.DirectiveInfo{
		".IN":   {inst.TypeInclude, a.ParseInclude, 0},
		"ORG":   {inst.TypeOrg, a.ParseAddress, 0},
		"OBJ":   {inst.TypeNone, nil, 0},
		".TF":   {inst.TypeNone, nil, 0},
		".EN":   {inst.TypeEnd, a.ParseNoArgDir, 0},
		"EQU":   {inst.TypeEqu, a.ParseEquate, 0},
		"DFB":   {inst.TypeData, a.ParseData, inst.DataBytes},
		"DW":    {inst.TypeData, a.ParseData, inst.DataWordsLe},
		"DDB":   {inst.TypeData, a.ParseData, inst.DataWordsBe},
		"ASC":   {inst.TypeData, a.ParseAscii, inst.DataAscii},
		"DCI":   {inst.TypeData, a.ParseAscii, inst.DataAsciiFlip},
		".BS":   {inst.TypeBlock, a.ParseBlockStorage, 0},
		".LIST": {inst.TypeNone, nil, 0},
		".PG":   {inst.TypeNone, nil, 0},
		".DO":   {inst.TypeIfdef, a.ParseDo, 0},
		".ELSE": {inst.TypeIfdefElse, a.ParseNoArgDir, 0},
		".FIN":  {inst.TypeIfdefEnd, a.ParseNoArgDir, 0},
		".MA":   {inst.TypeMacroStart, a.ParseMacroStart, 0},
		".EM":   {inst.TypeMacroEnd, a.ParseNoArgDir, 0},
		".US":   {inst.TypeNone, a.ParseNotImplemented, 0},
		"PAGE":  {inst.TypeNone, nil, 0}, // New page
		"SBTL":  {inst.TypeNone, nil, 0}, // Subtitle
		"SKP":   {inst.TypeNone, nil, 0}, // Skip lines
		"REP":   {inst.TypeNone, nil, 0}, // Repeat character
		"CHR":   {inst.TypeNone, nil, 0}, // Set repeated character
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

	a.OnOff = map[string]bool{
		"MSB": true, // MSB defaults to true, as per manual
		"LST": true, // Display listing: not used
	}

	a.SetAsciiVariation = func(in *inst.I, lp *lines.Parse) {
		if in.Command == "ASC" {
			if a.Setting("MSB") {
				in.Var = inst.DataAsciiHi
			} else {
				in.Var = inst.DataAscii
			}
			return
		}
		if in.Command == "DCI" {
			in.Var = inst.DataAsciiFlip
		} else {
			panic(fmt.Sprintf("Unknown ascii directive: '%s'", in.Command))
		}
	}

	return a
}
