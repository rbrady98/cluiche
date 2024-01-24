package main

import (
	"fmt"
	"time"
)

type Gameboy struct {
	cpu    *CPU
	memory *Memory
}

func NewGameboy(romPath string) (*Gameboy, error) {
	mem := NewMemory()
	cpu := NewCPU(mem)

	gb := &Gameboy{
		cpu,
		mem,
	}
	err := gb.memory.LoadROM(romPath)
	if err != nil {
		return nil, err
	}

	return gb, nil
}

func (g *Gameboy) Run() {
	frametime := 10 * time.Millisecond
	ticker := time.NewTicker(frametime)
	c := 0

	for range ticker.C {
		fmt.Println("Ticks:", c)
		g.Update()
		c++
	}
}

// Update updates the state for a single frame
func (g *Gameboy) Update() {
	g.cpu.tick()
}
