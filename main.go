package main

import (
	"fmt"
	"gioui.org/app"
	"gioui.org/f32"
	"gioui.org/io/event"
	"gioui.org/io/input"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"image"
	"image/color"
	"log"
	"math"
	"math/rand"
	"os"
	"time"
)

var engine Engine = Engine{}
var gameState State = engine.InitRandom()

const cellSizeDp = unit.Dp(75)

const textSize = unit.Sp(24)

var theme = material.NewTheme()

var circles []image.Point
var circlesHovered []image.Point

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

type Ball struct {
	Location     f32.Point
	Velocity     f32.Point
	Acceleration f32.Point
	color        color.NRGBA
}

type Direction int

const (
	Up Direction = iota
	Down
	Left
	Right
)

func onDragFar(dragStart, dragEnd f32.Point, gtx layout.Context) {
	println(fmt.Sprintf("Dragged far at %f, %f", dragStart.X, dragStart.Y))

	// find the cell at the dragStart
	cellX := int(gtx.Metric.PxToDp(int(dragStart.X)) / cellSizeDp)
	cellY := int(gtx.Metric.PxToDp(int(dragStart.Y)) / cellSizeDp)

	println(fmt.Sprintf("Cell at %d, %d", cellX, cellY))

	if gameState.Board.Cells[cellY][cellX] == Empty {
		println("Empty cell, skipping")
		return
	}

	// find the main drag direction (up, down, left, right)

	verticalDiff := float64(dragEnd.Y - dragStart.Y)
	horizontalDiff := float64(dragEnd.X - dragStart.X)

	dir := Direction(-1)

	if math.Abs(verticalDiff) > math.Abs(horizontalDiff) {
		// vertical drag
		if verticalDiff > 0 {
			println("Down")
			dir = Down
		} else {
			println("Up")
			dir = Up
		}
	} else {
		// horizontal drag
		if horizontalDiff > 0 {
			println("Right")
			dir = Right
		} else {
			println("Left")
			dir = Left
		}
	}

	if dir == -1 {
		panic("Invalid direction")
	}

	// convert dir to offset
	offset := f32.Point{X: 0, Y: 0}

	switch dir {
	case Up:
		offset = f32.Point{X: 0, Y: -1}
	case Down:
		offset = f32.Point{X: 0, Y: 1}
	case Left:
		offset = f32.Point{X: -1, Y: 0}
	case Right:
		offset = f32.Point{X: 1, Y: 0}
	}

	// swap the 2 cells in state
	gameState = engine.Swap(gameState, cellX, cellY, cellX+int(offset.X), cellY+int(offset.Y))

	gameState = explodedAndFallUntilStable(gameState)
}

func explodedAndFallUntilStable(gameState State) State {
	changed := true

	for changed {
		changed = false

		// explode while possible
		newGameState, explodedChanged := engine.ExplodeWhilePossible(gameState)
		gameState = newGameState

		if explodedChanged {
			changed = true
			newGameState2, fallChanged := engine.Fall(gameState)
			gameState = newGameState2

			if fallChanged {
				changed = true
			}
		}
	}

	return gameState
}

