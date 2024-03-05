package main

import (
	"log"
)

const (
	ClockSpeed     = 4213440
	CyclesPerFrame = 70224
)

type Gameboy struct {
	cpu    *CPU
	ppu    *PPU
	memory *Memory
}

func NewGameboy(romPath string) (*Gameboy, error) {
	mem := NewMemory()
	cpu := NewCPU(mem)
	ppu := NewPPU(cpu, mem)

	mem.cpu = cpu

	gb := &Gameboy{
		cpu,
		ppu,
		mem,
	}

	err := gb.memory.LoadROM(romPath)
	if err != nil {
		return nil, err
	}

	gb.GetCartType()

	return gb, nil
}

func (g *Gameboy) GetRomTitle() string {
	return g.memory.GetCartTitle()
}

func (g *Gameboy) GetCartType() {
	g.memory.GetCartidgeType()
}

// Update updates the state for a single frame
func (g *Gameboy) Update() {
	var frameCycles int

	for frameCycles < CyclesPerFrame {
		g.debugLog()

		c := g.cpu.Update()
		g.ppu.Update(c)
		frameCycles += c
	}
}

func (g *Gameboy) GetRenderedFrame() []byte {
	return g.ppu.frameBufferToBytes()
}

func (g *Gameboy) debugLog() {
	log.Printf(
		"A:%02X F:%02X B:%02X C:%02X D:%02X E:%02X H:%02X L:%02X SP:%04X PC:%04X PCMEM:%02X,%02X,%02X,%02X DIV:%02X TIMA:%02X TAC:%02X IF:%02X Internal TIMA counter: %d",
		g.cpu.registers.a,
		g.cpu.registers.f.toByte(),
		g.cpu.registers.b,
		g.cpu.registers.c,
		g.cpu.registers.d,
		g.cpu.registers.e,
		g.cpu.registers.h,
		g.cpu.registers.l,
		g.cpu.sp,
		g.cpu.pc,
		g.memory.Read(g.cpu.pc),
		g.memory.Read(g.cpu.pc+1),
		g.memory.Read(g.cpu.pc+2),
		g.memory.Read(g.cpu.pc+3),
		g.memory.Read(DIV),
		g.memory.Read(TIMA),
		g.memory.Read(TAC),
		g.memory.Read(InterruptFlagReg),
		g.cpu.timerCounter,
	)
}
