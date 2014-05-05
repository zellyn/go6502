package flavors

import (
	"encoding/hex"
	"os"
	"strings"
	"testing"

	"github.com/zellyn/go6502/asm"
	"github.com/zellyn/go6502/asm/flavors/scma"
	"github.com/zellyn/go6502/asm/lines"
)

func TestMultiline(t *testing.T) {
	o := lines.NewTestOpener()

	ss := asm.NewAssembler(scma.New(), o)
	// aa := asm.NewAssembler(as65.New(), o)
	// mm := asm.NewAssembler(merlin.New(), o)

	tests := []struct {
		a      *asm.Assembler      // assembler
		name   string              // name
		i      []string            // main file: lines
		ii     map[string][]string // other files: lines
		b      string              // bytes, expected
		active bool
	}{
		// We cannot determine L2, so we go wide.
		{ss, "Unknown label: wide", []string{
			"L1 LDA L2-L1",
			"L2 NOP",
		}, nil, "ad0300ea", false},

		// sc-asm sets instruction widths on the first pass
		{ss, "Later label: wide", []string{
			" LDA FOO",
			"FOO .EQ $FF",
			" NOP",
		}, nil, "adff00ea", true},

		// Sub-labels
		{ss, "Sublabels", []string{
			"L1 BEQ .1",
			".1 NOP",
			"L2 BEQ .2",
			".1 NOP",
			".2 NOP",
		}, nil, "f000eaf001eaea", false},

		// Includes: one level deep
		{ss, "Include A", []string{
			" BEQ OVER",
			" .IN SUBFILE1",
			"OVER NOP",
		}, map[string][]string{
			"SUBFILE1": {
				" LDA #$2a",
			},
		}, "f002a92aea", false},

		// Ifdefs: simple at first
		{ss, "Ifdef A", []string{
			"L1 .EQ $17",
			" .DO L1>$16",
			" LDA #$01",
			" .ELSE",
			" LDA #$02",
			" .FIN",
			" NOP",
		}, nil, "a901ea", true},

		// Ifdefs: else part
		{ss, "Ifdef B", []string{
			"L1 .EQ $16",
			" .DO L1>$16",
			" LDA #$01",
			" .ELSE",
			" LDA #$02",
			" .FIN",
			" NOP",
		}, nil, "a902ea", true},

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
		}, nil, "a901a903ea", true},

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
		}, nil, "a902a904ea", true},

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
		}, nil, "a901a902ea", true},

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
		}, nil, "a901a903ea", true},
	}

	for i, tt := range tests {
		if !tt.active {
			continue
		}
		tt.a.Reset()
		o.Clear()
		o["TESTFILE"] = strings.Join(tt.i, "\n")
		for k, v := range tt.ii {
			o[k] = strings.Join(v, "\n")
		}
		if err := tt.a.Load("TESTFILE"); err != nil {
			t.Errorf(`%d("%s"): tt.a.Load("TESTFILE") failed: %s`, i, tt.name, err)
			continue
		}
		if _, err := tt.a.Pass(true, false); err != nil {
			t.Errorf(`%d("%s"): tt.a.Pass(true, false) failed: %s`, i, tt.name, err)
			continue
		}
		isFinal, err := tt.a.Pass(true, true)
		if err != nil {
			t.Errorf(`%d("%s"): tt.a.Pass(true, true) failed: %s`, i, tt.name, err)
			continue
		}
		if !isFinal {
			t.Errorf(`%d("%s"): tt.a.Pass(true, true) couldn't finalize`, i, tt.name)
			continue
		}
		bb, err := tt.a.RawBytes()
		if err != nil {
			t.Errorf(`%d("%s"): tt.a.RawBytes() failed: %s`, i, tt.name, err)
			continue
		}
		hx := hex.EncodeToString(bb)
		if hx != tt.b {
			t.Errorf(`%d("%s"): tt.a.RawBytes()=[%s]; want [%s]`, i, tt.name, hx, tt.b)
			continue
		}
	}
}

func TestApplesoftBasic(t *testing.T) {
	if err := os.Chdir("../../../goapple2/source/applesoft"); err != nil {
		t.Fatal(err)
	}
	var o lines.OsOpener

	ss := asm.NewAssembler(scma.New(), o)
	ss.Reset()

	if err := ss.Load("S.ACF"); err != nil {
		t.Fatalf(`ss.Load("S.ACF") failed: %s`, err)
	}
	if _, err := ss.Pass(true, false); err != nil {
		t.Fatalf(`ss.Pass(true, false) failed: %s`, err)
	}
	isFinal, err := ss.Pass(true, true)
	if err != nil {
		t.Fatalf(`ss.Pass(true, true) failed: %s`, err)
	}
	if !isFinal {
		t.Fatalf(`ss.Pass(true, true) couldn't finalize`)
	}
	bb, err := ss.RawBytes()
	if err != nil {
		t.Fatalf(`ss.RawBytes() failed: %s`, err)
	}
	_ = bb
}
