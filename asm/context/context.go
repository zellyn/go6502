package context

import "fmt"

type Context interface {
	Set(name string, value uint16)
	Get(name string) (uint16, bool)
	SetAddr(uint16)
	ClearAddr(message string)
	ClearMesg() string
	GetAddr() (uint16, bool)
	Zero() (uint16, error) // type ZeroFunc
	RemoveChanged()
	AddrKnown() bool
	GetLastLabel() string
	SetLastLabel(string)
	Clear()
}

type SimpleContext struct {
	symbols   map[string]symbolValue
	addr      int32
	lastLabel string
	clearMesg string // Saved message describing why Addr was cleared.
}

type symbolValue struct {
	v       uint16
	changed bool // Did the value ever change?
}

func (sc *SimpleContext) fix() {
	if sc.symbols == nil {
		sc.symbols = make(map[string]symbolValue)
	}
}

func (sc *SimpleContext) Zero() (uint16, error) {
	return 0, fmt.Errorf("Not implemented: context.SimpleContext.Zero()")
}

func (sc *SimpleContext) Get(name string) (uint16, bool) {
	if name == "*" {
		return sc.GetAddr()
	}
	sc.fix()
	s, found := sc.symbols[name]
	return s.v, found
}

func (sc *SimpleContext) ClearAddr(message string) {
	sc.addr = -1
	sc.clearMesg = message
}

func (sc *SimpleContext) SetAddr(addr uint16) {
	sc.addr = int32(addr)
}

func (sc *SimpleContext) ClearMesg() string {
	return sc.clearMesg
}

func (sc *SimpleContext) GetAddr() (uint16, bool) {
	if sc.addr == -1 {
		return 0, false
	}
	return uint16(sc.addr), true
}

func (sc *SimpleContext) AddrKnown() bool {
	return sc.addr != -1
}

func (sc *SimpleContext) Set(name string, value uint16) {
	sc.fix()
	s, found := sc.symbols[name]
	if found && s.v != value {
		s.changed = true
	}
	s.v = value
	sc.symbols[name] = s
}

func (sc *SimpleContext) RemoveChanged() {
	sc.fix()
	for n, s := range sc.symbols {
		if s.changed {
			delete(sc.symbols, n)
		}
	}
}

func (sc *SimpleContext) GetLastLabel() string {
	return sc.lastLabel
}

func (sc *SimpleContext) SetLastLabel(l string) {
	sc.lastLabel = l
}

func (sc *SimpleContext) Clear() {
	sc.symbols = make(map[string]symbolValue)
}
