package oldschool

import (
	"encoding/hex"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/zellyn/go6502/asm/context"
	"github.com/zellyn/go6502/asm/expr"
	"github.com/zellyn/go6502/asm/flavors/common"
	"github.com/zellyn/go6502/asm/inst"
	"github.com/zellyn/go6502/asm/lines"
	"github.com/zellyn/go6502/opcodes"
)

const Letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz_"
const Digits = "0123456789"
const binarydigits = "01"
const hexdigits = Digits + "abcdefABCDEF"
const Whitespace = " \t"
const cmdChars = Letters + Digits + ".<>_:@"
const fileChars = Letters + Digits + "."

type DirectiveInfo struct {
	Type inst.Type
	Func func(inst.I, *lines.Parse) (inst.I, error)
	Var  int
}

type Requiredness int

const (
	ReqDisallowed Requiredness = iota
	ReqOptional
	ReqRequired
)

// Base implements the S-C Macro Assembler-compatible assembler flavor.
// See http://www.txbobsc.com/scsc/ and http://stjarnhimlen.se/apple2/
type Base struct {
	Directives map[string]DirectiveInfo
	Operators  map[string]expr.Operator
	context.SimpleContext
	context.LabelerBase
	LabelChars        string
	LabelColons       Requiredness
	ExplicitARegister Requiredness
	HexCommas         Requiredness
	SpacesForComment  int  // this many spaces after command means it's the comment field
	StringEndOptional bool // can omit closing delimeter from string args?
	SuffixForWide     bool // is eg. "LDA:" a force-wide on "LDA"? (Merlin)
	CommentChar       rune
	BinaryChar        rune
	MsbChars          string
	LsbChars          string
	ImmediateChars    string
	ExtraCommenty     func(string) bool
	SetAsciiVariation func(*inst.I, *lines.Parse)
	ParseMacroCall    func(inst.I, *lines.Parse) (inst.I, bool, error)
	operatorChars     string
	CharChars         string
	InvCharChars      string
	MacroArgSep       string
}

// Parse an entire instruction, or return an appropriate error.
func (a *Base) ParseInstr(line lines.Line) (inst.I, error) {
	lp := line.Parse
	in := inst.I{Line: &line}

	// Lines that start with a digit are considered to have a declared line number.
	if lp.AcceptRun(Digits) {
		s := lp.Emit()
		if len(s) != 4 {
			return in, line.Errorf("line number must be exactly 4 digits: %s", s)
		}
		if !lp.Consume(" ") && lp.Peek() != lines.Eol {
			return in, line.Errorf("line number (%s) followed by non-space", s)
		}
		i, err := strconv.ParseUint(s, 10, 16)
		if err != nil {
			return in, line.Errorf("invalid line number: %s: %s", s, err)
		}
		in.DeclaredLine = uint16(i)
	}

	// Flavor considers this line extra commenty for some reason?
	if a.ExtraCommenty != nil && a.ExtraCommenty(lp.Rest()) {
		in.Type = inst.TypeNone
		return in, nil
	}

	// Empty line or comment
	trimmed := strings.TrimSpace(lp.Rest())
	if trimmed == "" || trimmed[0] == '*' || rune(trimmed[0]) == a.CommentChar {
		in.Type = inst.TypeNone
		return in, nil
	}

	// See if we have a label at the start
	if lp.AcceptRun(a.LabelChars) {
		in.Label = lp.Emit()

		// Some need colons after labels, some allow them.
		switch a.LabelColons {
		case ReqRequired:
			if !lp.Consume(":") {
				return in, line.Errorf("label '%s' must end in colon", in.Label)
			}
		case ReqOptional:
			lp.Consume(":")
		}
	}

	// Ignore whitespace at the start or after the label.
	lp.IgnoreRun(Whitespace)

	if lp.Peek() == lines.Eol || lp.Peek() == a.CommentChar {
		in.Type = inst.TypeNone
		return in, nil
	}
	return a.ParseCmd(in, lp)
}

