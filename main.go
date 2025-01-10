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
	"gioui.org/widget"
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

const cellSizeDp = unit.Dp(75)

const textSize = unit.Sp(24)

var theme = material.NewTheme()

var clickable = widget.Clickable{}

func main() {
	go func() {
		window := new(app.Window)

		window.Option(app.Size(
			unit.Dp(gameState.Board.Width)*cellSizeDp,
			unit.Dp(gameState.Board.Height)*cellSizeDp,
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
	var ops op.Ops
	for {
		switch e := window.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
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
	//for i := 0; i < gameState.Board.Height; i++ {
	//	for j := 0; j < gameState.Board.Width; j++ {
	//		drawCell(cellSizeDp, gtx, j, i, gameState.Board.Cells[i][j])
	//	}
	//}

	// use the clickable widget to detect clicks on a square
	if clickable.Clicked(gtx) {
		println("Clicked!")
	}

	// offset
	op.Offset(image.Point{X: 30, Y: 30}).Add(gtx.Ops)

	// draw the square
	clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		// draw the square
		paint.Fill(gtx.Ops, color.NRGBA{R: 0x80, G: 0x80, B: 0x80, A: 0xff})

		return layout.Dimensions{
			Size: image.Point{
				X: int(cellSizeDp),
				Y: int(cellSizeDp),
			},
		}
	})
}

type CellWidget struct {
	X, Y      int
	Cell      Cell
	cellSize  unit.Dp
	clickable widget.Clickable
}

func (c *CellWidget) Layout(gtx layout.Context) layout.Dimensions {
	// set the absolute position of the cell
	x0 := gtx.Dp(unit.Dp(c.X) * c.cellSize)
	y0 := gtx.Dp(unit.Dp(c.Y) * c.cellSize)
	x1 := gtx.Dp(unit.Dp(c.X)*c.cellSize + c.cellSize)
	y1 := gtx.Dp(unit.Dp(c.Y)*c.cellSize + c.cellSize)

	round := 20

	rect := clip.RRect{
		Rect: image.Rect(x0, y0, x1, y1),
		SE:   round,
		SW:   round,
		NW:   round,
		NE:   round,
	}.Op(gtx.Ops)

	cellColor := getColor(c.Cell)

	paint.FillShape(gtx.Ops, cellColor, rect)

	// use the Clickable widget to detect clicks

	return layout.Dimensions{
		Size: image.Point{
			X: int(c.cellSize),
			Y: int(c.cellSize),
		},
	}
}

func drawCell(cellSize unit.Dp, gtx layout.Context, cellX int, cellY int, cell Cell) {
	cellWidget := CellWidget{
		X:         cellX,
		Y:         cellY,
		Cell:      cell,
		cellSize:  cellSize,
		clickable: widget.Clickable{},
	}

	cellWidget.Layout(gtx)
}

var emptyColor = color.NRGBA{R: 0, G: 0, B: 0, A: 0}
var redColor = color.NRGBA{R: 255, G: 0, B: 0, A: 255}
var yellowColor = color.NRGBA{R: 255, G: 255, B: 0, A: 255}
var greenColor = color.NRGBA{R: 0, G: 255, B: 0, A: 255}
var blueColor = color.NRGBA{R: 0, G: 0, B: 255, A: 255}
var purpleColor = color.NRGBA{R: 128, G: 0, B: 128, A: 255}
var orangeColor = color.NRGBA{R: 255, G: 165, B: 0, A: 255}
var maroon = color.NRGBA{R: 127, G: 0, B: 0, A: 255}

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
