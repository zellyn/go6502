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
	r := newRedbook("redbook-a")
	return r
}

func NewRedbookB() *RedBook {
	r := newRedbook("redbook-b")
	r.ExplicitARegister = oldschool.ReqRequired
	r.SpacesForComment = 3
	return r
}

func newRedbook(name string) *RedBook {
	r := &RedBook{}
	r.Name = name
	r.LabelChars = oldschool.Letters + oldschool.Digits + "."
	r.LabelColons = oldschool.ReqOptional
	r.ExplicitARegister = oldschool.ReqRequired
	r.StringEndOptional = true
	r.CommentChar = ';'
	r.MsbChars = "/"
	r.ImmediateChars = "#"
	r.HexCommas = oldschool.ReqOptional

	r.Directives = map[string]oldschool.DirectiveInfo{
		"ORG":    {inst.TypeOrg, r.ParseAddress, 0},
		"OBJ":    {inst.TypeNone, nil, 0},
		"ENDASM": {inst.TypeEnd, r.ParseNoArgDir, 0},
		"EQU":    {inst.TypeEqu, r.ParseEquate, inst.VarEquNormal},
		"EPZ":    {inst.TypeEqu, r.ParseEquate, inst.VarEquPageZero},
		"DFB":    {inst.TypeData, r.ParseData, inst.VarBytes},
		"DW":     {inst.TypeData, r.ParseData, inst.VarWordsLe},
		"DDB":    {inst.TypeData, r.ParseData, inst.VarWordsBe},
		"ASC":    {inst.TypeData, r.ParseAscii, inst.VarAscii},
		"DCI":    {inst.TypeData, r.ParseAscii, inst.VarAsciiFlip},
		"HEX":    {inst.TypeData, r.ParseHexString, inst.VarBytes},
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
				in.Var = inst.VarAsciiHi
			} else {
				in.Var = inst.VarAscii
			}
			return
		}
		if in.Command == "DCI" {
			in.Var = inst.VarAsciiFlip
		} else {
			panic(fmt.Sprintf("Unknown ascii directive: '%s'", in.Command))
		}
	}

	r.FixLabel = r.DefaultFixLabel
	r.IsNewParentLabel = r.DefaultIsNewParentLabel

	return r
}

func (r *RedBook) LocalMacroLabels() bool {
	return false
}