func (a *Base) DefaultOrigin() (uint16, error) {
	return 0x0800, nil
}

func (a *Base) SetWidthsOnFirstPass() bool {
	return true
}

// ParseCmd parses the "command" part of an instruction: we expect to be
// looking at a non-whitespace character.
func (a *Base) ParseCmd(in inst.I, lp *lines.Parse) (inst.I, error) {
	if !lp.AcceptRun(cmdChars) && !(a.Directives["="].Func != nil && lp.Accept("=")) {
		c := lp.Next()
		return in, in.Errorf("expecting instruction, found '%c' (%d)", c, c)
	}
	in.Command = lp.Emit()

	// Give ParseMacroCall a chance
	if a.ParseMacroCall != nil {
		i, isMacro, err := a.ParseMacroCall(in, lp)
		if err != nil {
			return in, err
		}
		if isMacro {
			return i, nil
		}
	}

	if dir, ok := a.Directives[in.Command]; ok {
		in.Type = dir.Type
		in.Var = dir.Var
		if dir.Func == nil {
			return in, nil
		}
		return dir.Func(in, lp)
	}

	if a.HasSetting(in.Command) {
		return a.ParseSetting(in, lp)
	}

	if summary, ok := opcodes.ByName[in.Command]; ok {
		in.Type = inst.TypeOp
		return a.ParseOpArgs(in, lp, summary, false)
	}

	// Merlin lets you say "LDA:" or "LDA@" or "LDAZ" to force non-zero-page.
	if a.SuffixForWide {
		prefix := in.Command[:len(in.Command)-1]
		if summary, ok := opcodes.ByName[prefix]; ok {
			in.Command = prefix
			in.Type = inst.TypeOp
			return a.ParseOpArgs(in, lp, summary, true)
		}
	}

	return in, in.Errorf(`unknown command/instruction: "%s"`, in.Command)
}

func (a *Base) ParseSetting(in inst.I, lp *lines.Parse) (inst.I, error) {
	in.Type = inst.TypeSetting
	lp.IgnoreRun(Whitespace)
	if !lp.AcceptRun(Letters) {
		c := lp.Next()
		return in, in.Errorf("expecting ON/OFF, found '%s'", c)
	}
	in.TextArg = lp.Emit()
	switch in.TextArg {
	case "ON":
		a.SettingOn(in.Command)
	case "OFF":
		a.SettingOff(in.Command)
	default:
		return in, in.Errorf("expecting ON/OFF, found '%s'", in.TextArg)
	}
	return in, nil

}

// ParseMacroArg parses a single macro argument. We expect to be looking at the first
// character of a macro argument.
func (a *Base) ParseMacroArg(in inst.I, lp *lines.Parse) (string, error) {
	if lp.Peek() == '"' {
		return a.ParseQuoted(in, lp)
	}
	lp.AcceptUntil(Whitespace + a.MacroArgSep)
	return lp.Emit(), nil
}

// ParseQuoted parses a single quoted string macro argument. We expect
// to be looking at the first quote.
func (a *Base) ParseQuoted(in inst.I, lp *lines.Parse) (string, error) {
	if !lp.Consume(`"`) {
		panic(fmt.Sprintf("ParseQuoted called not looking at a quote"))
	}
	for {
		lp.AcceptUntil(`"`)
		// We're done, unless there's an escaped quote
		if !lp.AcceptString(`""`) {
			break
		}
	}
	s := lp.Emit()
	if !lp.Consume(`"`) {
		c := lp.Peek()
		return "", in.Errorf("Expected closing quote; got %s", c)
	}

	c := lp.Peek()
	if c != ',' && c != ' ' && c != lines.Eol && c != '\t' {
		return "", in.Errorf("Unexpected char after quoted string: '%s'", c)
	}

	return strings.Replace(s, `""`, `"`, -1), nil
}