func draw(window *app.Window) error {
	var ops op.Ops

	tag := new(bool)

	var mouseLocation f32.Point

	pressed := false

	dragStart := f32.Point{X: -1, Y: -1}

	balls := make([]Ball, 0)
	balls = append(balls, Ball{
		Location: f32.Point{X: 0, Y: 0},
		Velocity: f32.Point{X: 0, Y: 0},
	})

	alreadySwapped := false

	go func() {
		for {
			// add a new ball at random location
			balls = append(balls, Ball{
				Location:     f32.Point{X: rand.Float32() * 1000, Y: rand.Float32() * 1000},
				Velocity:     f32.Point{X: 0, Y: 0},
				Acceleration: f32.Point{X: 0, Y: 0},
				color:        randomColor(),
			})

			keepMax := 1000

			if len(balls) > keepMax {
				// keep the last keepMax balls
				balls = balls[len(balls)-keepMax:]
			}

			// sleep for a while
			time.Sleep(100 * time.Millisecond)
		}
	}()

	for {
		switch e := window.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)

			//println(fmt.Sprintf("Drawing frame %d", displayedTick))

			drawGrid(gtx)
			//drawCircle(0, 0, gtx, redColor, 50)

			drawCircles(gtx)

			// print the mouse position

			event.Op(&ops, tag)

			source := e.Source

			mouseLocation, pressed, dragStart, alreadySwapped = handleEvents(source, tag, mouseLocation, pressed, dragStart, alreadySwapped)

			//println(fmt.Sprintf("Mouse location: %+v", mouseLocation))

			// draw circle at the drag start location
			drawCircle(int(dragStart.X), int(dragStart.Y), gtx, redColor, 10)

			// draw a circle at the mouse location
			color := redColor

			if dragStart.X != -1 && dragStart.Y != -1 {
				distance := distance(dragStart, mouseLocation)

				if pressed {
					color = slightOrange
				}

				if distance > 100 {
					color = slightBlue
					println(fmt.Sprintf("Drag threshold reached: %f at %f, %f", distance, mouseLocation.X, mouseLocation.Y))

					if !alreadySwapped {
						onDragFar(dragStart, mouseLocation, gtx)
						alreadySwapped = true
					}
				}

				drawCircle(int(dragStart.X), int(dragStart.Y), gtx, color, int(distance))

				if distance > 200 {
					// reset the drag start
					dragStart = f32.Point{X: -1, Y: -1}

					// add a new ball
					balls = append(balls, Ball{
						Location: f32.Point{X: 0, Y: 0},
						Velocity: f32.Point{X: 0, Y: 0},
						color:    randomColor(),
					})
				}
			} else {
				drawCircle(int(mouseLocation.X), int(mouseLocation.Y), gtx, slightRed, 10)
			}

			// draw the score with size
			material.Label(theme, unit.Sp(24), fmt.Sprintf("Score: %d", gameState.score)).Layout(gtx)

			e.Frame(gtx.Ops)

			window.Invalidate()
		}
	}
}

func handleEvents(source input.Source, tag *bool, mouseLocation f32.Point, pressed bool, dragStart f32.Point, alreadySwapped bool) (f32.Point, bool, f32.Point, bool) {
	for {
		ev, ok := source.Event(pointer.Filter{
			Target: tag,
			Kinds:  pointer.Move | pointer.Press | pointer.Release | pointer.Drag,
		})

		if !ok {
			break
		}

		if x, ok := ev.(pointer.Event); ok {
			switch x.Kind {
			case pointer.Move:
				mouseLocation = x.Position
			case pointer.Press:
				pressed = true
				dragStart = x.Position
			case pointer.Release:
				pressed = false
				alreadySwapped = false
				dragStart = f32.Point{X: -1, Y: -1}
			case pointer.Drag:
				mouseLocation = x.Position
			}
		}
	}
	return mouseLocation, pressed, dragStart, alreadySwapped
}

func randomColor() color.NRGBA {
	return color.NRGBA{
		R: uint8(rand.Intn(256)),
		G: uint8(rand.Intn(256)),
		B: uint8(rand.Intn(256)),
		A: 127,
	}
}

func distance(a, b f32.Point) float64 {
	return math.Sqrt(math.Pow(float64(a.X-b.X), 2) + math.Pow(float64(a.Y-b.Y), 2))
}

func drawCircles(gtx layout.Context) {
	for _, circle := range circles {
		drawCircle(circle.X, circle.Y, gtx, redColor, 50)
	}

	for _, circle := range circlesHovered {
		drawCircle(circle.X, circle.Y, gtx, slightDark, 20)
	}
}

func drawGrid(gtx layout.Context) {
	for i := 0; i < gameState.Board.Height; i++ {
		for j := 0; j < gameState.Board.Width; j++ {
			drawCell(cellSizeDp, gtx, j, i, gameState.Board.Cells[i][j])
		}
	}
}

type CellWidget struct {
	X, Y      int
	Cell      Cell
	cellSize  unit.Dp
	clickable *widget.Clickable
}

func drawCircle(
	x, y int,
	gtx layout.Context,
	color color.NRGBA,
	radius int,
) {
	//println(fmt.Sprintf("Drawing circle at %d, %d", x, y))

	// offset

	// draw the circle using clip
	ellipse := clip.Ellipse{
		// drawing with center at the coordinates
		Min: image.Point{X: int(unit.Dp(x - radius)), Y: int(unit.Dp(y - radius))},
		Max: image.Point{X: int(unit.Dp(x + radius)), Y: int(unit.Dp(y + radius))},
	}

	paint.FillShape(gtx.Ops, color, ellipse.Op(gtx.Ops))
}

