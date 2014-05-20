/* Package ihex implements enough of reading and writing Intel HEX files
to suffice for our purposes. */
package ihex

import (
	"encoding/hex"
	"fmt"
	"io"
)

type RecWidth int

const (
	Width16 RecWidth = 16
	Width32 RecWidth = 32
)

type RecType byte

const (
	TypeData RecType = 0 // Simple data
	TypeEof  RecType = 1 // End of file
	TypeELAR RecType = 4 // Extended Linear Address Record
)

// A Chunk represents a run of bytes, starting at a specific address.
type Chunk struct {
	Addr uint32
	Data []byte
}

// A Writer writes bytes at addresses to an Intel HEX formatted file.
//
// As returned by NewWriter, a Writer writes records of 32 bytes,
// terminated by a newline. The exported fields can be changed to
// customize the details before the first call to Write.
type Writer struct {
	Width   RecWidth
	UseCRLF bool // True to use \r\n as the line terminator
	w       io.Writer
	extAddr uint32
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		Width: Width32,
		w:     w,
	}
}

func (w *Writer) eol() string {
	if w.UseCRLF {
		return "\r\n"
	} else {
		return "\n"
	}
}

func (w *Writer) setExtendedAddress(addr uint32) error {
	addr = addr &^ 0xffff
	if w.extAddr == addr {
		return nil
	}
	return w.writeRecord(TypeELAR, 0, []byte{byte(addr >> 24), byte(addr >> 16)})
}

func (w *Writer) Write(addr uint32, data []byte) error {
	if err := w.setExtendedAddress(addr); err != nil {
		return err
	}
	addr = addr & 0xffff
	step := int(w.Width)
	for i := 0; i < len(data); i += step {
		end := i + step
		if end > len(data) {
			end = len(data)
		}
		if err := w.writeRecord(TypeData, addr+uint32(i), data[i:end]); err != nil {
			return err
		}
	}
	return nil
}

func (w *Writer) End() error {
	_, err := fmt.Fprintf(w.w, ":00000001FF%s", w.eol())
	return err
}

func checksum(data []byte) byte {
	var s byte
	for _, b := range data {
		s += b
	}
	return (s ^ 0xff) + 1
}

func (w *Writer) writeRecord(t RecType, addr uint32, b []byte) error {
	if len(b) > 0xff {
		panic(fmt.Sprintf("writeRecord called with >255 bytes: %d", len(b)))
	}
	data := []byte{byte(len(b)), byte(addr >> 8), byte(addr), byte(t)}
	data = append(data, b...)
	data = append(data, checksum(data))
	_, err := fmt.Fprintf(w.w, ":%s%s", hex.EncodeToString(data), w.eol())
	return err
}
