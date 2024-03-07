package main

const (
	ScreenWidth  = 160
	ScreenHeight = 144
	ScreenPixels = 23040

	Mode0 byte = 0 // Horizontal Blank
	Mode1 byte = 1 // Vertical Blank
	Mode2 byte = 2 // OAM Scan
	Mode3 byte = 3 // Drawing

	// Registers
	LCDC uint16 = 0xFF40 // LCD Control register
	LY   uint16 = 0xFF44 // LCD Y coordinate
	STAT uint16 = 0xFF41 // LCD status register

	SCY uint16 = 0xFF42 // BG Viewport Y
	SCX uint16 = 0xFF43 // BG Viewport X
	WY  uint16 = 0xFF4A // Window Y
	WX  uint16 = 0xFF4B // Window X
)

type PPU struct {
	mem *Memory
	cpu *CPU

	dots int

	frame [ScreenHeight][ScreenWidth][3]byte
}

func NewPPU(cpu *CPU, mem *Memory) *PPU {
	return &PPU{
		mem: mem,
		cpu: cpu,
	}
}

// Update
func (p *PPU) Update(cycles int) {
	status := p.mem.Read(STAT)
	mode := getMode(status)
	line := p.getLine()

	p.dots += cycles

	if line == 144 {
		p.setMode(status, Mode1)
		mode = Mode1

		if TestBit(status, 4) {
			p.cpu.requestInterrupt(1)
		}

		p.cpu.requestInterrupt(0)
	}

	switch mode {
	case Mode2:
		// OAM Read
		if p.dots >= 80 {
			// fmt.Println("Finished OAM READ")

			// move to drawing
			p.setMode(status, Mode3)
			p.dots = 0
		}

	case Mode3:
		// drawing
		if p.dots >= 172 {
			// fmt.Println("Mode 3 complete: finished drawing")
			// move to h blank
			p.setMode(status, Mode0)
			p.dots = 0

			if TestBit(status, 3) {
				p.cpu.requestInterrupt(1)
			}

			// draw the scanline as we finish the mode
			p.RenderBackground(p.mem.Read(LCDC))
		}

	case Mode0:
		if p.dots >= 204 {
			// fmt.Println("Mode 0 complete")
			p.dots = 0
			p.setLine(line + 1)
			p.setMode(status, Mode2)

			if TestBit(status, 5) {
				p.cpu.requestInterrupt(1)
			}
		}

	case Mode1: // V Blank
		if p.dots >= 456 {
			p.dots = 0

			if line == 153 {
				p.setLine(0)
				p.setMode(status, Mode2)

				if TestBit(status, 5) {
					p.cpu.requestInterrupt(1)
				}
			} else {
				p.setLine(line + 1)
			}
		}
	}
}

func (p *PPU) RenderBackground(control byte) {
	// get the tile offset that we should be using

	tileDataAddr := p.getTileDataAddress(control)
	tileMapAddr := p.getTileMapAddress(control)

	// set current x and y position considering scroll
	yPos := (p.getLine() + int(p.mem.Read(SCY))) % 256
	tileYOffset := uint16(yPos/8) * 32

	for pixel := byte(0); pixel < ScreenWidth; pixel++ {
		xPos := uint16(pixel+p.mem.Read(SCX)) % 256
		tileXOffset := xPos / 8

		tileNumAddr := tileMapAddr + tileYOffset + tileXOffset

		tileNum := p.mem.Read(tileNumAddr)
		var tileAddr uint16
		if tileDataAddr == 0x9000 {
			tileAddr = uint16(int32(tileDataAddr) + int32(int8(tileNum))*16)
		} else {
			tileAddr = tileDataAddr + (uint16(tileNum) * 16)
		}

		// read correct two bytes based on current line
		yOffset := (yPos % 8) * 2
		xOffset := 7 - (xPos % 8)

		d1 := p.mem.Read(tileAddr + uint16(yOffset))
		d2 := p.mem.Read(tileAddr + uint16(yOffset) + 1)

		colourID := ((d2>>xOffset)&1)<<1 | (d1>>xOffset)&1

		palette := p.mem.Read(0xFF47)

		colour := (palette >> (colourID * 2) & 0x3)
		r, g, b := toScreenColour(colour)

		p.DrawPixel(int(pixel), p.getLine(), r, g, b)
	}
}
func (p *PPU) RenderSprites() {}

func (p *PPU) DrawPixel(x, y int, r, g, b byte) {
	p.frame[y][x][0] = r
	p.frame[y][x][1] = g
	p.frame[y][x][2] = b
}

func (p *PPU) getTileMapAddress(control byte) uint16 {
	var base uint16 = 0x9800

	// TODO: handle window bit change
	if TestBit(control, 3) {
		base = 0x9C00
	}

	return base
}

func (p *PPU) getTileDataAddress(control byte) uint16 {
	var base uint16 = 0x9000

	if TestBit(control, 4) {
		base = 0x8000
	}

	return base
}

func toScreenColour(color byte) (r, g, b byte) {
	switch color {
	case 0:
		return 255, 255, 255
	case 1:
		return 192, 192, 192
	case 2:
		return 96, 96, 96
	case 3:
		return 0, 0, 0
	default:
		return 255, 255, 255
	}
}

func getMode(stat byte) byte {
	return (stat & 0x3)
}

func (p *PPU) setMode(stat byte, mode byte) {
	s := (stat & 0xFC) | mode
	p.mem.Write(STAT, s)
}

func (p *PPU) getLine() int {
	return int(p.mem.Read(LY))
}

func (p *PPU) setLine(line int) {
	p.mem.Write(LY, byte(line))
}

func (p *PPU) frameBufferToBytes() []byte {
	frame := make([]byte, 0, 4*ScreenHeight*ScreenWidth)
	for y := 0; y < ScreenHeight; y++ {
		for x := 0; x < ScreenWidth; x++ {
			frame = append(frame, p.frame[y][x][0])
			frame = append(frame, p.frame[y][x][1])
			frame = append(frame, p.frame[y][x][2])
			frame = append(frame, 0xFF)
		}
	}

	return frame
}
