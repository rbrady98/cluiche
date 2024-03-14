package main

const (
	ClockSpeed     = 4213440
	CyclesPerFrame = 70224
)

type Gameboy struct {
	cpu    *CPU
	ppu    *PPU
	memory *Memory
	// input lower nibble contains d pad inputs and higher nibble contains buttons
	input *Input
}

func NewGameboy(romPath string) (*Gameboy, error) {
	mem := NewMemory()
	cpu := NewCPU(mem)
	ppu := NewPPU(cpu, mem)
	input := NewInput()

	mem.cpu = cpu
	mem.input = input

	gb := &Gameboy{
		cpu:    cpu,
		ppu:    ppu,
		memory: mem,
		input:  input,
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
		// g.debugLog()

		c := g.cpu.Update()
		g.ppu.Update(c)
		frameCycles += c
	}
}

func (g *Gameboy) GetRenderedFrame() []byte {
	return g.ppu.frameBufferToBytes()
}

func (g *Gameboy) UpdateButtons(pressed, released []Button) {
	for _, p := range pressed {
		g.input.PressButton(g.cpu, p)
	}

	for _, r := range released {
		g.input.ReleaseButton(r)
	}
}
