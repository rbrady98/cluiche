package main

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
	for frameCycles < CyclesPerFrame/4 {
		// fmt.Println(frameCycles)
		g.cpu.Update()
		g.ppu.Update(4)
		// tick the timers

		frameCycles += 4
	}
}

func (g *Gameboy) GetRenderedFrame() []byte {
	return g.ppu.frameBufferToBytes()
}
