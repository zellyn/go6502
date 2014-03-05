package lines

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

type Opener interface {
	Open(filename string) (io.ReadCloser, error)
}

type OsOpener struct{}

func (o OsOpener) Open(filename string) (io.ReadCloser, error) {
	return os.Open(filename)
}

type FileLineSource struct {
	context Context
	lines   []string
	size    int
	curr    int
}

func (fls *FileLineSource) Next() (line Line, done bool, err error) {
	if fls.curr >= fls.size {
		return Line{}, true, nil
	}
	fls.curr++
	return NewLine(fls.lines[fls.curr-1], fls.curr, &fls.context), false, nil
}

func (fls FileLineSource) Context() Context {
	return fls.context
}

func NewFileLineSource(filename string, context Context, opener Opener) (LineSource, error) {
	fls := &FileLineSource{context: context}
	file, err := opener.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fls.lines = append(fls.lines, scanner.Text())
	}

	if err = scanner.Err(); err != nil {
		return nil, err
	}

	fls.size = len(fls.lines)
	return fls, nil
}

type TestOpener map[string]string

func (o TestOpener) Open(filename string) (io.ReadCloser, error) {
	contents, ok := o[filename]
	if ok {
		return ioutil.NopCloser(strings.NewReader(contents)), nil
	} else {
		return nil, fmt.Errorf("File not found: %s", filename)
	}
}

func (o TestOpener) Clear() {
	for k := range o {
		delete(o, k)
	}
}

func NewTestOpener() TestOpener {
	return make(TestOpener)
}