// ParseOpArgs parses the arguments to an assembly op. We expect to be
// looking at the first non-op character (probably whitespace)
func (a *Base) ParseOpArgs(in inst.I, lp *lines.Parse, summary opcodes.OpSummary, forceWide bool) (inst.I, error) {
	// MODE_IMPLIED: we don't really care what comes next: it's a comment.
	if summary.Modes == opcodes.MODE_IMPLIED {
		op := summary.Ops[0]
		in.Data = []byte{op.Byte}
		in.WidthKnown = true
		in.MinWidth = 1
		in.MaxWidth = 1
		in.Final = true
		in.Mode = opcodes.MODE_IMPLIED
		return in, nil
	}

	// Nothing else on the line? Must be MODE_A
	lp.AcceptRun(Whitespace)
	ws := lp.Emit()
	atEnd := false
	if a.SpacesForComment != 0 && len(ws) >= a.SpacesForComment {
		atEnd = true
	}
	if lp.Peek() == lines.Eol || lp.Peek() == a.CommentChar {
		atEnd = true
	}

	if atEnd {
		// Nothing left on line except comments.
		if !summary.AnyModes(opcodes.MODE_A) {
			return in, in.Errorf("%s with no arguments", in.Command)
		}
		op, ok := summary.OpForMode(opcodes.MODE_A)
		if !ok {
			panic(fmt.Sprintf("%s doesn't support accumulator mode", in.Command))
		}
		in.Data = []byte{op.Byte}
		in.WidthKnown = true
		in.MinWidth = 1
		in.MaxWidth = 1
		in.Final = true
		in.Mode = opcodes.MODE_A
		return in, nil
	}

	indirect := lp.Consume("(")
	if indirect && !summary.AnyModes(opcodes.MODE_INDIRECT_ANY) {
		return in, in.Errorf("%s doesn't support any indirect modes", in.Command)
	}
	xy := '-'
	expr, err := a.ParseExpression(in, lp)
	if err != nil {
		return in, err
	}
	if !indirect && (expr.Text == "a" || expr.Text == "A") {
		if !summary.AnyModes(opcodes.MODE_A) {
			return in, in.Errorf("%s doesn't support A mode", in.Command)
		}
		switch a.ExplicitARegister {
		case ReqDisallowed:
			return in, in.Errorf("Assembler flavor doesn't support A mode", in.Command)
		case ReqOptional, ReqRequired:
			op, ok := summary.OpForMode(opcodes.MODE_A)
			if !ok {
				panic(fmt.Sprintf("%s doesn't support accumulator mode", in.Command))
			}
			in.Data = []byte{op.Byte}
			in.WidthKnown = true
			in.MinWidth = 1
			in.MaxWidth = 1
			in.Final = true
			in.Mode = opcodes.MODE_A
			in.Exprs = nil
			return in, nil

		}
	}
	in.Exprs = append(in.Exprs, expr)
	comma := lp.Consume(",")
	if comma {
		if lp.Consume("xX") {
			xy = 'x'
		} else if lp.Consume("yY") {
			if indirect {
				return in, in.Errorf(",Y unexpected inside parens")
			}
			xy = 'y'
		} else {
			return in, in.Errorf("X or Y expected after comma")
		}
	}
	comma2 := false
	if indirect {
		if !lp.Consume(")") {
			return in, in.Errorf("Expected closing paren")
		}
		comma2 = lp.Consume(",")
		if comma2 {
			if comma {
				return in, in.Errorf("Cannot have ,X or ,Y twice.")
			}
			if !lp.Consume("yY") {
				return in, in.Errorf("Only ,Y can follow parens.")
			}
			xy = 'y'
		}
	}

	return common.DecodeOp(a, in, summary, indirect, xy, forceWide)
}

