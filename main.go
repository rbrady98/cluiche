package main

import (
	"log"

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

	game := NewGame(160, 144, "./roms/dr-mario.gb")
	ebiten.SetWindowSize(160*3, 144*3)
	// ebiten.SetWindowTitle(gb.GetRomTitle())
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal("game error:", err)
	}
}
