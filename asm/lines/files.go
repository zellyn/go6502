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

func NewFileLineSource(filename string, context Context, opener Opener, prefix int) (LineSource, error) {
	if prefix < 0 {
		return nil, fmt.Errorf("NewFileLineSource: want prefix >= 0; got %d", prefix)
	}
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

	return NewSimpleLineSource(context, ls, prefix), nil
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

var whitespace = " \t\f\r\n"
var prefixChars = "abcdefABCDEF0123456789:"

func linePrefix(line string) int {
	start := 0
	for i, c := range line {
		switch {
		case strings.IndexRune(whitespace, c) >= 0:
			start = i + 1
		case strings.IndexRune(prefixChars, c) < 0:
			return start
		}
	}
	return -1
}

// GuessFilePrefixSize takes an assembly listing with address, bytes,
// and possibly line numbers, and tries to guess the size of the
// prefix that will skip past all that, to the label column.
func GuessFilePrefixSize(filename string, opener Opener) (prefix int, err error) {
	counts := make(map[int]int)
	max := 0
	lines := 0
	rc, err := opener.Open(filename)
	if err != nil {
		return 0, err
	}
	defer func() {
		err = rc.Close()
	}()

	scanner := bufio.NewScanner(rc)
	for scanner.Scan() {
		lines++
		s := scanner.Text()
		prefix := linePrefix(s)
		if prefix >= 0 {
			counts[prefix] = counts[prefix] + 1
			if prefix > max {
				max = prefix
			}
		}
	}

	min := max
	for p, _ := range counts {
		if p < min {
			min = p
		}
	}

	return min, nil
}
