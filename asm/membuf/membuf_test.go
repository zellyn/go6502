package membuf

import (
	"reflect"
	"testing"
)

func TestSimpleAdd(t *testing.T) {
	m := Membuf{}
	m.Write(0x100, []byte("abcde"))
	got := m.Pieces()
	want := []Piece{
		{0x100, []byte("abcde")},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("m.Pieces()=%v; want %v", got, want)
	}
}

func TestMultiAdd(t *testing.T) {
	m := Membuf{}
	m.Write(0x100, []byte("abcde"))
	m.Write(0x108, []byte("fghij"))
	got := m.Pieces()
	want := []Piece{
		{0x100, []byte("abcde")},
		{0x108, []byte("fghij")},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("m.Pieces()=%v; want %v", got, want)
	}
}

func TestOverlappingAdd(t *testing.T) {
	m := Membuf{}
	m.Write(0x100, []byte("abcde"))
	m.Write(0x108, []byte("fghij"))
	m.Write(0x104, []byte("klmno"))
	got := m.Pieces()
	want := []Piece{
		{0x100, []byte("abcdklmnoghij")},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("m.Pieces()=%v; want %v", got, want)
	}
}

func TestDelete(t *testing.T) {
	m := Membuf{}
	m.Write(0x100, []byte("abcdefghijklmnop"))

	// Delete one off the end
	m.Delete(0x10f, 1)
	got := m.Pieces()
	want := []Piece{
		{0x100, []byte("abcdefghijklmno")},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("m.Pieces()=%v; want %v", got, want)
	}

	// Delete a section overlapping the end
	m.Delete(0x10e, 10)
	got = m.Pieces()
	want = []Piece{
		{0x100, []byte("abcdefghijklmn")},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("m.Pieces()=%v; want %v", got, want)
	}

	// Delete unfilled data: no change
	m.Delete(0x1000, 10)
	got = m.Pieces()
	want = []Piece{
		{0x100, []byte("abcdefghijklmn")},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("m.Pieces()=%v; want %v", got, want)
	}

	// Delete one off the start
	m.Delete(0x100, 1)
	got = m.Pieces()
	want = []Piece{
		{0x101, []byte("bcdefghijklmn")},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("m.Pieces()=%v; want %v", got, want)
	}

	// Delete a section overlapping the start
	m.Delete(0xfe, 4)
	got = m.Pieces()
	want = []Piece{
		{0x102, []byte("cdefghijklmn")},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("m.Pieces()=%v; want %v", got, want)
	}

	// Delete a bit in the middle
	m.Delete(0x106, 4)
	got = m.Pieces()
	want = []Piece{
		{0x102, []byte("cdef")},
		{0x10a, []byte("klmn")},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("m.Pieces()=%v; want %v", got, want)
	}

	// Delete a zero-length chunk: nop
	m.Delete(0x103, 0)
	got = m.Pieces()
	want = []Piece{
		{0x102, []byte("cdef")},
		{0x10a, []byte("klmn")},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("m.Pieces()=%v; want %v", got, want)
	}

	// Delete everything
	m.Delete(0x0, 0x1000)
	got = m.Pieces()
	if len(got) != 0 {
		t.Fatalf("m.Pieces()=%v; want nil or {}", got, want)
	}
}
