package main

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type Game struct {
	gb *Gameboy

	width  int
	height int

	img *ebiten.Image
}

func NewGame(w, h int, romPath string) *Game {
	gb, err := NewGameboy(romPath)
	if err != nil {
		return nil
	}

	return &Game{
		width:  w,
		height: h,
		img:    ebiten.NewImage(w, h),
		gb:     gb,
	}
}

func (g *Game) Update() error {
	g.gb.Update()
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.WritePixels(g.gb.GetRenderedFrame())
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (width, height int) {
	return 160, 144
}
