package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	game := NewGame(160*2, 144*2, "./roms/cpu_instrs.gb")
	ebiten.SetWindowSize(160*4, 144*4)
	ebiten.SetWindowTitle(game.gb.GetRomTitle())
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal("game error:", err)
	}
}
