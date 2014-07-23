package lines

import "fmt"

type Context struct {
	Filename    string // Pointer to the filename
	Parent      *Line  // Pointer to parent line (eg. include, macro)
	MacroCall   int
	MacroLocals map[string]bool
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

func (c *Context) GetMacroCall() int {
	if c == nil {
		return 0
	}
	if c.MacroCall > 0 {
		return c.MacroCall
	}
	return c.Parent.GetMacroCall()
}

func (l *Line) GetMacroLocals() map[string]bool {
	if l == nil || l.Context == nil {
		return nil
	}
	return l.Context.MacroLocals
}

func (l *Line) GetMacroCall() int {
	if l == nil {
		return 0
	}
	return l.Context.GetMacroCall()
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

type SimpleLineSource struct {
	context Context
	lines   []string
	size    int
	curr    int
	prefix  int
}

func (sls *SimpleLineSource) Next() (line Line, done bool, err error) {
	if sls.curr >= sls.size {
		return Line{}, true, nil
	}
	sls.curr++
	l := NewLine(sls.lines[sls.curr-1], sls.curr, &sls.context)
	for i := 0; i < sls.prefix; i++ {
		l.Parse.Next()
		l.Parse.Ignore()
	}
	return l, false, nil
}

func (sls SimpleLineSource) Context() Context {
	return sls.context
}
func NewSimpleLineSource(context Context, ls []string, prefix int) LineSource {
	return &SimpleLineSource{
		context: context,
		lines:   ls,
		size:    len(ls),
		prefix:  prefix,
	}
}
