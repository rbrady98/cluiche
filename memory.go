package main

import (
	"fmt"
	"os"
)

const (
	CartridgeROM = 0x8000
)

type Memory struct {
	mem [0xFFFF + 1]byte

	cpu   *CPU
	input *Input
}

func NewMemory() *Memory {
	m := &Memory{}

	m.mem[0xFF00] = 0xCF
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

	if addr == JOYP {
		return m.input.GetInput(m.mem[JOYP])
	}

	return m.mem[addr]
}

func (m *Memory) Write(addr uint16, val byte) {
	// if addr == 0xFF01 {
	// 	fmt.Print(string(m.mem[0xFF01]))
	// }
	if addr == JOYP {
		m.mem[addr] = val & 0x30
		return
	}

	// reset the divider register when written to
	if addr == DIV {
		m.mem[addr] = 0
		m.cpu.dividerCounter = 0
		m.cpu.SetClockFreq()
		return
	}

	// reset if tac is written to
	if addr == TAC {
		curFreq := m.cpu.GetClockFreq()
		m.mem[addr] = val
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

func (m *Memory) doDMATransfer(value byte) {
	// TODO: how can we get this to take 160 cycles
	addr := uint16(value) << 8

	for i := uint16(0); i < 0xA0; i++ {
		m.mem[0xFE00+i] = m.Read(addr + i)
	}
}
