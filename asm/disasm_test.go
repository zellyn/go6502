package asm

import "testing"

func TestSymbolRE(t *testing.T) {
	tests := []struct {
		line    string
		label   string
		address string
	}{
		{"SHAPEL EQU   $1A POINTER TO", "SHAPEL", "1A"},
	}

	for i, tt := range tests {
		groups := symRe.FindStringSubmatch(tt.line)
		if groups == nil {
			t.Errorf(`%d. Unable to parse '%s'`, i, tt.line)
		}
		if groups[1] != tt.label {
			t.Errorf(`%d. want label='%s'; got '%s' for line: '%s'`, i, tt.label, groups[1], tt.line)
		}
		if groups[2] != tt.address {
			t.Errorf(`%d. want address='%s'; got '%s' for line: '%s'`, i, tt.address, groups[2], tt.line)
		}
	}
}
