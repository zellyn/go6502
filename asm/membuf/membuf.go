package membuf

import "fmt"

type Membuf struct {
	// 0 means unused
	// 1-257 means byte+1
	data []int16
}

type Piece struct {
	Addr int
	Data []byte
}

func (m *Membuf) Delete(addr, size int) {
	l := len(m.data)
	for i := 0; i < size; i++ {
		a := addr + i
		if a >= l {
			return
		}
		m.data[a] = 0
	}
}

func (m *Membuf) Write(addr int, bs []byte) {
	if addr < 0 {
		panic(fmt.Sprintf("addr < 0: %d", addr))
	}
	if len(bs) == 0 {
		return
	}
	e := addr + len(bs)
	if len(m.data) < e {
		m.data = append(m.data, make([]int16, e-len(m.data))...)
	}
	for i, b := range bs {
		m.data[addr+i] = int16(b) + 1
	}
}

func (m *Membuf) Pieces() []Piece {
	ps := []Piece{}
	var p *Piece
	for a, d := range m.data {
		if d == 0 {
			if p != nil {
				ps = append(ps, *p)
				p = nil
			}
		} else {
			if p == nil {
				p = &Piece{Addr: a}
			}
			p.Data = append(p.Data, byte(d-1))
		}
	}
	if p != nil {
		ps = append(ps, *p)
	}
	return ps
}

func (m *Membuf) Piece(fill byte) Piece {
	var p Piece
	started := false
	for a, d := range m.data {
		if d == 0 {
			if started {
				p.Data = append(p.Data, fill)
			}
		} else {
			if !started {
				started = true
				p.Addr = a
			}
			p.Data = append(p.Data, byte(d-1))
		}
	}
	return p
}
