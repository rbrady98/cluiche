package main

import (
	// "bufio"
	// "fmt"
	// "os"
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

	return gb, nil
}

func (g *Gameboy) GetRomTitle() string {
	return g.memory.GetCartTitle()
}

// Update updates the state for a single frame
func (g *Gameboy) Update() {
	var frameCycles int

	// TODO: track cycles spent on operation
	for frameCycles < CyclesPerFrame {
		// fmt.Println(frameCycles)
		g.cpu.Update()
    g.ppu.Update(1)
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