func (a *Base) ParseAddress(in inst.I, lp *lines.Parse) (inst.I, error) {
	lp.IgnoreRun(Whitespace)
	expr, err := a.ParseExpression(in, lp)
	if err != nil {
		return in, err
	}
	in.Exprs = append(in.Exprs, expr)
	in.WidthKnown = true
	in.MinWidth = 0
	in.MaxWidth = 0
	in.Final = true
	return in, nil
}

func (a *Base) ParseAscii(in inst.I, lp *lines.Parse) (inst.I, error) {
	lp.IgnoreRun(Whitespace)
	a.SetAsciiVariation(&in, lp)
	var invert, invertLast byte
	switch in.Var {
	case inst.DataAscii:
	case inst.DataAsciiHi:
		invert = 0x80
	case inst.DataAsciiFlip:
		invertLast = 0x80
	case inst.DataAsciiHiFlip:
		invert = 0x80
		invertLast = 0x80
	default:
		panic(fmt.Sprintf("ParseAscii with weird Variation: %d", in.Var))
	}
	delim := lp.Next()
	if delim == lines.Eol || strings.IndexRune(Whitespace, delim) >= 0 {
		return in, in.Errorf("%s expects delimeter, found '%s'", in.Command, delim)
	}
	lp.Ignore()
	lp.AcceptUntil(string(delim))
	delim2 := lp.Next()
	if delim != delim2 && !(delim2 == lines.Eol && a.StringEndOptional) {
		return in, in.Errorf("%s: expected closing delimeter '%s'; got '%s'", in.Command, delim, delim2)
	}
	lp.Backup()
	in.Data = []byte(lp.Emit())
	for i := range in.Data {
		in.Data[i] ^= invert
		if i == len(in.Data)-1 {
			in.Data[i] ^= invertLast
		}
	}
	return in, nil
}

func (a *Base) ParseBlockStorage(in inst.I, lp *lines.Parse) (inst.I, error) {
	lp.IgnoreRun(Whitespace)
	ex, err := a.ParseExpression(in, lp)
	if err != nil {
		return in, err
	}
	in.Exprs = append(in.Exprs, ex)
	return in, nil
}

func (a *Base) ParseData(in inst.I, lp *lines.Parse) (inst.I, error) {
	lp.IgnoreRun(Whitespace)
	for {
		ex, err := a.ParseExpression(in, lp)
		if err != nil {
			return in, err
		}
		in.Exprs = append(in.Exprs, ex)
		if !lp.Consume(",") {
			break
		}
	}
	return in, nil
}

func (a *Base) ParseDo(in inst.I, lp *lines.Parse) (inst.I, error) {
	lp.IgnoreRun(Whitespace)
	expr, err := a.ParseExpression(in, lp)
	if err != nil {
		return in, err
	}
	in.Exprs = append(in.Exprs, expr)
	in.WidthKnown = true
	in.MinWidth = 0
	in.MaxWidth = 0
	in.Final = true
	return in, nil
}

func (a *Base) ParseEquate(in inst.I, lp *lines.Parse) (inst.I, error) {
	lp.IgnoreRun(Whitespace)
	expr, err := a.ParseExpression(in, lp)
	if err != nil {
		return in, err
	}
	in.Exprs = append(in.Exprs, expr)
	in.WidthKnown = true
	in.MinWidth = 0
	in.MaxWidth = 0
	in.Final = true
	return in, nil
}

func (a *Base) ParseHexString(in inst.I, lp *lines.Parse) (inst.I, error) {
	lp.AcceptRun(Whitespace)
	for {
		lp.Ignore()
		if !lp.AcceptRun(hexdigits) {
			return in, in.Errorf("%s expects hex digits; got '%s'", in.Command, lp.Next())
		}
		hs := lp.Emit()
		if len(hs)%2 != 0 {
			return in, in.Errorf("%s expects pairs of hex digits; got %d", in.Command, len(hs))
		}
		data, err := hex.DecodeString(hs)
		if err != nil {
			return in, in.Errorf("%s: error decoding hex string: %s", in.Command, err)
		}
		in.Data = append(in.Data, data...)

		// Keep going if we allow commas and have one
		if a.HexCommas == ReqDisallowed || !lp.Accept(",") {
			break
		}
	}
	return in, nil
}

