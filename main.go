package main

import (
	"fmt"
	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"image"
	"image/color"
	"log"
	"os"
	"time"
)

var displayedTick int = 0

var engine Engine = Engine{}
var gameState State = engine.InitRandom()

const cellSizeDp = 100

const textSize = unit.Sp(24)

func main() {
	go func() {
		window := new(app.Window)

		window.Option(app.Size(
			unit.Dp(float32(gameState.Board.Width*cellSizeDp)),
			unit.Dp(float32(gameState.Board.Height*cellSizeDp)),
		))

		err := run(window)
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

func tickEmitter(tickChannel chan int) {
	for i := 0; ; i++ {
		tickChannel <- i
		time.Sleep(1 * time.Second)
	}
}

func invalidator(tickChannel chan int, window *app.Window) {
	for t := range tickChannel {
		println(fmt.Sprintf("Tick: %d", t))
		displayedTick = t
		window.Invalidate()
	}
}

func draw(window *app.Window) error {

	maroon := color.NRGBA{R: 127, G: 0, B: 0, A: 255}

	theme := material.NewTheme()
	var ops op.Ops
	for {
		switch e := window.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:

			println("New frame")

			gtx := app.NewContext(&ops, e)
			drawGrid(gtx)
			drawTick(theme, maroon, gtx, textSize)
			e.Frame(gtx.Ops)
		}
	}
}

func drawTick(theme *material.Theme, maroon color.NRGBA, gtx layout.Context, textSize unit.Sp) {
	title := material.Label(theme, textSize, fmt.Sprintf("Tick: %d", displayedTick))
	title.Color = maroon
	title.Alignment = text.Start
	title.Layout(gtx)
}

func drawGrid(gtx layout.Context) {
	for i := 0; i < gameState.Board.Height; i++ {
		for j := 0; j < gameState.Board.Width; j++ {
			drawCell(cellSizeDp, gtx, j, i, 20, gameState.Board.Cells[i][j])
		}
	}
}

func drawCell(cellSize int, gtx layout.Context, cellX int, cellY int, round int, cell Cell) {
	x0 := gtx.Dp(unit.Dp(cellX * cellSize))
	y0 := gtx.Dp(unit.Dp(cellY * cellSize))
	x1 := gtx.Dp(unit.Dp((cellX + 1) * cellSize))
	y1 := gtx.Dp(unit.Dp((cellY + 1) * cellSize))

	rect := clip.RRect{
		Rect: image.Rect(x0, y0, x1, y1),
		SE:   round,
		SW:   round,
		NW:   round,
		NE:   round,
	}.Op(gtx.Ops)

	cellColor := getColor(cell)

	paint.FillShape(gtx.Ops, cellColor, rect)
}

var emptyColor = color.NRGBA{R: 0, G: 0, B: 0, A: 0}
var redColor = color.NRGBA{R: 255, G: 0, B: 0, A: 255}
var yellowColor = color.NRGBA{R: 255, G: 255, B: 0, A: 255}
var greenColor = color.NRGBA{R: 0, G: 255, B: 0, A: 255}
var blueColor = color.NRGBA{R: 0, G: 0, B: 255, A: 255}
var purpleColor = color.NRGBA{R: 128, G: 0, B: 128, A: 255}
var orangeColor = color.NRGBA{R: 255, G: 165, B: 0, A: 255}

func getColor(cell Cell) color.NRGBA {
	switch cell {
	case Empty:
		return emptyColor
	case Red:
		return redColor
	case Yellow:
		return yellowColor
	case Green:
		return greenColor
	case Blue:
		return blueColor
	case Purple:
		return purpleColor
	case Orange:
		return orangeColor
	default:
		panic("Invalid cell")
	}

}

func run(window *app.Window) error {

	tickChannel := make(chan int)

	go tickEmitter(tickChannel)
	go invalidator(tickChannel, window)
	draw(window)

	return nil
}
