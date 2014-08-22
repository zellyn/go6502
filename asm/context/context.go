package context

import "fmt"

type Context interface {
	Set(name string, value uint16)
	Get(name string) (uint16, bool)
	SetAddr(uint16)
	GetAddr() (uint16, bool)
	DivZero() *uint16
	SetDivZero(uint16)
	RemoveChanged()
	Clear()
	SettingOn(name string) error
	SettingOff(name string) error
	Setting(name string) bool
	HasSetting(name string) bool
	AddMacroName(name string)
	HasMacroName(name string) bool
	PushMacroCall(name string, number int, locals map[string]bool)
	PopMacroCall() bool
	GetMacroCall() (string, int, map[string]bool)
	SetOnOffDefaults(map[string]bool)
	LastLabel() string
	SetLastLabel(label string)
}

type macroCall struct {
	name   string
	number int
	locals map[string]bool
}

type SimpleContext struct {
	symbols       map[string]symbolValue
	addr          int32
	lastLabel     string
	highbit       byte // OR-mask for ASCII high bit
	onOff         map[string]bool
	onOffDefaults map[string]bool
	macroNames    map[string]bool
	macroCalls    []macroCall
	divZeroVal    *uint16
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

func (sc *SimpleContext) Get(name string) (uint16, bool) {
	if name == "*" {
		return sc.GetAddr()
	}
	sc.fix()
	s, found := sc.symbols[name]
	return s.v, found
}

func (sc *SimpleContext) SetAddr(addr uint16) {
	sc.addr = int32(addr)
}

func (sc *SimpleContext) GetAddr() (uint16, bool) {
	if sc.addr == -1 {
		return 0, false
	}
	return uint16(sc.addr), true
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

func (sc *SimpleContext) Clear() {
	sc.symbols = make(map[string]symbolValue)
	sc.highbit = 0x00
	sc.macroNames = nil
	sc.resetOnOff()
}

func (sc *SimpleContext) SettingOn(name string) error {
	if !sc.HasSetting(name) {
		return fmt.Errorf("no settable variable named '%s'", name)
	}
	if sc.onOff == nil {
		sc.onOff = map[string]bool{name: true}
	} else {
		sc.onOff[name] = true
	}
	return nil
}

func (sc *SimpleContext) SettingOff(name string) error {
	if !sc.HasSetting(name) {
		return fmt.Errorf("no settable variable named '%s'", name)
	}
	if sc.onOff == nil {
		sc.onOff = map[string]bool{name: false}
	} else {
		sc.onOff[name] = false
	}
	return nil
}

func (sc *SimpleContext) Setting(name string) bool {
	return sc.onOff[name]
}

func (sc *SimpleContext) HasSetting(name string) bool {
	_, ok := sc.onOff[name]
	return ok
}

func (sc *SimpleContext) AddMacroName(name string) {
	if sc.macroNames == nil {
		sc.macroNames = make(map[string]bool)
	}
	sc.macroNames[name] = true
}

func (sc *SimpleContext) HasMacroName(name string) bool {
	return sc.macroNames[name]
}

func (sc *SimpleContext) resetOnOff() {
	sc.onOff = make(map[string]bool)
	for k, v := range sc.onOffDefaults {
		sc.onOff[k] = v
	}
}

func (sc *SimpleContext) SetOnOffDefaults(defaults map[string]bool) {
	sc.onOffDefaults = defaults
	sc.resetOnOff()
}

func (sc *SimpleContext) PushMacroCall(name string, number int, locals map[string]bool) {
	sc.macroCalls = append(sc.macroCalls, macroCall{name, number, locals})
}

func (sc *SimpleContext) PopMacroCall() bool {
	if len(sc.macroCalls) == 0 {
		return false
	}
	sc.macroCalls = sc.macroCalls[0 : len(sc.macroCalls)-1]
	return true
}

func (sc *SimpleContext) GetMacroCall() (string, int, map[string]bool) {
	if len(sc.macroCalls) == 0 {
		return "", 0, nil
	}
	mc := sc.macroCalls[len(sc.macroCalls)-1]
	return mc.name, mc.number, mc.locals
}

func (sc *SimpleContext) LastLabel() string {
	return sc.lastLabel
}

func (sc *SimpleContext) SetLastLabel(l string) {
	sc.lastLabel = l
}

func (sc *SimpleContext) DivZero() *uint16 {
	return sc.divZeroVal
}

func (sc *SimpleContext) SetDivZero(val uint16) {
	sc.divZeroVal = &val
}
