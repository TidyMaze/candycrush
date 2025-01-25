package main

import (
	"gioui.org/app"
	"gioui.org/unit"
	"gioui.org/widget"
	"log"
	"os"
)

func main() {
	go func() {
		window := new(app.Window)

		window.Option(app.Size(
			unit.Dp(gameState.Board.Width)*cellSizeDp,
			unit.Dp(gameState.Board.Height)*cellSizeDp,
		))

		// create clickables
		clickables = make([]widget.Clickable, gameState.Board.Width*gameState.Board.Height)

		err := run(window)
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}
