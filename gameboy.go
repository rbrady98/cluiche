package main

import (
	"log"
)

const CyclesPerFrame = 70224

type Gameboy struct {
	cpu    *CPU
	ppu    *PPU
	memory *Memory
}

func NewGameboy(romPath string) (*Gameboy, error) {
	mem := NewMemory()
	cpu := NewCPU(mem)
	ppu := NewPPU(mem)

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

	// TODO: track cycles spent on operation
	for frameCycles < CyclesPerFrame {
		g.debugLog()

		g.cpu.Update()

		// log for debugging
		// g.ppu.Update(1)
		// if frameCycles%1000 == 0 {
		//     fmt.Println(g.cpu.registers)
		// 	fmt.Println("Waiting for enter key to conitnue")
		// 	bufio.NewReader(os.Stdin).ReadBytes('\n')
		// }
		// tick the timers

		frameCycles += 1
	}
}

func (g *Gameboy) GetRenderedFrame() []byte {
	return g.ppu.frameBufferToBytes()
}

func (g *Gameboy) debugLog() {
	log.Printf(
		"A:%02X F:%02X B:%02X C:%02X D:%02X E:%02X H:%02X L:%02X SP:%04X PC:%04X PCMEM:%02X,%02X,%02X,%02X",
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
	)
}
