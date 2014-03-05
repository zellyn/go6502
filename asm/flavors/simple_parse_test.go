package flavors

import (
	"encoding/hex"
	"testing"

	"github.com/zellyn/go6502/asm"
	"github.com/zellyn/go6502/asm/flavors/as65"
	"github.com/zellyn/go6502/asm/flavors/merlin"
	"github.com/zellyn/go6502/asm/flavors/scma"
	"github.com/zellyn/go6502/asm/lines"
)

func TestSimpleCommonFunctions(t *testing.T) {
	ss := scma.New()
	aa := as65.New()
	mm := merlin.New()

	tests := []struct {
		a asm.Flavor // assembler flavor
		i string     // input string
		p string     // printed instruction, expected
		b string     // bytes, expected
	}{
		{ss, "* Comment", "{-}", ""},
		{ss, "* Comment", "{-}", ""},
		{aa, "; Comment", "{-}", ""},
		{mm, "* Comment", "{-}", ""},
		{ss, "Label", "{- 'Label'}", ""},
		{aa, "Label", "{- 'Label'}", ""},
		{mm, "Label", "{- 'Label'}", ""},
		{ss, " .IN FILE.NAME", "{inc 'FILE.NAME'}", ""},
		{ss, " .IN S.DEFS", "{inc 'S.DEFS'}", ""},
		{aa, ` include "FILE.NAME"`, "{inc 'FILE.NAME'}", ""},
		{mm, " PUT !FILE.NAME", "{inc 'FILE.NAME'}", ""},
		{ss, " .TI 76,Title here", "{-}", ""},
		{aa, ` title "Title here"`, "{-}", ""},
		{mm, ` TTL "Title here"`, "{-}", ""},
		{ss, " .TF OUT.BIN", "{-}", ""},
		{mm, " DSK OUTFILE", "{-}", ""},
		{mm, " SAV OUTFILE", "{-}", ""},
		{ss, " .OR $D000", "{org $d000}", ""},
		{aa, " org $D000", "{org $d000}", ""},
		{mm, " ORG $D000", "{org $d000}", ""},
		// {ss, " .TA *-1234", "{target (- * $04d2)}", ""},
		{ss, " .DA $1234", "{data $1234}", "3412"},
		{aa, " dw $1234", "{data $1234}", "3412"},
		{mm, " DW $1234", "{data $1234}", "3412"},
		{ss, " .DA/$1234,#$1234,$1234", "{data (msb $1234),(lsb $1234),$1234}", "12343412"},
		{ss, " ROL", "{ROL/a}", "2a"},
		{aa, " rol a", "{ROL/a}", "2a"},
		{mm, " ROL", "{ROL/a}", "2a"},
		{ss, " ROL  Comment after two spaces", "{ROL/a}", "2a"},
		{ss, " ROL $1234", "{ROL/abs $1234}", "2e3412"},
		{aa, " rol $1234", "{ROL/abs $1234}", "2e3412"},
		{mm, " ROL $1234", "{ROL/abs $1234}", "2e3412"},
		{ss, " ROL $12", "{ROL/zp $0012}", "2612"},
		{aa, " rol $12", "{ROL/zp $0012}", "2612"},
		{mm, " ROL $12", "{ROL/zp $0012}", "2612"},
		{ss, " LDA #$12", "{LDA/imm (lsb $0012)}", "a912"},
		{aa, " lda #$12", "{LDA/imm (lsb $0012)}", "a912"},
		{mm, " LDA #$12", "{LDA/imm (lsb $0012)}", "a912"},
		{ss, " JMP $1234", "{JMP/abs $1234}", "4c3412"},
		{aa, " jmp $1234", "{JMP/abs $1234}", "4c3412"},
		{mm, " JMP $1234", "{JMP/abs $1234}", "4c3412"},
		{ss, " JMP ($1234)", "{JMP/ind $1234}", "6c3412"},
		{aa, " jmp ($1234)", "{JMP/ind $1234}", "6c3412"},
		{mm, " JMP ($1234)", "{JMP/ind $1234}", "6c3412"},
		{ss, " BEQ $2345", "{BEQ/rel $2345}", "f0fe"},
		{aa, " beq $2345", "{BEQ/rel $2345}", "f0fe"},
		{mm, " BEQ $2345", "{BEQ/rel $2345}", "f0fe"},
		{ss, " BEQ $2347", "{BEQ/rel $2347}", "f000"},
		{aa, " beq $2347", "{BEQ/rel $2347}", "f000"},
		{mm, " BEQ $2347", "{BEQ/rel $2347}", "f000"},
		{ss, " BEQ $2343", "{BEQ/rel $2343}", "f0fc"},
		{aa, " beq $2343", "{BEQ/rel $2343}", "f0fc"},
		{mm, " BEQ $2343", "{BEQ/rel $2343}", "f0fc"},
		{ss, " LDA $1234", "{LDA/abs $1234}", "ad3412"},
		{aa, " lda $1234", "{LDA/abs $1234}", "ad3412"},
		{mm, " LDA $1234", "{LDA/abs $1234}", "ad3412"},
		{ss, " LDA $1234,X", "{LDA/absX $1234}", "bd3412"},
		{aa, " lda $1234,x", "{LDA/absX $1234}", "bd3412"},
		{mm, " LDA $1234,X", "{LDA/absX $1234}", "bd3412"},
		{ss, " STA $1234,Y", "{STA/absY $1234}", "993412"},
		{aa, " sta $1234,y", "{STA/absY $1234}", "993412"},
		{mm, " STA $1234,Y", "{STA/absY $1234}", "993412"},
		{ss, " LDA $12", "{LDA/zp $0012}", "a512"},
		{aa, " lda $12", "{LDA/zp $0012}", "a512"},
		{mm, " LDA $12", "{LDA/zp $0012}", "a512"},
		{ss, " LDA $12,X", "{LDA/zpX $0012}", "b512"},
		{aa, " lda $12,x", "{LDA/zpX $0012}", "b512"},
		{mm, " LDA $12,X", "{LDA/zpX $0012}", "b512"},
		{ss, " LDX $12,Y", "{LDX/zpY $0012}", "b612"},
		{aa, " ldx $12,y", "{LDX/zpY $0012}", "b612"},
		{mm, " LDX $12,Y", "{LDX/zpY $0012}", "b612"},
		{ss, " LDA ($12),Y", "{LDA/indY $0012}", "b112"},
		{aa, " lda ($12),y", "{LDA/indY $0012}", "b112"},
		{mm, " LDA ($12),Y", "{LDA/indY $0012}", "b112"},
		{ss, " LDA ($12,X)", "{LDA/indX $0012}", "a112"},
		{aa, " lda ($12,x)", "{LDA/indX $0012}", "a112"},
		{mm, " LDA ($12,X)", "{LDA/indX $0012}", "a112"},
		{ss, ` .AS "ABC"`, "{data}", "414243"},
		{ss, ` .AT "ABC"`, "{data}", "4142c3"},
		{ss, ` .AS /ABC/`, "{data}", "414243"},
		{ss, ` .AT /ABC/`, "{data}", "4142c3"},
		{ss, ` .AS -"ABC"`, "{data}", "c1c2c3"},
		{ss, ` .AT -"ABC"`, "{data}", "c1c243"},
		{ss, ` .AS -dABCd`, "{data}", "c1c2c3"},
		{ss, ` .AT -dABCd`, "{data}", "c1c243"},
		{ss, " .HS 0001ffAb", "{data}", "0001ffab"},
		{ss, "A.B .EQ *-C.D", "{= 'A.B' (- * C.D)}", ""},
		{ss, " .BS $8", "{block $0008}", ""},
		{ss, " .DO A<$3", "{if (< A $0003)}", ""},
		{ss, " .ELSE", "{else}", ""},
		{ss, " .FIN", "{endif}", ""},
		{ss, "Label .MA", "{macro 'Label'}", ""},
		{ss, " .EM", "{endm}", ""},
		{ss, " .EN", "{end}", ""},
		{ss, `>SAM AB,$12,"A B","A, B, "" C"`,
			`{call SAM {"AB", "$12", "A B", "A, B, \" C"}}`, ""},
	}

	// TODO(zellyn): Add tests for finalization of four SCMA directives:
	// "Labels used in operand expressions after .OR, TA, .BS,
	//  and .EQ directives must be defined prior to use (to prevent an
	//  undefined or ambiguous location counter)."

	for i, tt := range tests {
		// TODO(zellyn): Test AS65 and Merlin too.
		if tt.a != ss {
			continue
		}

		// Initialize to a known state for testing.
		tt.a.Clear()
		tt.a.SetAddr(0x2345)
		tt.a.Set("A.B", 0x6789)
		tt.a.Set("C.D", 0x789a)

		inst, err := tt.a.ParseInstr(lines.NewSimple(tt.i))
		if err != nil {
			t.Errorf(`%d. %T.ParseInstr("%s") => error: %s`, i, tt.a, tt.i, err)
			continue
		}
		if inst.Line.Parse == nil {
			t.Errorf("Got empty inst.Line.Parse on input '%s'", tt.i)
		}
		_, err = inst.Compute(tt.a, true, true)
		if err != nil {
			t.Errorf(`%d. %s.Compute(tt.a, true, true) => error: %s`, i, inst, err)
			continue
		}
		if inst.String() != tt.p {
			t.Errorf(`%d. %T.ParseInstr("%s") = %s; want %s`, i, tt.a, tt.i, inst.String(), tt.p)
			continue
		}

		if tt.b != "?" {
			hx := hex.EncodeToString(inst.Data)
			if hx != tt.b {
				t.Fatalf(`%d. %T.ParseInstr("%s").Data = [%s]; want [%s]`, i, tt.a, tt.i, hx, tt.b)
			}
		}
	}
}

func TestSimpleErrors(t *testing.T) {
	ss := scma.New()
	aa := as65.New()
	mm := merlin.New()

	tests := []struct {
		a asm.Flavor // assembler flavor
		i string     // input string
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
		// TODO(zellyn): Test AS65 and Merlin too.
		if tt.a != ss {
			continue
		}
		inst, err := tt.a.ParseInstr(lines.NewSimple(tt.i))
		if err == nil {
			t.Errorf(`%d. %T.ParseInstr("%s") want err; got %s`, i, tt.a, tt.i, inst)
			continue
		}
	}
}
