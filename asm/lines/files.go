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

func NewFileLineSource(filename string, context Context, opener Opener) (LineSource, error) {
	var ls []string
	file, err := opener.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		ls = append(ls, scanner.Text())
	}

	if err = scanner.Err(); err != nil {
		return nil, err
	}

	return NewSimpleLineSource(context, ls), nil
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
