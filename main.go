package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	game := NewGame(160, 144, "./roms/misc-instrs.gb")
	ebiten.SetWindowSize(640, 480)
	// ebiten.SetWindowTitle(gb.GetRomTitle())
	ebiten.SetTPS(1)
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal("game error:", err)
	}
}