func (a *Base) ParseInclude(in inst.I, lp *lines.Parse) (inst.I, error) {
	lp.IgnoreRun(Whitespace)
	if !lp.AcceptRun(fileChars) {
		return in, in.Errorf("Expecting filename, found '%c'", lp.Next())
	}
	in.TextArg = lp.Emit()
	in.WidthKnown = true
	in.MinWidth = 0
	in.MaxWidth = 0
	in.Final = true
	return in, nil
}

// For assemblers where the macro name follows the macro directive.
func (a *Base) ParseMacroStart(in inst.I, lp *lines.Parse) (inst.I, error) {
	lp.IgnoreRun(Whitespace)
	if !lp.AcceptRun(cmdChars) {
		return in, in.Errorf("Expecting valid macro name, found '%c'", lp.Next())
	}
	in.TextArg = lp.Emit()
	in.WidthKnown = true
	in.MinWidth = 0
	in.MaxWidth = 0
	in.Final = true
	return in, nil
}

// For assemblers where the macro name is the label, followed by the directive.
func (a *Base) MarkMacroStart(in inst.I, lp *lines.Parse) (inst.I, error) {
	in.TextArg, in.Label = in.Label, ""
	in.WidthKnown = true
	in.MinWidth = 0
	in.MaxWidth = 0
	in.Final = true
	return in, nil
}

func (a *Base) ParseNoArgDir(in inst.I, lp *lines.Parse) (inst.I, error) {
	in.WidthKnown = true
	in.MinWidth = 0
	in.MaxWidth = 0
	in.Final = true
	return in, nil
}

func (a *Base) ParseNotImplemented(in inst.I, lp *lines.Parse) (inst.I, error) {
	return in, in.Errorf("not implemented (yet?): %s", in.Command)
}

func (a *Base) ParseExpression(in inst.I, lp *lines.Parse) (*expr.E, error) {

	if a.operatorChars == "" {
		for k, _ := range a.Operators {
			a.operatorChars += k
		}
	}

	var outer *expr.E
	if lp.AcceptRun(a.MsbChars + a.LsbChars + a.ImmediateChars) {
		pc := lp.Emit()
		switch len(pc) {
		case 1:
			switch {
			case strings.Contains(a.MsbChars, pc[:1]):
				outer = &expr.E{Op: expr.OpMsb}
			case strings.Contains(a.LsbChars+a.ImmediateChars, pc[:1]):
				outer = &expr.E{Op: expr.OpLsb}
			}
		case 2:
			err := in.Errorf("Got strange number prefix: '%s'", pc)
			switch {
			case !strings.Contains(a.ImmediateChars, pc[:1]):
				return &expr.E{}, err
			case strings.Contains(a.MsbChars, pc[1:]):
				outer = &expr.E{Op: expr.OpMsb}
			case strings.Contains(a.LsbChars, pc[1:]):
				outer = &expr.E{Op: expr.OpLsb}
			default:
				return &expr.E{}, err
			}
		default:
			return &expr.E{}, in.Errorf("Expected one or two number prefixes, got '%s'", pc)
		}
	}

	tree, err := a.ParseTerm(in, lp)
	if err != nil {
		return &expr.E{}, err
	}

	for lp.Accept(a.operatorChars) {
		c := lp.Emit()
		right, err := a.ParseTerm(in, lp)
		if err != nil {
			return &expr.E{}, err
		}
		tree = &expr.E{Op: a.Operators[c], Left: tree, Right: right}
	}

	if outer != nil {
		outer.Left = tree
		return outer, nil
	}
	return tree, nil
}

