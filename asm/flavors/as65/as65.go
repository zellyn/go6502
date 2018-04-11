package as65

import (
	"errors"
	"fmt"

	"github.com/zellyn/go6502/asm/context"
	"github.com/zellyn/go6502/asm/expr"
	"github.com/zellyn/go6502/asm/flavors/common"
	"github.com/zellyn/go6502/asm/inst"
	"github.com/zellyn/go6502/asm/lines"
	"github.com/zellyn/go6502/asm/opcodes"
)

var defaultCMap = [256]byte{
	0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
	0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f,
	0x20, 0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28, 0x29, 0x2a, 0x2b, 0x2c, 0x2d, 0x2e, 0x2f,
	0x30, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39, 0x3a, 0x3b, 0x3c, 0x3d, 0x3e, 0x3f,
	0x40, 0x41, 0x42, 0x43, 0x44, 0x45, 0x46, 0x47, 0x48, 0x49, 0x4a, 0x4b, 0x4c, 0x4d, 0x4e, 0x4f,
	0x50, 0x51, 0x52, 0x53, 0x54, 0x55, 0x56, 0x57, 0x58, 0x59, 0x5a, 0x5b, 0x5c, 0x5d, 0x5e, 0x5f,
	0x60, 0x61, 0x62, 0x63, 0x64, 0x65, 0x66, 0x67, 0x68, 0x69, 0x6a, 0x6b, 0x6c, 0x6d, 0x6e, 0x6f,
	0x70, 0x71, 0x72, 0x73, 0x74, 0x75, 0x76, 0x77, 0x78, 0x79, 0x7a, 0x7b, 0x7c, 0x7d, 0x7e, 0x7f,
	0x80, 0x81, 0x82, 0x83, 0x84, 0x85, 0x86, 0x87, 0x88, 0x89, 0x8a, 0x8b, 0x8c, 0x8d, 0x8e, 0x8f,
	0x90, 0x91, 0x92, 0x93, 0x94, 0x95, 0x96, 0x97, 0x98, 0x99, 0x9a, 0x9b, 0x9c, 0x9d, 0x9e, 0x9f,
	0xa0, 0xa1, 0xa2, 0xa3, 0xa4, 0xa5, 0xa6, 0xa7, 0xa8, 0xa9, 0xaa, 0xab, 0xac, 0xad, 0xae, 0xaf,
	0xb0, 0xb1, 0xb2, 0xb3, 0xb4, 0xb5, 0xb6, 0xb7, 0xb8, 0xb9, 0xba, 0xbb, 0xbc, 0xbd, 0xbe, 0xbf,
	0xc0, 0xc1, 0xc2, 0xc3, 0xc4, 0xc5, 0xc6, 0xc7, 0xc8, 0xc9, 0xca, 0xcb, 0xcc, 0xcd, 0xce, 0xcf,
	0xd0, 0xd1, 0xd2, 0xd3, 0xd4, 0xd5, 0xd6, 0xd7, 0xd8, 0xd9, 0xda, 0xdb, 0xdc, 0xdd, 0xde, 0xdf,
	0xe0, 0xe1, 0xe2, 0xe3, 0xe4, 0xe5, 0xe6, 0xe7, 0xe8, 0xe9, 0xea, 0xeb, 0xec, 0xed, 0xee, 0xef,
	0xf0, 0xf1, 0xf2, 0xf3, 0xf4, 0xf5, 0xf6, 0xf7, 0xf8, 0xf9, 0xfa, 0xfb, 0xfc, 0xfd, 0xfe, 0xff,
}

// As65 implements an as65-compatible assembler flavor.
type As65 struct {
	common.Base
	cmap [256]byte
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

		"db":  {inst.TypeData, a.ParseBytes, inst.VarBytes},
		"fcb": {inst.TypeData, a.ParseData, inst.VarBytes},
		"ds":  {inst.TypeData, a.ParseData, inst.VarBytesZero},
		"rmb": {inst.TypeData, a.ParseData, inst.VarBytesZero},
		"dw":  {inst.TypeData, a.ParseData, inst.VarWordsLe},
		"fcw": {inst.TypeData, a.ParseData, inst.VarWordsLe},
		"fdb": {inst.TypeData, a.ParseData, inst.VarWordsLe},

		"cmap": {inst.TypeNone, a.ParseCMap, 0},

		"DDB":   {inst.TypeData, a.ParseData, inst.VarWordsBe},
		"ASC":   {inst.TypeData, a.ParseAscii, inst.VarAscii},
		"DCI":   {inst.TypeData, a.ParseAscii, inst.VarAsciiFlip},
		"HEX":   {inst.TypeData, a.ParseHexString, inst.VarBytes},
		"PAGE":  {inst.TypeNone, nil, 0}, // New page
		"TITLE": {inst.TypeNone, nil, 0}, // Title
		"SBTL":  {inst.TypeNone, nil, 0}, // Subtitle
		"SKP":   {inst.TypeNone, nil, 0}, // Skip lines
		"REP":   {inst.TypeNone, nil, 0}, // Repeat character
		"CHR":   {inst.TypeNone, nil, 0}, // Set repeated character
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

// ParseCMap parses a character mapping.
// TODO(zellyn): implement
func (a *As65) ParseCMap(ctx context.Context, in inst.I, lp *lines.Parse) (inst.I, error) {
	lp.IgnoreRun(common.Whitespace)
	if lp.Peek() == lines.Eol || lp.Peek() == a.CommentChar {
		a.cmap = defaultCMap
		return in, nil
	}
	return in, errors.New("TODO(zellyn): parse cmap instruction")
}

// ParseBytes parses a db instruction.
func (a *As65) ParseBytes(ctx context.Context, in inst.I, lp *lines.Parse) (inst.I, error) {
	lp.IgnoreRun(common.Whitespace)
	return in, nil
}

func (a *As65) ParseByteList(lp *lines.Parse, cmap *[256]byte) (bytes []byte, final bool, err error) {
	return nil, false, nil
}
