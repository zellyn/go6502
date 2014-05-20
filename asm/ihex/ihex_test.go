package ihex

import (
	"bytes"
	"testing"
)

func TestWrite(t *testing.T) {
	want := ":0501000061626364650b\n" +
		":00000001FF\n"
	var b bytes.Buffer
	w := NewWriter(&b)
	err := w.Write(0x100, []byte("abcde"))
	if err != nil {
		t.Fatal(err)
	}
	if err := w.End(); err != nil {
		t.Fatal(err)
	}
	got := string(b.Bytes())
	if got != want {
		t.Errorf("Got \n%s; want \n%s", got, want)
	}
}
