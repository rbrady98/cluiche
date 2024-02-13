package main

import (
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	file := "./logs/gb.log"
	logFile, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		panic(err.Error())
	}
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))
	log.SetOutput(logFile)

	game := NewGame(160, 144, "./roms/09-op-r,r.gb")
	ebiten.SetWindowSize(640, 480)
	// ebiten.SetWindowTitle(gb.GetRomTitle())
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal("game error:", err)
	}
}
