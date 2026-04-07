//go:build js

package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	ebiten.SetTPS(TPS)

	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
