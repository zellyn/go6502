package as65

import (
	"fmt"

	"github.com/zellyn/go6502/asm/context"
	"github.com/zellyn/go6502/asm/expr"
	"github.com/zellyn/go6502/asm/flavors/common"
	"github.com/zellyn/go6502/asm/inst"
	"github.com/zellyn/go6502/asm/lines"
	"github.com/zellyn/go6502/opcodes"
)

// As65 implements an as65-compatible assembler flavor.
type As65 struct {
	common.Base
}

func New(sets opcodes.Set) *As65 {
	a := &As65{}
	a.Name = "as65"
	a.OpcodesByName = opcodes.ByName(sets)
	a.LabelChars = common.Letters + common.Digits + "."
	a.LabelColons = common.ReqOptional
	a.ExplicitARegister = common.ReqRequired
	a.StringEndOptional = true
	a.CommentChar = ';'
	a.MsbChars = "/"
	a.ImmediateChars = "#"
	a.HexCommas = common.ReqOptional
	a.DefaultOriginVal = 0x0800

	a.Directives = map[string]common.DirectiveInfo{
		"ORG":    {inst.TypeOrg, a.ParseOrg, 0},
		"OBJ":    {inst.TypeNone, nil, 0},
		"ENDASM": {inst.TypeEnd, a.ParseNoArgDir, 0},
		"EQU":    {inst.TypeEqu, a.ParseEquate, inst.VarEquNormal},
		"EPZ":    {inst.TypeEqu, a.ParseEquate, inst.VarEquPageZero},
		"DFB":    {inst.TypeData, a.ParseData, inst.VarBytes},
		"DW":     {inst.TypeData, a.ParseData, inst.VarWordsLe},
		"DDB":    {inst.TypeData, a.ParseData, inst.VarWordsBe},
		"ASC":    {inst.TypeData, a.ParseAscii, inst.VarAscii},
		"DCI":    {inst.TypeData, a.ParseAscii, inst.VarAsciiFlip},
		"HEX":    {inst.TypeData, a.ParseHexString, inst.VarBytes},
		"PAGE":   {inst.TypeNone, nil, 0}, // New page
		"TITLE":  {inst.TypeNone, nil, 0}, // Title
		"SBTL":   {inst.TypeNone, nil, 0}, // Subtitle
		"SKP":    {inst.TypeNone, nil, 0}, // Skip lines
		"REP":    {inst.TypeNone, nil, 0}, // Repeat character
		"CHR":    {inst.TypeNone, nil, 0}, // Set repeated character
	}

	a.EquateDirectives = map[string]bool{
		"EQU": true,
		"EPZ": true,
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

	a.InitContextFunc = func(ctx context.Context) {
		ctx.SetOnOffDefaults(map[string]bool{
			"MSB": true, // MSB defaults to true, as per manual
			"LST": true, // Display listing: not used
		})
	}

	a.SetAsciiVariation = func(ctx context.Context, in *inst.I, lp *lines.Parse) {
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

	a.FixLabel = a.DefaultFixLabel
	a.IsNewParentLabel = a.DefaultIsNewParentLabel

	return a
}
