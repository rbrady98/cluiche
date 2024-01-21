package main

import (
	"os"
)

const (
	CartridgeROM = 0x8000
	VRAM         = 0xA000 // 8kB video ram
	ExternalRAM  = 0xC000 // ram from the cartridge
	WRAM         = 0xE000
	EchoRAM      = 0xE000 // echo of internal ram, can be ignored
	OAM          = 0xFE00 // sprite object memory
	Empty        = 0xFEA0 // empty memory but unusable for io
	IOPorts      = 0xFF00 // memory mapped io lives here
	Empty2       = 0xFE4C // empty memory but unusable for io
	HRAM         = 0xFF80 // internal ram at the top of the map
)

type Memory struct {
	mem [0xFFFF]byte
}

func NewMemory() *Memory {
	return &Memory{}
}

func (m *Memory) Read(addr uint16) byte {
	return m.mem[addr]
}

func (m *Memory) Write(addr uint16, val byte) {
	m.mem[addr] = val
}

func (m *Memory) LoadROM(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	copy(m.mem[:CartridgeROM], data)
	return nil
}
