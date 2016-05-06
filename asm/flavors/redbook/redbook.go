package redbook

import (
	"fmt"

	"github.com/zellyn/go6502/asm/context"
	"github.com/zellyn/go6502/asm/expr"
	"github.com/zellyn/go6502/asm/flavors/common"
	"github.com/zellyn/go6502/asm/inst"
	"github.com/zellyn/go6502/asm/lines"
	"github.com/zellyn/go6502/opcodes"
)

// RedBook implements a Redbook-listing-compatible-ish assembler flavor.
type RedBook struct {
	common.Base
}

func NewRedbookA(sets opcodes.Set) *RedBook {
	r := newRedbook("redbook-a", sets)
	return r
}

func NewRedbookB(sets opcodes.Set) *RedBook {
	r := newRedbook("redbook-b", sets)
	r.ExplicitARegister = common.ReqRequired
	r.SpacesForComment = 3
	return r
}

func newRedbook(name string, sets opcodes.Set) *RedBook {
	r := &RedBook{}
	r.Name = name
	r.OpcodesByName = opcodes.ByName(sets)
	r.LabelChars = common.Letters + common.Digits + "."
	r.LabelColons = common.ReqOptional
	r.ExplicitARegister = common.ReqRequired
	r.StringEndOptional = true
	r.CommentChar = ';'
	r.MsbChars = "/"
	r.ImmediateChars = "#"
	r.HexCommas = common.ReqOptional
	r.DefaultOriginVal = 0x0800

	r.Directives = map[string]common.DirectiveInfo{
		"ORG":    {inst.TypeOrg, r.ParseOrg, 0},
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

	r.EquateDirectives = map[string]bool{
		"EQU": true,
		"EPZ": true,
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

	r.InitContextFunc = func(ctx context.Context) {
		ctx.SetOnOffDefaults(map[string]bool{
			"MSB": true, // MSB defaults to true, as per manual
			"LST": true, // Display listing: not used
		})
	}

	r.SetAsciiVariation = func(ctx context.Context, in *inst.I, lp *lines.Parse) {
		if in.Command == "ASC" {
			if ctx.Setting("MSB") {
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
