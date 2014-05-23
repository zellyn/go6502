package scma

import (
	"github.com/zellyn/go6502/asm/expr"
	"github.com/zellyn/go6502/asm/flavors/oldschool"
	"github.com/zellyn/go6502/asm/inst"
)

// SCMA implements the S-C Macro Assembler-compatible assembler flavor.
// See http://www.txbobsc.com/scsc/ and http://stjarnhimlen.se/apple2/
type SCMA struct {
	oldschool.Base
}

func New() *SCMA {
	a := &SCMA{}
	a.LabelChars = oldschool.Letters + oldschool.Digits + ".:"
	a.LabelColons = oldschool.ReqDisallowed
	a.ExplicitARegister = oldschool.ReqDisallowed

	a.Directives = map[string]oldschool.DirectiveInfo{
		".IN":   {inst.TypeInclude, a.ParseInclude},
		".OR":   {inst.TypeOrg, a.ParseAddress},
		".TA":   {inst.TypeTarget, a.ParseNotImplemented},
		".TF":   {inst.TypeNone, nil},
		".EN":   {inst.TypeEnd, a.ParseNoArgDir},
		".EQ":   {inst.TypeEqu, a.ParseEquate},
		".DA":   {inst.TypeData, a.ParseData},
		".HS":   {inst.TypeData, a.ParseHexString},
		".AS":   {inst.TypeData, a.ParseAscii},
		".AT":   {inst.TypeData, a.ParseAscii},
		".BS":   {inst.TypeBlock, a.ParseBlockStorage},
		".TI":   {inst.TypeNone, nil},
		".LIST": {inst.TypeNone, nil},
		".PG":   {inst.TypeNone, nil},
		".DO":   {inst.TypeIfdef, a.ParseDo},
		".ELSE": {inst.TypeIfdefElse, a.ParseNoArgDir},
		".FIN":  {inst.TypeIfdefEnd, a.ParseNoArgDir},
		".MA":   {inst.TypeMacroStart, a.ParseMacroStart},
		".EM":   {inst.TypeMacroEnd, a.ParseNoArgDir},
		".US":   {inst.TypeNone, a.ParseNotImplemented},
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

func (a *SCMA) Zero() (uint16, error) {
	return uint16(0xffff), nil
}
