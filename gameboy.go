package main

import "time"

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
	frametime := 1 * time.Second
	ticker := time.NewTicker(frametime)

	for range ticker.C {
		g.Update()
	}
}

// Update updates the state for a single frame
func (g *Gameboy) Update() {
	g.cpu.tick()
}
