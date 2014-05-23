package lines

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

const Eol rune = '\n'

/*
Parse is a struct representing the parse/lex of a single line. It's
based on the lexer presented in Rob Pike's "Lexical Scannign in Go"
talk: http://cuddle.googlecode.com/hg/talk/lex.html, but it's simpler,
because assembly files are entirely line-based.
*/
type Parse struct {
	line  string
	start int
	pos   int
	width int
}

func NewParse(text string) *Parse {
	text = strings.TrimRightFunc(text, unicode.IsSpace)
	return &Parse{line: text}
}

func (lp *Parse) Emit() string {
	emitted := lp.line[lp.start:lp.pos]
	lp.start = lp.pos
	return emitted
}

func (lp *Parse) Ignore() {
	lp.start = lp.pos
}

func (lp *Parse) Next() (c rune) {
	if lp.pos >= len(lp.line) {
		lp.width = 0
		return Eol
	}
	c, lp.width = utf8.DecodeRuneInString(lp.line[lp.pos:])
	lp.pos += lp.width
	return c
}

func (lp *Parse) Backup() {
	lp.pos -= lp.width
}

func (lp *Parse) Peek() rune {
	c := lp.Next()
	lp.Backup()
	return c
}

func (lp *Parse) Accept(valid string) bool {
	if len(valid) == 0 {
		panic(`valid = ""`)
	}
	if strings.IndexRune(valid, lp.Next()) >= 0 {
		return true
	}
	lp.Backup()
	return false
}

// Consume accepts and ignores a single character of input.
func (lp *Parse) Consume(valid string) bool {
	if len(valid) == 0 {
		panic(`valid = ""`)
	}
	if strings.IndexRune(valid, lp.Next()) >= 0 {
		lp.start = lp.pos
		return true
	}
	lp.Backup()
	return false
}

func (lp *Parse) AcceptRun(valid string) bool {
	if len(valid) == 0 {
		panic(`valid = ""`)
	}
	some := false
	for strings.IndexRune(valid, lp.Next()) >= 0 {
		some = true
	}
	lp.Backup()
	return some
}

func (lp *Parse) AcceptUntil(until string) bool {
	until += "\n"
	some := false
	for strings.IndexRune(until, lp.Next()) < 0 {
		some = true
	}
	lp.Backup()
	return some
}

func (lp *Parse) IgnoreRun(valid string) bool {
	if len(valid) == 0 {
		panic(`valid = ""`)
	}
	if lp.AcceptRun(valid) {
		lp.Ignore()
		return true
	}
	return false
}

func (lp *Parse) AcceptString(prefix string) bool {
	if len(prefix) == 0 {
		panic(`prefix = ""`)
	}
	if strings.HasPrefix(lp.line[lp.pos:], prefix) {
		lp.pos += len(prefix)
		return true
	}
	return false
}

func (lp *Parse) Rest() string {
	return lp.line[lp.pos:]
}

func (lp *Parse) Text() string {
	return lp.line
}
