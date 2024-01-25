package main

const (
	CarryFlagBitPosition     byte = 4
	HalfCarryFlagBitPosition byte = 5
	SubtractFlagBitPosition  byte = 6
	ZeroFlagBitPosition      byte = 7
)

type Registers struct {
	a byte
	b byte
	c byte
	d byte
	e byte
	f *FlagRegisters
	h byte
	l byte
}

func NewRegisters() *Registers {
	return &Registers{
		a: 0x01,
		b: 0x00,
		c: 0x13,
		d: 0x00,
		e: 0xd8,
		h: 0x01,
		l: 0x4d,
		f: flagsFromByte(0x01),
	}
}

func (r *Registers) getAF() uint16 {
	return uint16(r.a)<<8 | uint16(r.f.toByte())
}

func (r *Registers) setAF(value uint16) {
	r.a = byte((value & 0xFF00) >> 8)
	r.f = flagsFromByte(byte((value & 0xFF) >> 8))
}

func (r *Registers) getBC() uint16 {
	return uint16(r.b)<<8 | uint16(r.c)
}

func (r *Registers) setBC(value uint16) {
	r.b = byte((value & 0xFF00) >> 8)
	r.c = byte((value & 0xFF) >> 8)
}

func (r *Registers) getDE() uint16 {
	return uint16(r.d)<<8 | uint16(r.e)
}

func (r *Registers) setDE(value uint16) {
	r.d = byte((value & 0xFF00) >> 8)
	r.e = byte((value & 0xFF) >> 8)
}

func (r *Registers) getHL() uint16 {
	return uint16(r.h)<<8 | uint16(r.l)
}

func (r *Registers) setHL(value uint16) {
	r.h = byte((value & 0xFF00) >> 8)
	r.l = byte((value & 0xFF) >> 8)
}

type FlagRegisters struct {
	Carry     bool
	HalfCarry bool
	Subtract  bool
	Zero      bool
}

func (f *FlagRegisters) toByte() byte {
	var b byte
	if f.Carry {
		b = b | 1<<CarryFlagBitPosition
	}
	if f.HalfCarry {
		b = b | 1<<HalfCarryFlagBitPosition
	}
	if f.Subtract {
		b = b | 1<<SubtractFlagBitPosition
	}
	if f.Zero {
		b = b | 1<<ZeroFlagBitPosition
	}

	return b
}

func flagsFromByte(value byte) *FlagRegisters {
	carry := (value>>CarryFlagBitPosition)&0b1 == 1
	halfCarry := (value>>HalfCarryFlagBitPosition)&0b1 == 1
	subtract := (value>>SubtractFlagBitPosition)&0b1 == 1
	zero := (value>>ZeroFlagBitPosition)&0b1 == 1

	return &FlagRegisters{
		carry,
		halfCarry,
		subtract,
		zero,
	}
}
