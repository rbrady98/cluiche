package main

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
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
		panic(err)
	}

	ebiten.SetTPS(60)
	ebiten.SetVsyncEnabled(false)

	return &Game{
		width:  w,
		height: h,
		img:    ebiten.NewImage(w, h),
		gb:     gb,
	}
}

func (g *Game) Update() error {
	p, r := Buttons()
	g.gb.UpdateButtons(p, r)
	g.gb.Update()

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.WritePixels(g.gb.GetRenderedFrame())
	ebitenutil.DebugPrint(screen, fmt.Sprintf("fps: %.2f\ntps: %.2f", ebiten.ActualFPS(), ebiten.ActualTPS()))
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (width, height int) {
	return 160, 144
}

var keyMap = map[ebiten.Key]Button{
	ebiten.KeyArrowUp:    ButtonUp,
	ebiten.KeyArrowDown:  ButtonDown,
	ebiten.KeyArrowLeft:  ButtonLeft,
	ebiten.KeyArrowRight: ButtonRight,
	ebiten.KeyZ:          ButtonA,
	ebiten.KeyX:          ButtonB,
	ebiten.KeyComma:      ButtonStart,
	ebiten.KeyPeriod:     ButtonSelect,
}

// Buttons returns the two slices containing the pressed and released buttons for the current frame.
func Buttons() ([]Button, []Button) {
	var p []Button
	var r []Button

	for key, button := range keyMap {
		if ok := inpututil.IsKeyJustPressed(key); ok {
			p = append(p, button)
		}

		if ok := inpututil.IsKeyJustReleased(key); ok {
			r = append(r, button)
		}
	}

	return p, r
}
