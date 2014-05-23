package redbook

import (
	"github.com/zellyn/go6502/asm/expr"
	"github.com/zellyn/go6502/asm/flavors/oldschool"
	"github.com/zellyn/go6502/asm/inst"
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

	a.Directives = map[string]oldschool.DirectiveInfo{
		".IN":   {inst.TypeInclude, a.ParseInclude},
		"ORG":   {inst.TypeOrg, a.ParseAddress},
		"OBJ":   {inst.TypeNone, nil},
		".TF":   {inst.TypeNone, nil},
		".EN":   {inst.TypeEnd, a.ParseNoArgDir},
		"EQU":   {inst.TypeEqu, a.ParseEquate},
		".DA":   {inst.TypeData, a.ParseData},
		"DFB":   {inst.TypeDataBytes, a.ParseData},
		"DW":    {inst.TypeDataWords, a.ParseData},
		".HS":   {inst.TypeData, a.ParseHexString},
		".AS":   {inst.TypeData, a.ParseAscii},
		".AT":   {inst.TypeData, a.ParseAscii},
		".BS":   {inst.TypeBlock, a.ParseBlockStorage},
		".LIST": {inst.TypeNone, nil},
		".PG":   {inst.TypeNone, nil},
		".DO":   {inst.TypeIfdef, a.ParseDo},
		".ELSE": {inst.TypeIfdefElse, a.ParseNoArgDir},
		".FIN":  {inst.TypeIfdefEnd, a.ParseNoArgDir},
		".MA":   {inst.TypeMacroStart, a.ParseMacroStart},
		".EM":   {inst.TypeMacroEnd, a.ParseNoArgDir},
		".US":   {inst.TypeNone, a.ParseNotImplemented},
		"PAGE":  {inst.TypeNone, nil}, // New page
		"LST":   {inst.TypeNone, nil}, // Listing on/off
		"SBTL":  {inst.TypeNone, nil}, // Subtitle
		"SKP":   {inst.TypeNone, nil}, // Skip lines
		"REP":   {inst.TypeNone, nil}, // Repeat character
		"CHR":   {inst.TypeNone, nil}, // Set repeated character
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

	return a
}
