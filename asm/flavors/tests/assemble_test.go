package flavors

import (
	"encoding/hex"
	"reflect"
	"strings"
	"testing"

	"github.com/zellyn/go6502/asm"
	"github.com/zellyn/go6502/asm/flavors/merlin"
	"github.com/zellyn/go6502/asm/flavors/redbook"
	"github.com/zellyn/go6502/asm/flavors/scma"
	"github.com/zellyn/go6502/asm/lines"
	"github.com/zellyn/go6502/asm/membuf"
	"github.com/zellyn/go6502/opcodes"
)

// h converts from hex or panics.
func h(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return b
}

type asmFactory func() *asm.Assembler

func TestMultiline(t *testing.T) {
	o := lines.NewTestOpener()

	ss := asmFactory(func() *asm.Assembler {
		return asm.NewAssembler(scma.New(opcodes.SetSweet16), o)
	})
	ra := asmFactory(func() *asm.Assembler {
		return asm.NewAssembler(redbook.NewRedbookA(opcodes.SetSweet16), o)
	})
	mm := asmFactory(func() *asm.Assembler {
		return asm.NewAssembler(merlin.New(opcodes.SetSweet16), o)
	})

	tests := []struct {
		af     asmFactory          // assembler factory
		name   string              // name
		i      []string            // main file: lines
		ii     map[string][]string // other files: lines
		b      string              // bytes, expected
		ps     []membuf.Piece      // output bytes expected
		active bool
	}{
		// We cannot determine L2, so we go wide.
		{ss, "Unknown label: wide", []string{
			"L1 LDA L2-L1",
			"L2 NOP",
		}, nil, "ad0300ea", nil, true},

		// sc-asm sets instruction widths on the first pass
		{ss, "Later label: wide", []string{
			" LDA FOO",
			"FOO .EQ $FF",
			" NOP",
		}, nil, "adff00ea", nil, true},

		// Sub-labels
		{ss, "ss:Sublabels", []string{
			"L1 BEQ .1",
			".1 NOP",
			"L2 BEQ .2",
			".1 NOP",
			".2 NOP",
		}, nil, "f000eaf001eaea", nil, true},

		// Sub-labels
		{mm, "mm:Sublabels", []string{
			"L1   BEQ :ONE",
			":ONE NOP",
			"L2   BEQ :TWO",
			":ONE NOP",
			":TWO NOP",
		}, nil, "f000eaf001eaea", nil, true},

		// Includes: one level deep
		{ss, "Include A", []string{
			" BEQ OVER",
			" .IN SUBFILE1",
			"OVER NOP",
		}, map[string][]string{
			"SUBFILE1": {
				" LDA #$2a",
			},
		}, "f002a92aea", nil, true},

		// Ifdefs: simple at first
		{ss, "Ifdef A", []string{
			"L1 .EQ $17",
			" .DO L1>$16",
			" LDA #$01",
			" .ELSE",
			" LDA #$02",
			" .FIN",
			" NOP",
		}, nil, "a901ea", nil, true},

		// Ifdefs: else part
		{ss, "Ifdef B", []string{
			"L1 .EQ $16",
			" .DO L1>$16",
			" LDA #$01",
			" .ELSE",
			" LDA #$02",
			" .FIN",
			" NOP",
		}, nil, "a902ea", nil, true},

		// Ifdefs: multiple else, true
		{ss, "Ifdef C", []string{
			"L1 .EQ $17",
			" .DO L1>$16",
			" LDA #$01",
			" .ELSE",
			" LDA #$02",
			" .ELSE",
			" LDA #$03",
			" .ELSE",
			" LDA #$04",
			" .FIN",
			" NOP",
		}, nil, "a901a903ea", nil, true},

		// Ifdefs: multiple else, false
		{ss, "Ifdef D", []string{
			"L1 .EQ $16",
			" .DO L1>$16",
			" LDA #$01",
			" .ELSE",
			" LDA #$02",
			" .ELSE",
			" LDA #$03",
			" .ELSE",
			" LDA #$04",
			" .FIN",
			" NOP",
		}, nil, "a902a904ea", nil, true},

		// Ifdef based on org, true
		{ss, "Ifdef/org A", []string{
			" .OR $1000",
			" LDA #$01",
			" .DO *=$1002",
			" LDA #$02",
			" .ELSE",
			" LDA #$03",
			" .FIN",
			" NOP",
		}, nil, "a901a902ea", nil, true},

		// Ifdef based on org, false
		{ss, "Ifdef/org B", []string{
			" .OR $1000",
			" LDA #$01",
			" .DO *=$1003",
			" LDA #$02",
			" .ELSE",
			" LDA #$03",
			" .FIN",
			" NOP",
		}, nil, "a901a903ea", nil, true},

		// Macro, simple
		{ss, "Macro, simple", []string{
			"       .MA INCD    MACRO NAME",
			"       INC ]1      CALL PARAMETER",
			"       BNE :1      PRIVATE LABEL",
			"       INC ]1+1",
			":1",
			"       .EM         END OF DEFINITION",
			"       >INCD PTR",
			"PTR .HS 0000",
			"ZPTR .EQ $42",
			"       >INCD ZPTR",
		}, nil, "ee0808d003ee09080000e642d002e643", nil, true},

		// Macro, conditional assembly
		{ss, "Macro, conditional assembly", []string{
			" *--------------------------------",
			" *      DEMONSTRATE CONDITIONAL ASSEMBLY IN",
			" *--------------------------------   MACRO",
			"        .MA INCD",
			"        .DO ]#=2",
			"        INC ]1,]2",
			"        BNE :3",
			"        INC ]1+1,]2",
			":3",
			"        .ELSE",
			"        INC ]1",
			"        BNE :3",
			"        INC ]1+1",
			":3",
			"        .FIN",
			"        .EM",
			" *--------------------------------",
			"         >INCD $12",
			"         >INCD $1234",
			"         >INCD $12,X",
			"         >INCD $1234,X",
		}, nil, "e612d002e613ee3412d003ee3512f612d002f613fe3412d003fe3512", nil, true},

		// Macros, nested
		{ss, "Macros, nested", []string{
			"        .MA CALL",
			"        JSR ]1",
			"        .DO ]#>1",
			"        JSR ]2",
			"        .FIN",
			"        .DO ]#>2",
			"        JSR ]3",
			"        .FIN",
			"        .EM",
			"        >CALL SAM,TOM,JOE",
			"        >CALL SAM,TOM",
			"SAM    RTS",
			"JOE    RTS",
			"TOM    RTS",
		}, nil, "200f08201108201008200f08201108606060", nil, true},

		// Check outputs when origin is changed.
		{ss, "Origin A", []string{
			" .OR $1000",
			" LDA #$01",
			" .OR $2000",
			" LDA #$02",
			" .OR $1002",
			" LDA #$03",
			" NOP",
		}, nil, "a901a902a903ea", []membuf.Piece{
			{0x1000, h("a901a903ea")},
			{0x2000, h("a902")},
		}, true},

		// Check turning MSB on and off
		{ra, "MSB toggle", []string{
			" ASC 'AB'",
			" MSB OFF",
			" ASC 'AB'",
			" MSB ON",
			" ASC 'AB'",
		}, nil, "c1c24142c1c2", nil, true},

		// Merlin: macros, local labels
		{mm, "Macros: local labels", []string{
			"L1  NOP",
			"M1  MAC",
			"    INC ]1",
			"    BNE L1",
			"L1  NOP",
			"    <<<",
			"    >>> M1.$42",
			"    PMC M1($43;$44",
			"    M1 $44",
		}, nil, "eae642d000eae643d000eae644d000ea", nil, true},

		// Merlin: macros in other files
		{mm, "macro in include", []string{
			" ORG $E000",
			" USE MACROS",
			"PROMPT = $0033",
			" INCW PROMPT",
			" NOP",
		}, map[string][]string{
			"T.MACROS": {
				"* Increment word",
				"INCW     MAC",
				"         INC ]1",
				"         BNE INCW_END",
				"         INC ]1+1",
				"INCW_END EOM",
			},
		}, "e633d002e634ea", nil, true},

		// Merlin: macros override the 3+1 zero-page rule
		{mm, "macro vs zero page override", []string{
			" USE MACROS",
			"A1 = $0033",
			"A2 = $0035",
			" CMPW A1;A2",
			" NOP",
		}, map[string][]string{
			"T.MACROS": {
				"* Compare word",
				"CMPW     MAC",
				"         LDA ]1",
				"         CMP ]2",
				"         LDA ]1+1",
				"         SBC ]2+1",
				"         EOM",
			},
		}, "a533c535a534e536ea", nil, true},

		// Merlin: DB with unknown addresses
		{mm, "DB with unknown addresses", []string{
			"ONE = $0123",
			"TWO = $4567",
			" DB <ONE,<$FFFF,<$FFFF,<TWO",
			" DB <THREE,<FOUR",
			"THREE = $89ab",
			"FOUR = $cdef",
			" NOP",
		}, nil, "23ffff67abefea", nil, true},
	}

	for i, tt := range tests {
		if !tt.active {
			continue
		}
		a := tt.af()
		if tt.b == "" && len(tt.ps) == 0 {
			t.Fatalf(`%d("%s" - %s): test case must specify bytes or pieces`, i, tt.name, a.Flavor)
		}
		o.Clear()
		o["TESTFILE"] = strings.Join(tt.i, "\n")
		for k, v := range tt.ii {
			o[k] = strings.Join(v, "\n")
		}
		if err := a.Load("TESTFILE", 0); err != nil {
			t.Fatalf(`%d("%s" - %s): a.Load("TESTFILE") failed: %s`, i, tt.name, a.Flavor, err)
		}
		err := a.Pass2()
		if err != nil {
			t.Fatalf(`%d("%s" - %s): a.Pass(true) failed: %s`, i, tt.name, a.Flavor, err)
		}

		if tt.b != "" {
			bb, err := a.RawBytes()
			if err != nil {
				t.Fatalf(`%d("%s" - %s): a.RawBytes() failed: %s`, i, tt.name, a.Flavor, err)
			}
			hx := hex.EncodeToString(bb)
			if hx != tt.b {
				t.Fatalf(`%d("%s" - %s): a.RawBytes()=[%s]; want [%s]`, i, tt.name, a.Flavor, hx, tt.b)
			}
		}
		if len(tt.ps) != 0 {
			m, err := a.Membuf()
			if err != nil {
				t.Fatalf(`%d("%s" - %s): a.Membuf() failed: %s`, i, tt.name, a.Flavor, err)
			}
			ps := m.Pieces()
			if !reflect.DeepEqual(ps, tt.ps) {
				t.Fatalf(`%d("%s" - %s): tt.Membuf().Pieces() = %v; want %v`, i, tt.name, a.Flavor, ps, tt.ps)
			}
		}
	}
}
