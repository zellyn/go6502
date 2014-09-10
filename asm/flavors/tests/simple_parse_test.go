package flavors

import (
	"encoding/hex"
	"testing"

	"github.com/zellyn/go6502/asm/context"
	"github.com/zellyn/go6502/asm/flavors"
	"github.com/zellyn/go6502/asm/flavors/as65"
	"github.com/zellyn/go6502/asm/flavors/merlin"
	"github.com/zellyn/go6502/asm/flavors/redbook"
	"github.com/zellyn/go6502/asm/flavors/scma"
	"github.com/zellyn/go6502/asm/inst"
	"github.com/zellyn/go6502/asm/lines"
)

func TestSimpleCommonFunctions(t *testing.T) {
	ss := scma.New()
	ra := redbook.NewRedbookA()
	rb := redbook.NewRedbookB()
	// aa := as65.New()
	mm := merlin.New()

	tests := []struct {
		f flavors.F // assembler flavor
		i string    // input string
		p string    // printed instruction, expected
		b string    // bytes, expected
	}{
		// {aa, " beq $2343", "{BEQ/rel $2343}", "f0fc"},
		// {aa, " beq $2345", "{BEQ/rel $2345}", "f0fe"},
		// {aa, " beq $2347", "{BEQ/rel $2347}", "f000"},
		// {aa, " dw $1234", "{data/wle $1234}", "3412"},
		// {aa, " jmp $1234", "{JMP/abs $1234}", "4c3412"},
		// {aa, " jmp ($1234)", "{JMP/ind $1234}", "6c3412"},
		// {aa, " lda #$12", "{LDA/imm (lsb $0012)}", "a912"},
		// {aa, " lda $12", "{LDA/zp $0012}", "a512"},
		// {aa, " lda $12,x", "{LDA/zpX $0012}", "b512"},
		// {aa, " lda $1234", "{LDA/abs $1234}", "ad3412"},
		// {aa, " lda $1234,x", "{LDA/absX $1234}", "bd3412"},
		// {aa, " lda ($12),y", "{LDA/indY $0012}", "b112"},
		// {aa, " lda ($12,x)", "{LDA/indX $0012}", "a112"},
		// {aa, " ldx $12,y", "{LDX/zpY $0012}", "b612"},
		// {aa, " org $D000", "{org $d000}", ""},
		// {aa, " rol $12", "{ROL/zp $0012}", "2612"},
		// {aa, " rol $1234", "{ROL/abs $1234}", "2e3412"},
		// {aa, " rol a", "{ROL/a}", "2a"},
		// {aa, " sta $1234,y", "{STA/absY $1234}", "993412"},
		// {aa, "; Comment", "{-}", ""},
		// {aa, "Label", "{- 'Label'}", ""},
		// {aa, ` include "FILE.NAME"`, "{inc 'FILE.NAME'}", ""},
		// {aa, ` title "Title here"`, "{-}", ""},
		// {ss, " .TA *-1234", "{target (- * $04d2)}", ""},
		{mm, " <<<", `{endm}`, ""},
		{mm, " >>> M1,$42 ;$43", `{call M1 {"$42"}}`, ""},
		{mm, " >>> M1.$42", `{call M1 {"$42"}}`, ""},
		{mm, " >>> M1/$42;$43", `{call M1 {"$42", "$43"}}`, ""},
		{mm, " BEQ $2343", "{BEQ/rel $2343}", "f0fc"},
		{mm, " BEQ $2345", "{BEQ/rel $2345}", "f0fe"},
		{mm, " BEQ $2347", "{BEQ/rel $2347}", "f000"},
		{mm, " DA $12,$34,$1234", "{data/wle $0012,$0034,$1234}", "120034003412"},
		{mm, " DA $1234", "{data/wle $1234}", "3412"},
		{mm, " DB <L4,<L5", "{data/b (lsb L4),(lsb L5)}", "deef"},
		{mm, " DDB $12,$34,$1234", "{data/wbe $0012,$0034,$1234}", "001200341234"},
		{mm, " DFB $12,$34,$1234", "{data/b $0012,$0034,$1234}", "123434"},
		{mm, " DFB $34,100,$81A2-$77C4,%1011,>$81A2-$77C4", "{data/b $0034,$0064,(- $81a2 $77c4),$000b,(msb (- $81a2 $77c4))}", "3464de0b09"},
		{mm, " DSK OUTFILE", "{-}", ""},
		{mm, " EOM", `{endm}`, ""},
		{mm, " HEX 00,01,FF,AB", "{data/b}", "0001ffab"},
		{mm, " HEX 0001FFAB", "{data/b}", "0001ffab"},
		{mm, " INCW $42;$43", `{call INCW {"$42", "$43"}}`, ""},
		{mm, " JMP $1234", "{JMP/abs $1234}", "4c3412"},
		{mm, " JMP ($1234)", "{JMP/ind $1234}", "6c3412"},
		{mm, " LDA #$12", "{LDA/imm (lsb $0012)}", "a912"},
		{mm, " LDA #$1234", "{LDA/imm (lsb $1234)}", "a934"},
		{mm, " LDA #/$1234", "{LDA/imm (msb $1234)}", "a912"},
		{mm, " LDA #<$1234", "{LDA/imm (lsb $1234)}", "a934"},
		{mm, " LDA #>$1234", "{LDA/imm (msb $1234)}", "a912"},
		{mm, " LDA $12", "{LDA/zp $0012}", "a512"},
		{mm, " LDA $12", "{LDA/zp $0012}", "a512"},
		{mm, " LDA $12,X", "{LDA/zpX $0012}", "b512"},
		{mm, " LDA $1234", "{LDA/abs $1234}", "ad3412"},
		{mm, " LDA $1234", "{LDA/abs $1234}", "ad3412"},
		{mm, " LDA $1234,X", "{LDA/absX $1234}", "bd3412"},
		{mm, " LDA ($12),Y", "{LDA/indY $0012}", "b112"},
		{mm, " LDA ($12,X)", "{LDA/indX $0012}", "a112"},
		{mm, " LDA: $12", "{LDA/abs $0012}", "ad1200"},
		{mm, " LDA@ $12", "{LDA/abs $0012}", "ad1200"},
		{mm, " LDAX $12", "{LDA/abs $0012}", "ad1200"},
		{mm, " LDX $12,Y", "{LDX/zpY $0012}", "b612"},
		{mm, " ORG $D000", "{org $d000}", ""},
		{mm, " PMC M1($42", `{call M1 {"$42"}}`, ""},
		{mm, " PMC M1-$42", `{call M1 {"$42"}}`, ""},
		{mm, " PUT !FILE.NAME", "{inc 'FILE.NAME'}", ""},
		{mm, " ROL $12", "{ROL/zp $0012}", "2612"},
		{mm, " ROL $1234", "{ROL/abs $1234}", "2e3412"},
		{mm, " ROL", "{ROL/a}", "2a"},
		{mm, " SAV OUTFILE", "{-}", ""},
		{mm, " STA $1234,Y", "{STA/absY $1234}", "993412"},
		{mm, "* Comment", "{-}", ""},
		{mm, "ABC = $800", "{= 'ABC' $0800}", ""},
		{mm, "L1 = 'A.2", "{= 'L1' (| $0041 $0002)}", ""},
		{mm, "L1 = *-2", "{= 'L1' (- * $0002)}", ""},
		{mm, "L1 = 1234+%10111", "{= 'L1' (+ $04d2 $0017)}", ""},
		{mm, "L1 = 2*L2+$231", "{= 'L1' (+ (* $0002 L2) $0231)}", ""},
		{mm, "L1 = L2!'A'", "{= 'L1' (^ L2 $0041)}", ""},
		{mm, "L1 = L2&$7F", "{= 'L1' (& L2 $007f)}", ""},
		{mm, "L1 = L2-L3", "{= 'L1' (- L2 L3)}", ""},
		{mm, "L1 = L2.%10000000", "{= 'L1' (| L2 $0080)}", ""},
		{mm, "Label ;Comment", "{- 'Label'}", ""},
		{mm, "Label", "{- 'Label'}", ""},
		{mm, "Label", "{- 'Label'}", ""},
		{mm, "Label;Comment", "{- 'Label'}", ""},
		{mm, "MacroName MAC", `{macro "MacroName"}`, ""},
		{mm, "MacroName MAC", `{macro "MacroName"}`, ""},
		{mm, ` ASC !ABC!`, "{data/b}", "c1c2c3"},
		{mm, ` ASC "ABC"`, "{data/b}", "c1c2c3"},
		{mm, ` ASC #ABC#`, "{data/b}", "c1c2c3"},
		{mm, ` ASC $ABC$`, "{data/b}", "c1c2c3"},
		{mm, ` ASC %ABC%`, "{data/b}", "c1c2c3"},
		{mm, ` ASC &ABC&`, "{data/b}", "c1c2c3"},
		{mm, ` ASC 'ABC'`, "{data/b}", "414243"},
		{mm, ` ASC (ABC(`, "{data/b}", "414243"},
		{mm, ` ASC )ABC)`, "{data/b}", "414243"},
		{mm, ` ASC +ABC+`, "{data/b}", "414243"},
		{mm, ` ASC ?ABC?`, "{data/b}", "414243"},
		{mm, ` DCI !ABC!`, "{data/b}", "c1c243"},
		{mm, ` DCI "ABC"`, "{data/b}", "c1c243"},
		{mm, ` DCI #ABC#`, "{data/b}", "c1c243"},
		{mm, ` DCI $ABC$`, "{data/b}", "c1c243"},
		{mm, ` DCI %ABC%`, "{data/b}", "c1c243"},
		{mm, ` DCI &ABC&`, "{data/b}", "c1c243"},
		{mm, ` DCI 'ABC'`, "{data/b}", "4142c3"},
		{mm, ` DCI (ABC(`, "{data/b}", "4142c3"},
		{mm, ` DCI )ABC)`, "{data/b}", "4142c3"},
		{mm, ` DCI +ABC+`, "{data/b}", "4142c3"},
		{mm, ` DCI ?ABC?`, "{data/b}", "4142c3"},
		{mm, ` TTL "Title here"`, "{-}", ""},
		{mm, `L1 = "A+3`, "{= 'L1' (+ $00c1 $0003)}", ""},
		{mm, `L1 = L2!"A"`, "{= 'L1' (^ L2 $00c1)}", ""},
		{ra, "                       ; Comment", "{-}", ""},
		{ra, " DDB $12,$34,$1234", "{data/wbe $0012,$0034,$1234}", "001200341234"},
		{ra, " DFB $12", "{data/b $0012}", "12"},
		{ra, " DFB $12,$34,$1234", "{data/b $0012,$0034,$1234}", "123434"},
		{ra, " DW $12,$34,$1234", "{data/wle $0012,$0034,$1234}", "120034003412"},
		{ra, " LST OFF", "{set LST OFF}", ""},
		{ra, " LST ON", "{set LST ON}", ""},
		{ra, " MSB OFF", "{set MSB OFF}", ""},
		{ra, " MSB ON", "{set MSB ON}", ""},
		{ra, " ORG $D000", "{org $d000}", ""},
		{ra, " ROL  A", "{ROL/a}", "2a"}, // two spaces is no big deal
		{ra, " ROL A", "{ROL/a}", "2a"},
		{ra, "* Comment", "{-}", ""},
		{ra, "Label", "{- 'Label'}", ""},
		{ra, "Label:", "{- 'Label'}", ""},
		{ra, ` ASC "ABC"`, "{data/b}", "c1c2c3"},
		{ra, ` ASC $ABC$ ;comment`, "{data/b}", "c1c2c3"},
		{ra, ` ASC $ABC`, "{data/b}", "c1c2c3"},
		{ra, ` ASC -ABC-`, "{data/b}", "c1c2c3"},
		{ra, ` DCI "ABC"`, "{data/b}", "4142c3"},
		{ra, ` SBTL Title here`, "{-}", ""},
		{ra, ` TITLE Title here`, "{-}", ""},
		{rb, " ROL   Comment after three spaces", "{ROL/a}", "2a"},
		{rb, " ROL   X", "{ROL/a}", "2a"}, // two spaces = comment
		{rb, " ROL", "{ROL/a}", "2a"},
		{ss, "                                        far-out-comment", "{-}", ""},
		{ss, " .BS $8", "{block $0008}", "xxxxxxxxxxxxxxxx"},
		{ss, " .DA $1234", "{data $1234}", "3412"},
		{ss, " .DA/$1234,#$1234,$1234", "{data (msb $1234),(lsb $1234),$1234}", "12343412"},
		{ss, " .DO A<$3", "{if (< A $0003)}", ""},
		{ss, " .ELSE", "{else}", ""},
		{ss, " .EM", "{endm}", ""},
		{ss, " .EN", "{end}", ""},
		{ss, " .FIN", "{endif}", ""},
		{ss, " .HS 0001FFAB", "{data/b}", "0001ffab"},
		{ss, " .IN FILE.NAME", "{inc 'FILE.NAME'}", ""},
		{ss, " .IN S.DEFS", "{inc 'S.DEFS'}", ""},
		{ss, " .MA MacroName", `{macro "MacroName"}`, ""},
		{ss, " .OR $D000", "{org $d000}", ""},
		{ss, " .TF OUT.BIN", "{-}", ""},
		{ss, " .TI 76,Title here", "{-}", ""},
		{ss, " BEQ $2343", "{BEQ/rel $2343}", "f0fc"},
		{ss, " BEQ $2345", "{BEQ/rel $2345}", "f0fe"},
		{ss, " BEQ $2347", "{BEQ/rel $2347}", "f000"},
		{ss, " CMP #';'+1", "{CMP/imm (lsb (+ $003b $0001))}", "c93c"},
		{ss, " JMP $1234", "{JMP/abs $1234}", "4c3412"},
		{ss, " JMP ($1234)", "{JMP/ind $1234}", "6c3412"},
		{ss, " LDA #$12", "{LDA/imm (lsb $0012)}", "a912"},
		{ss, " LDA $12", "{LDA/zp $0012}", "a512"},
		{ss, " LDA $12,X", "{LDA/zpX $0012}", "b512"},
		{ss, " LDA $1234", "{LDA/abs $1234}", "ad3412"},
		{ss, " LDA $1234,X", "{LDA/absX $1234}", "bd3412"},
		{ss, " LDA ($12),Y", "{LDA/indY $0012}", "b112"},
		{ss, " LDA ($12,X)", "{LDA/indX $0012}", "a112"},
		{ss, " LDX #']+$80", "{LDX/imm (lsb (+ $005d $0080))}", "a2dd"},
		{ss, " LDX $12,Y", "{LDX/zpY $0012}", "b612"},
		{ss, " ROL  Comment after two spaces", "{ROL/a}", "2a"},
		{ss, " ROL  X", "{ROL/a}", "2a"}, // two spaces = comment
		{ss, " ROL $12", "{ROL/zp $0012}", "2612"},
		{ss, " ROL $1234", "{ROL/abs $1234}", "2e3412"},
		{ss, " ROL", "{ROL/a}", "2a"},
		{ss, " STA $1234,Y", "{STA/absY $1234}", "993412"},
		{ss, "* Comment", "{-}", ""},
		{ss, "A.B .EQ *-C.D", "{= 'A.B' (- * C.D)}", ""},
		{ss, "Label", "{- 'Label'}", ""},
		{ss, ` .AS "ABC"`, "{data/b}", "414243"},
		{ss, ` .AS -"ABC"`, "{data/b}", "c1c2c3"},
		{ss, ` .AS -DABCD`, "{data/b}", "c1c2c3"},
		{ss, ` .AS /ABC/`, "{data/b}", "414243"},
		{ss, ` .AT "ABC"`, "{data/b}", "4142c3"},
		{ss, ` .AT -"ABC"`, "{data/b}", "c1c243"},
		{ss, ` .AT -DABCD`, "{data/b}", "c1c243"},
		{ss, ` .AT /ABC/`, "{data/b}", "4142c3"},
		{ss, `>SAM AB,$12,"A B","A, B, "" C"`, `{call SAM {"AB", "$12", "A B", "A, B, \" C"}}`, ""},
		// {ss, " LDA #3/0", "{LDA/imm (lsb (/ $0003 $0000))}", "a9ff"},
	}

	// TODO(zellyn): Add tests for finalization of four SCMA directives:
	// "Labels used in operand expressions after .OR, TA, .BS,
	//  and .EQ directives must be defined prior to use (to prevent an
	//  undefined or ambiguous location counter)."

	for i, tt := range tests {
		// TODO(zellyn): Test AS65 and Merlin too.

		// Initialize to a known state for testing.
		ctx := &context.SimpleContext{}
		ctx.Clear()
		ctx.SetAddr(0x2345)
		ctx.Set("A.B", 0x6789)
		ctx.Set("C.D", 0x789a)
		ctx.Set("L2", 0x6789)
		ctx.Set("L3", 0x789a)
		ctx.AddMacroName("INCW")
		ctx.AddMacroName("M1")
		tt.f.InitContext(ctx)

		in, err := tt.f.ParseInstr(ctx, lines.NewSimple(tt.i), flavors.ParseModeNormal)
		if err != nil {
			t.Errorf(`%d. %s.ParseInstr("%s") => error: %s`, i, tt.f, tt.i, err)
			continue
		}
		if in.Line.Parse == nil {
			t.Errorf("Got empty in.Line.Parse on input '%s'", tt.i)
		}

		ctx.Set("L4", 0xbcde)
		ctx.Set("L5", 0xcdef)

		if in.Type != inst.TypeOrg {
			if err = in.Compute(ctx); err != nil {
				t.Errorf(`%d. %s.ParseInstr("%s"): %s.Compute(tt.f) => error: %s`, i, tt.f, tt.i, in, err)
				continue
			}
		}
		if in.String() != tt.p {
			t.Errorf(`%d. %s.ParseInstr("%s") = %s; want %s`, i, tt.f, tt.i, in.String(), tt.p)
			continue
		}

		if tt.b != "?" {
			hx := hex.EncodeToString(in.Data)
			// xxxxxx sets the width, but doesn't expect actual data
			if hx != tt.b && (len(tt.b) == 0 || tt.b[0] != 'x') {
				t.Errorf(`%d. %s.ParseInstr("%s").Data = [%s]; want [%s] (%s)`, i, tt.f, tt.i, hx, tt.b, in)
				continue
			}

			// Check length
			w := uint16(len(tt.b) / 2)

			if in.Width != w {
				t.Errorf(`%d. %s.Width=%d; want %d`, i, in, in.Width, w)
				continue
			}
		}
	}
}

func TestSimpleErrors(t *testing.T) {
	ss := scma.New()
	aa := as65.New()
	mm := merlin.New()

	tests := []struct {
		f flavors.F // assembler flavor
		i string    // input string
	}{

		{ss, " LDA"},            // missing arg
		{aa, " lda"},            //
		{mm, " LDA"},            //
		{aa, " rol"},            // missing arg (for assemblers that need "A")
		{ss, " .DA $1234,"},     // data: trailing comma
		{ss, `>MACRO "ABC`},     // macro: unclosed quote on arg
		{ss, `>MACRO "ABC"$12`}, // macro: stuff after closing quote
	}

	for i, tt := range tests {
		if tt.f != ss {
			continue
		}
		ctx := &context.SimpleContext{}
		in, err := tt.f.ParseInstr(ctx, lines.NewSimple(tt.i), flavors.ParseModeNormal)
		if err == nil {
			t.Errorf(`%d. %s.ParseInstr("%s") want err; got %s`, i, tt.f, tt.i, in)
			continue
		}
	}
}
