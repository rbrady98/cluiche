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

	// bgColourMap caches the background/window pixels on a scanline which use colour id 0
	// so that sprites can be correctly drawn when they have the priority flag set
	bgColourMap []bool
}

func NewPPU(cpu *CPU, mem *Memory) *PPU {
	return &PPU{
		mem:         mem,
		cpu:         cpu,
		bgColourMap: make([]bool, ScreenWidth),
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
		if p.dots >= 80 {
			p.setMode(status, Mode3)
			p.dots = 0
		}

	case Mode3:
		// drawing
		if p.dots >= 172 {
			p.setMode(status, Mode0)
			p.dots = 0

			if TestBit(status, 3) {
				p.cpu.requestInterrupt(1)
			}

			lcdc := p.mem.Read(LCDC)

			p.RenderBackground(lcdc)
			if TestBit(lcdc, 1) {
				p.RenderSprites(lcdc)
			}

			clear(p.bgColourMap)
		}

	case Mode0:
		if p.dots >= 204 {
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
	scx := p.mem.Read(SCX)
	scy := p.mem.Read(SCY)
	wx := p.mem.Read(WX) - 7
	wy := p.mem.Read(WY)

	currentLine := p.getLine()
	inWindow := p.inWindow(control, int(wy), currentLine)
	tileDataAddr := p.getTileDataAddress(control)
	tileMapAddr := p.getTileMapAddress(control, inWindow)
	palette := p.mem.Read(0xFF47)

	// set current x and y position considering scroll
	yPos := (currentLine + int(scy)) % 256
	if inWindow {
		yPos = currentLine - int(wy)
	}

	tileYOffset := uint16(yPos/8) * 32

	for pixel := byte(0); pixel < ScreenWidth; pixel++ {
		xPos := uint16(pixel+scx) % 256
		if inWindow && pixel >= wx {
			xPos = uint16(pixel - wx)
		}

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

		colourID := toColourID(d1, d2, byte(xOffset))

		p.RenderPixel(colourID, palette, int(pixel), currentLine)

		p.bgColourMap[pixel] = colourID == 0
	}
}

func (p *PPU) RenderSprites(control byte) {
	currentLine := p.getLine()
	pal1 := p.mem.Read(0xFF48)
	pal2 := p.mem.Read(0xFF49)

	size := 8
	if TestBit(control, 2) {
		size = 16
	}

	var spriteXPositions [ScreenWidth]int

	// loop over all 40 sprites, check which sprites are to be rendered
	spriteLimit := 0
	for i := uint16(0); i < 40; i++ {
		addr := i * 4

		// read the sprite into memory
		yPos := p.mem.Read(0xFE00+addr) - 16
		xPos := p.mem.Read(0xFE00+addr+1) - 8
		tileIdx := p.mem.Read(0xFE00 + addr + 2)
		flags := p.mem.Read(0xFE00 + addr + 3)

		// check if the sprite should be rendered
		if currentLine < int(yPos) || currentLine >= int(yPos)+size {
			continue
		}

		spriteLimit++
		if spriteLimit > 10 {
			break
		}

		xFlip := TestBit(flags, 5)
		yFlip := TestBit(flags, 6)
		priority := TestBit(flags, 7)

		// get the line of the sprite tile that we need to render
		line := currentLine - int(yPos)
		if yFlip {
			line = size - line - 1
		}

		dataAddr := 0x8000 + (uint16(tileIdx) * 16) + uint16(line*2)
		d1 := p.mem.Read(dataAddr)
		d2 := p.mem.Read(dataAddr + 1)

		// draw the tile line
		for tilePixel := byte(0); tilePixel < 8; tilePixel++ {
			x := xPos + (7 - tilePixel)
			if xFlip {
				x = xPos + tilePixel
			}

			if x > ScreenWidth {
				continue
			}

			// check the sprite priority of the current pixel
			if spriteXPositions[x] != 0 && spriteXPositions[x] <= int(xPos)+1 {
				continue
			}

			// if we have bg priority flag set and the colour id of current pixel in the background
			// is not colour id 0 then we skip this as we are drawing sprites below the bg
			if priority && !p.bgColourMap[x] {
				continue
			}

			colourID := toColourID(d1, d2, tilePixel)
			if colourID == 0 {
				continue
			}

			pal := pal1
			if TestBit(flags, 4) {
				pal = pal2
			}

			p.RenderPixel(colourID, pal, int(x), currentLine)

			spriteXPositions[x] = int(xPos + 1) // add one here to use 0 as empty value
		}
	}
}

func (p *PPU) RenderPixel(colourID byte, palette byte, x int, y int) {
	colour := (palette >> (colourID * 2) & 0x3)

	r, g, b := toScreenColour(colour)

	p.DrawPixel(x, y, r, g, b)
}

func (p *PPU) DrawPixel(x, y int, r, g, b byte) {
	p.frame[y][x][0] = r
	p.frame[y][x][1] = g
	p.frame[y][x][2] = b
}

func (p *PPU) getTileMapAddress(control byte, inWindow bool) uint16 {
	var base uint16 = 0x9800

	bit := 3
	if inWindow {
		bit = 6
	}

	if TestBit(control, bit) {
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

func (p *PPU) inWindow(control byte, wy int, line int) bool {
	return TestBit(control, 5) && line >= wy
}

func toColourID(d1, d2, pixel byte) byte {
	return ((d2>>pixel)&1)<<1 | (d1>>pixel)&1
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
