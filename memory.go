package main

import (
	"os"

	"github.com/rbrady98/cluiche/cartridge"
)

const (
	//
	CartridgeROM            = 0x8000
	VRAM                    = 0xA000
	ExternalRAM             = 0xC000
	WRAM                    = 0xE000
	EchoRAM                 = 0xFE00
	OAM                     = 0xFEA0
	NotUsable               = 0xFF00
	IO                      = 0xFF80
	HRAM                    = 0xFFFF
	InterruptEnableRegister = 0xFFFF
)

type Memory struct {
	// cart memory
	cart *cartridge.Cart
	// VRAM
	vram [0x2000]byte
	// WRAM
	wram [0x2000]byte
	// OAM
	oam [0x100]byte
	// IO
	io [0x80]byte
	// HRAM
	hram [0x7F]byte

	interruptEnable byte
	// mem  [0xFFFF + 1]byte

	cpu   *CPU
	input *Input
}

func NewMemory() *Memory {
	m := &Memory{}

	m.io[0x00] = 0xCF
	m.io[0x05] = 0x00
	m.io[0x06] = 0x00
	m.io[0x07] = 0xF8
	m.io[0x0F] = 0xE1
	m.io[0x10] = 0x80
	m.io[0x11] = 0xBF
	m.io[0x12] = 0xF3
	m.io[0x14] = 0xBF
	m.io[0x16] = 0x3F
	m.io[0x17] = 0x00
	m.io[0x19] = 0xBF
	m.io[0x1A] = 0x7F
	m.io[0x1B] = 0xFF
	m.io[0x1C] = 0x9F
	m.io[0x1E] = 0xBF
	m.io[0x20] = 0xFF
	m.io[0x21] = 0x00
	m.io[0x22] = 0x00
	m.io[0x23] = 0xBF
	m.io[0x24] = 0x77
	m.io[0x25] = 0xF3
	m.io[0x26] = 0xF1
	m.io[0x40] = 0x91
	m.io[0x42] = 0x00
	m.io[0x43] = 0x00
	m.io[0x45] = 0x00
	m.io[0x47] = 0xFC
	m.io[0x48] = 0xFF
	m.io[0x49] = 0xFF
	m.io[0x4A] = 0x00
	m.io[0x4B] = 0x00
	m.interruptEnable = 0x00

	return m
}

func (m *Memory) Read(addr uint16) byte {
	switch {
	case addr < CartridgeROM:
		return m.cart.Read(addr)

	case addr < VRAM:
		return m.vram[addr-CartridgeROM]

	case addr < ExternalRAM:
		return m.cart.Read(addr)

	case addr < WRAM:
		return m.wram[addr-ExternalRAM]

	case addr < EchoRAM:
		// noop not implementing echo ram
		return 0xFF

	case addr < OAM:
		return m.oam[addr-EchoRAM]

	case addr < NotUsable:
		// noop gameboy is not allowed to read from this addr
		return 0xFF

	case addr < IO:
		if addr == JOYP {
			return m.input.GetInput(m.io[JOYP-0xFF00])
		}

		return m.io[addr-NotUsable]

	case addr < HRAM:
		return m.hram[addr-IO]

	default:
		return m.interruptEnable
	}
}

func (m *Memory) Write(addr uint16, val byte) {
	switch {
	case addr < CartridgeROM:
		m.cart.WriteROM(addr, val)

	case addr < VRAM:
		m.vram[addr-CartridgeROM] = val

	case addr < ExternalRAM:
		m.cart.WriteRAM(addr, val)

	case addr < WRAM:
		m.wram[addr-ExternalRAM] = val

	case addr < EchoRAM:
		// noop not implementing echo ram

	case addr < OAM:
		m.oam[addr-EchoRAM] = val

	case addr < NotUsable:
		// noop gameboy is not allowed to read from this addr

	case addr < IO:
		// TODO: if this gets longer write a WriteIO method
		if addr == JOYP {
			m.io[addr-NotUsable] = val & 0x30
			return
		}

		// reset the divider register when written to
		if addr == DIV {
			m.io[addr-NotUsable] = 0
			m.cpu.dividerCounter = 0
			m.cpu.SetClockFreq()
			return
		}

		// reset if tac is written to
		if addr == TAC {
			curFreq := m.cpu.GetClockFreq()
			m.io[addr-NotUsable] = val
			newFreq := m.cpu.GetClockFreq()

			if newFreq != curFreq {
				m.cpu.SetClockFreq()
			}

			return
		}

		// DMA transfer
		if addr == 0xFF46 {
			m.doDMATransfer(val)
			return
		}

		m.io[addr-NotUsable] = val

	case addr < HRAM:
		m.hram[addr-IO] = val

	default:
		m.interruptEnable = val
	}
}

func (m *Memory) LoadROM(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	c, err := cartridge.NewCart(data)
	if err != nil {
		return err
	}

	m.cart = c

	return nil
}

func (m *Memory) GetCartTitle() string {
	return m.cart.Title()
}

func (m *Memory) doDMATransfer(value byte) {
	// TODO: how can we get this to take 160 cycles
	addr := uint16(value) << 8

	for i := uint16(0); i < 0xA0; i++ {
		m.oam[i] = m.Read(addr + i)
	}
}
