package main

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