var clickables []widget.Clickable

func drawCell(cellSize unit.Dp, gtx layout.Context, cellX int, cellY int, cell Cell) {

	if cellX < 0 || cellY < 0 {
		panic(fmt.Sprintf("Invalid negative cell position: %d, %d", cellX, cellY))
	}

	cellWidget := CellWidget{
		X:         cellX,
		Y:         cellY,
		Cell:      cell,
		cellSize:  cellSize,
		clickable: &clickables[cellY*gameState.Board.Width+cellX],
	}

	// use the clickable widget to detect clicks on a square
	if cellWidget.clickable.Clicked(gtx) {
		location := cellWidget.clickable.History()[0]

		if location.Position.X < 0 || location.Position.Y < 0 {
			panic(fmt.Sprintf("Invalid negative click local position: %+v", location.Position))
		}

		println(fmt.Sprintf("Clicked! %d, %d for cell at coord %d, %d", location.Position.X, location.Position.Y, cellX, cellY))

		// last location
		//last := cellWidget.clickable.History()[0]

		x := unit.Dp(cellX)*cellWidget.cellSize + gtx.Metric.PxToDp(location.Position.X)
		y := unit.Dp(cellY)*cellWidget.cellSize + gtx.Metric.PxToDp(location.Position.Y)

		if x < 0 || y < 0 {
			panic(fmt.Sprintf("Invalid negative click global position: %+v", location.Position))
		}

		// add a circle at the clicked position
		circles = append(circles, image.Point{X: gtx.Dp(x), Y: gtx.Dp(y)})
	}

	if cellWidget.clickable.Hovered() {
		println(fmt.Sprintf("Hovered cell at coord %d, %d", cellX, cellY))

		x := unit.Dp(cellX)*cellWidget.cellSize + cellWidget.cellSize/2
		y := unit.Dp(cellY)*cellWidget.cellSize + cellWidget.cellSize/2

		circlesHovered = append(circlesHovered, image.Point{X: gtx.Dp(x), Y: gtx.Dp(y)})
	}

	// offset
	cellGlobalX := cellX * gtx.Dp(cellWidget.cellSize)
	cellGlobalY := cellY * gtx.Dp(cellWidget.cellSize)

	if cellGlobalX < 0 || cellGlobalY < 0 {
		panic(fmt.Sprintf("Invalid negative global cell position: %d, %d", cellGlobalX, cellGlobalY))
	}

	//print(fmt.Sprintf("Drawing cell at %d, %d\n", cellGlobalX, cellGlobalY))

	stack := op.Offset(image.Point{X: cellGlobalX, Y: cellGlobalY}).Push(gtx.Ops)
	defer stack.Pop()

	// draw the square
	cellWidget.clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		// draw the square
		paint.Fill(gtx.Ops, getColor(cell))

		return layout.Dimensions{
			Size: image.Point{
				X: gtx.Dp(cellSizeDp),
				Y: gtx.Dp(cellSizeDp),
			},
		}
	})
}

var emptyColor = color.NRGBA{R: 0, G: 0, B: 0, A: 0}
var redColor = color.NRGBA{R: 255, G: 0, B: 0, A: 255}
var yellowColor = color.NRGBA{R: 255, G: 255, B: 0, A: 255}
var greenColor = color.NRGBA{R: 0, G: 255, B: 0, A: 255}
var blueColor = color.NRGBA{R: 0, G: 0, B: 255, A: 255}
var purpleColor = color.NRGBA{R: 128, G: 0, B: 128, A: 255}
var orangeColor = color.NRGBA{R: 255, G: 165, B: 0, A: 255}
var maroon = color.NRGBA{R: 127, G: 0, B: 0, A: 255}
var slightDark = color.NRGBA{R: 0, G: 0, B: 0, A: 127}

var slightGreen = color.NRGBA{R: 0, G: 255, B: 0, A: 127}
var slightBlue = color.NRGBA{R: 0, G: 0, B: 255, A: 127}
var slightRed = color.NRGBA{R: 255, G: 0, B: 0, A: 127}
var slightOrange = color.NRGBA{R: 255, G: 165, B: 0, A: 127}

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

	draw(window)

	return nil
}
