package main

import (
	"fmt"
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
	mem [0xFFFF + 1]byte

	cpu *CPU
}

func NewMemory() *Memory {
	m := &Memory{}

	m.mem[0xFF05] = 0x00
	m.mem[0xFF06] = 0x00
	m.mem[0xFF07] = 0xF8
	m.mem[0xFF0F] = 0xE1
	m.mem[0xFF10] = 0x80
	m.mem[0xFF11] = 0xBF
	m.mem[0xFF12] = 0xF3
	m.mem[0xFF14] = 0xBF
	m.mem[0xFF16] = 0x3F
	m.mem[0xFF17] = 0x00
	m.mem[0xFF19] = 0xBF
	m.mem[0xFF1A] = 0x7F
	m.mem[0xFF1B] = 0xFF
	m.mem[0xFF1C] = 0x9F
	m.mem[0xFF1E] = 0xBF
	m.mem[0xFF20] = 0xFF
	m.mem[0xFF21] = 0x00
	m.mem[0xFF22] = 0x00
	m.mem[0xFF23] = 0xBF
	m.mem[0xFF24] = 0x77
	m.mem[0xFF25] = 0xF3
	m.mem[0xFF26] = 0xF1
	m.mem[0xFF40] = 0x91
	m.mem[0xFF42] = 0x00
	m.mem[0xFF43] = 0x00
	m.mem[0xFF45] = 0x00
	m.mem[0xFF47] = 0xFC
	m.mem[0xFF48] = 0xFF
	m.mem[0xFF49] = 0xFF
	m.mem[0xFF4A] = 0x00
	m.mem[0xFF4B] = 0x00
	m.mem[0xFFFF] = 0x00

	return m
}

func (m *Memory) Read(addr uint16) byte {
	// just for debugging
	// if addr == 0xFF44 || addr == 0xff02 {
	// 	return 0x90
	// }
	return m.mem[addr]
}

func (m *Memory) Write(addr uint16, val byte) {
	if addr == 0xFF01 {
		fmt.Print(string(m.mem[0xFF01]))
	}

	// reset the divider register when written to
	if addr == DIV {
		m.mem[addr] = 0
		m.cpu.dividerCounter = 0
		m.cpu.SetClockFreq()
		return
	}

	// might reset if tac is written to
	if addr == TAC {
		curFreq := m.cpu.GetClockFreq()
		m.mem[addr] = val
		newFreq := m.cpu.GetClockFreq()

		if newFreq != curFreq {
			m.cpu.SetClockFreq()
		}

		return
	}

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

func (m *Memory) GetCartTitle() string {
	title := m.mem[0x0134:0x0143]

	return string(title)
}

func (m *Memory) GetLicenseeCode() {
	code := m.mem[0x0144:0x0145]

	fmt.Println(string(code))
}

func (m *Memory) GetCartidgeType() {
	t := m.Read(0x0147)

	fmt.Println(t)
}
