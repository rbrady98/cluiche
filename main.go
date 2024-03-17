package main

import (
	"log"
	// "os"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	// file := "./logs/gb.log"
	// logFile, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o666)
	// if err != nil {
	// 	panic(err.Error())
	// }
	// log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))
	// log.SetOutput(logFile)

	game := NewGame(160*2, 144*2, "./roms/kirbys-dreamland.gb")
	ebiten.SetWindowSize(160*4, 144*4)
	ebiten.SetWindowTitle(game.gb.GetRomTitle())
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal("game error:", err)
	}
}
