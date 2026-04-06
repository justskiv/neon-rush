package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	// Fit window to monitor with some padding.
	mw, mh := ebiten.Monitor().Size()
	scale := min(
		float64(mw)*0.85/ScreenWidth,
		float64(mh)*0.85/ScreenHeight,
	)
	if scale < 1 {
		scale = 1
	}

	ebiten.SetWindowSize(int(ScreenWidth*scale), int(ScreenHeight*scale))
	ebiten.SetWindowTitle("Neon Rush")
	ebiten.SetTPS(TPS)

	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
