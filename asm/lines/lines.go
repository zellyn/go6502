package lines

import "fmt"

type Context struct {
	Filename string // Pointer to the filename
	Parent   *Line  // Pointer to parent line (eg. include, macro)
}

type Line struct {
	LineNo         int      // Actual line number in file
	DeclaredLineNo int      // Declared line number in file, or 0
	Context        *Context // Pointer to the file/include/macro context
	Parse          *Parse   // The lineparser
}

type LineSource interface {
	Next() (line Line, done bool, err error)
	Context() Context
}

func NewLine(s string, lineNo int, context *Context) Line {
	l := Line{
		LineNo:  lineNo,
		Context: context,
		Parse:   NewParse(s),
	}
	return l
}

var testFilename = "(test)"

func NewSimple(s string) Line {
	return NewLine(s, 0, &Context{Filename: testFilename})
}

func (l Line) Text() string {
	return l.Parse.Text()
}

func (l Line) String() string {
	return fmt.Sprintf("%s:%d: %s", l.Context.Filename, l.LineNo, l.Parse.Text())
}

func (l Line) Errorf(format string, a ...interface{}) error {
	filename := "unknown file"
	if l.Context != nil {
		filename = l.Context.Filename
	}
	return fmt.Errorf(fmt.Sprintf("%s:%d: %s", filename, l.LineNo, format), a...)
}

func (l Line) Sprintf(format string, a ...interface{}) string {
	filename := "unknown file"
	if l.Context != nil {
		filename = l.Context.Filename
	}
	return fmt.Sprintf(fmt.Sprintf("%s:%d: %s", filename, l.LineNo, format), a...)
}
