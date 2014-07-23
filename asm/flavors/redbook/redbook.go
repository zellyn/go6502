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

func NewRedbookA() *RedBook {
	r := newRedbook()
	return r
}

func NewRedbookB() *RedBook {
	r := newRedbook()
	r.ExplicitARegister = oldschool.ReqRequired
	r.SpacesForComment = 3
	return r
}

func newRedbook() *RedBook {
	r := &RedBook{}

	r.LabelChars = oldschool.Letters + oldschool.Digits + "."
	r.LabelColons = oldschool.ReqOptional
	r.ExplicitARegister = oldschool.ReqRequired
	r.StringEndOptional = true
	r.CommentChar = ';'
	r.MsbChars = "/"
	r.ImmediateChars = "#"

	r.Directives = map[string]oldschool.DirectiveInfo{
		"ORG":    {inst.TypeOrg, r.ParseAddress, 0},
		"OBJ":    {inst.TypeNone, nil, 0},
		"ENDASM": {inst.TypeEnd, r.ParseNoArgDir, 0},
		"EQU":    {inst.TypeEqu, r.ParseEquate, inst.EquNormal},
		"EPZ":    {inst.TypeEqu, r.ParseEquate, inst.EquPageZero},
		"DFB":    {inst.TypeData, r.ParseData, inst.DataBytes},
		"DW":     {inst.TypeData, r.ParseData, inst.DataWordsLe},
		"DDB":    {inst.TypeData, r.ParseData, inst.DataWordsBe},
		"ASC":    {inst.TypeData, r.ParseAscii, inst.DataAscii},
		"DCI":    {inst.TypeData, r.ParseAscii, inst.DataAsciiFlip},
		".DO":    {inst.TypeIfdef, r.ParseDo, 0},
		".ELSE":  {inst.TypeIfdefElse, r.ParseNoArgDir, 0},
		".FIN":   {inst.TypeIfdefEnd, r.ParseNoArgDir, 0},
		".MA":    {inst.TypeMacroStart, r.ParseMacroStart, 0},
		".EM":    {inst.TypeMacroEnd, r.ParseNoArgDir, 0},
		".US":    {inst.TypeNone, r.ParseNotImplemented, 0},
		"PAGE":   {inst.TypeNone, nil, 0}, // New page
		"TITLE":  {inst.TypeNone, nil, 0}, // Title
		"SBTL":   {inst.TypeNone, nil, 0}, // Subtitle
		"SKP":    {inst.TypeNone, nil, 0}, // Skip lines
		"REP":    {inst.TypeNone, nil, 0}, // Repeat character
		"CHR":    {inst.TypeNone, nil, 0}, // Set repeated character
	}
	r.Operators = map[string]expr.Operator{
		"*": expr.OpMul,
		"/": expr.OpDiv,
		"+": expr.OpPlus,
		"-": expr.OpMinus,
		"<": expr.OpLt,
		">": expr.OpGt,
		"=": expr.OpEq,
	}

	r.SetOnOffDefaults(map[string]bool{
		"MSB": true, // MSB defaults to true, as per manual
		"LST": true, // Display listing: not used
	})

	r.SetAsciiVariation = func(in *inst.I, lp *lines.Parse) {
		if in.Command == "ASC" {
			if r.Setting("MSB") {
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

	return r
}

func (r *RedBook) LocalMacroLabels() bool {
	return false
}