func (a *Base) ParseTerm(in inst.I, lp *lines.Parse) (*expr.E, error) {
	ex := &expr.E{}
	top := ex

	// Unary minus: just wrap the current expression
	if lp.Consume("-") {
		top = &expr.E{Op: expr.OpMinus, Left: ex}
	}

	// Current location
	if lp.Consume("*") {
		ex.Op = expr.OpLeaf
		ex.Text = "*"
		return top, nil
	}

	// Hex
	if lp.Consume("$") {
		if !lp.AcceptRun(hexdigits) {
			c := lp.Next()
			return &expr.E{}, in.Errorf("expecting hex number, found '%c' (%d)", c, c)
		}
		s := lp.Emit()
		i, err := strconv.ParseUint(s, 16, 16)
		if err != nil {
			return &expr.E{}, in.Errorf("invalid hex number: %s: %s", s, err)
		}
		ex.Op = expr.OpLeaf
		ex.Val = uint16(i)
		return top, nil
	}

	// Binary
	if lp.Consume(string(a.BinaryChar)) {
		if !lp.AcceptRun(binarydigits) {
			c := lp.Next()
			return &expr.E{}, in.Errorf("expecting binary number, found '%c' (%d)", c, c)
		}
		s := lp.Emit()
		i, err := strconv.ParseUint(s, 2, 16)
		if err != nil {
			return &expr.E{}, in.Errorf("invalid binary number: %s: %s", s, err)
		}
		ex.Op = expr.OpLeaf
		ex.Val = uint16(i)
		return top, nil
	}

	// Decimal
	if lp.AcceptRun(Digits) {
		s := lp.Emit()
		i, err := strconv.ParseUint(s, 10, 16)
		if err != nil {
			return &expr.E{}, in.Errorf("invalid number: %s: %s", s, err)
		}
		ex.Op = expr.OpLeaf
		ex.Val = uint16(i)
		return top, nil
	}

	// Character
	allChars := a.CharChars + a.InvCharChars
	if allChars != "" && lp.Accept(allChars) {
		quote := lp.Emit()
		c := lp.Next()
		if c == lines.Eol {
			return &expr.E{}, in.Errorf("end of line after quote")
		}
		ex.Op = expr.OpLeaf
		ex.Val = uint16(c)
		if strings.Contains(a.InvCharChars, quote[:1]) {
			ex.Val |= 0x80
		}
		lp.Consume(quote) // optional closing quote
		lp.Ignore()
		return top, nil
	}

	// Label
	if !lp.AcceptRun(a.LabelChars) {
		c := lp.Next()
		return &expr.E{}, in.Errorf("expecting *, (hex) number, or label; found '%c' (%d)", c, c)
	}

	ex.Op = expr.OpLeaf
	ex.Text = lp.Emit()
	return top, nil
}

var macroArgRe = regexp.MustCompile("][0-9]+")

func (a *Base) ReplaceMacroArgs(line string, args []string, kwargs map[string]string) (string, error) {
	var err error
	line = strings.Replace(line, "]#", strconv.Itoa(len(args)), -1)
	line = string(macroArgRe.ReplaceAllFunc([]byte(line), func(in []byte) []byte {
		n, _ := strconv.Atoi(string(in[1:]))
		if n > 0 && n <= len(args) {
			return []byte(args[n-1])
		}
		return in
	}))
	return line, err
}

func (a *Base) IsNewParentLabel(label string) bool {
	return label != "" && label[0] != '.'
}

func (a *Base) FixLabel(label string, macroCall int, locals map[string]bool) (string, error) {
	switch {
	case label == "":
		return label, nil
	case label[0] == '.':
		if last := a.LastLabel(); last == "" {
			return "", fmt.Errorf("sublabel '%s' without previous label", label)
		} else {
			return fmt.Sprintf("%s/%s", last, label), nil
		}
	case label[0] == ':':
		if macroCall == 0 {
			return "", fmt.Errorf("macro-local label '%s' seen outside macro", label)
		} else {
			return fmt.Sprintf("%s/%d", label, macroCall), nil
		}
	}
	return label, nil
}
