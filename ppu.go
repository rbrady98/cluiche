package main

const (
	ScreenWidth  = 160
	ScreenHeight = 144

	Mode0 int = iota // Horizontal Blank
	Mode1            // Vertical Blank
	Mode2            // OAM Scan
	Mode3            // Drawing

	// Registers
	LCDC uint16 = 0xFF40 // LCD Control register
	LY   uint16 = 0xFF44 // LCD Y coordinate
	STAT uint16 = 0xFF41 // LCD status register
)

type PPU struct {
	mem  *Memory
	dots int
}

// Update
func (p *PPU) Update(cycles int) {
	status := p.mem.Read(STAT)
	mode := getMode(status)
	line := p.getLine()

	p.dots += cycles

	switch mode {
	case Mode2:
		// OAM Read
		if p.dots >= 80 {
			// move to drawing
			p.setMode(Mode3)
			p.dots = 0
		}

	case Mode3:
		// drawing
		if p.dots >= 172 {
			// move to h blank
			p.setMode(Mode0)
			p.dots = 0

			// draw the scanline as we finish the mode
		}

	case Mode0:
		if p.dots >= 204 {
			p.dots = 0
			p.setLine(line + 1)
		}

	case Mode1: // V Blank
		if p.dots >= 456 {
			p.dots = 0

			// TODO: after the vblank is finished we have fully rendered a frame
			// set the current line to 0, push the framebuffer to the screen
			if line == 153 {
				p.setLine(0)
			} else {
				p.setLine(line + 1)
			}
		}
	}

	// TODO: set the new status back into the status register
}

func (p *PPU) RenderBackground() {}
func (p *PPU) RenderSprites()    {}
func (p *PPU) getTilePatternBaseAddress(control byte) uint16 {
	var base uint16 = 0x8000

	if TestBit(control, 4) {
		base = 0x8800
	}

	return base
}

func getMode(stat byte) int {
	return int(stat & 0xFC)
}

func (p *PPU) setMode(mode int) int {
	return 0
}

func (p *PPU) getLine() int {
	return int(p.mem.Read(LY))
}

func (p *PPU) setLine(line int) {
	p.mem.Write(LY, byte(line))
}
