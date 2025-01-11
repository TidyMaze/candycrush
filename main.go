package main

import (
	"fmt"
	"gioui.org/app"
	"gioui.org/f32"
	"gioui.org/io/event"
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

	targetLocation := f32.Point{X: 0, Y: 0}

	for {
		switch e := window.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)

			//println(fmt.Sprintf("Drawing frame %d", displayedTick))

			//drawGrid(gtx)
			//drawCircle(0, 0, gtx, redColor, 50)

			//windowWidth := gtx.Dp(cellSizeDp) * gameState.Board.Width
			//windowHeight := gtx.Dp(cellSizeDp) * gameState.Board.Height

			//drawCircle(windowWidth, windowHeight, gtx, redColor, 50)
			//drawCircles(gtx)

			// print the mouse position

			event.Op(&ops, tag)

			source := e.Source

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
						dragStart = f32.Point{X: -1, Y: -1}
					case pointer.Drag:
						mouseLocation = x.Position
					}
				}
			}

			//println(fmt.Sprintf("Mouse location: %+v", mouseLocation))

			for iBall, _ := range balls {

				ball := &balls[iBall]

				frictionX := float32(0)
				frictionY := float32(0)

				ball.Acceleration.X = rand.Float32()*5 + (targetLocation.X-ball.Location.X)*0.01
				ball.Acceleration.Y = rand.Float32()*5 + (targetLocation.Y-ball.Location.Y)*0.01

				if ball.Velocity.X != 0 || ball.Velocity.Y != 0 {
					velocityMagnitude := float32(math.Sqrt(math.Pow(float64(ball.Velocity.X), 2) + math.Pow(float64(ball.Velocity.Y), 2)))
					frictionForce := 0.01 * velocityMagnitude

					frictionX = -frictionForce * (ball.Velocity.X / velocityMagnitude)
					frictionY = -frictionForce * (ball.Velocity.Y / velocityMagnitude)

					ball.Acceleration.X += frictionX
					ball.Acceleration.Y += frictionY
				}

				ball.Velocity.X += ball.Acceleration.X
				ball.Velocity.Y += ball.Acceleration.Y

				ball.Location.X += ball.Velocity.X + 0.5*ball.Acceleration.X
				ball.Location.Y += ball.Velocity.Y + 0.5*ball.Acceleration.Y

				// reset the acceleration
				ball.Acceleration.X -= frictionX
				ball.Acceleration.Y -= frictionY

				// draw a circle at the ball location
				drawCircle(int(ball.Location.X), int(ball.Location.Y), gtx, ball.color, 50)
			}

			// draw a circle at the mouse location
			color := redColor

			if dragStart.X != -1 && dragStart.Y != -1 {
				distance := distance(dragStart, mouseLocation)

				if pressed {
					color = slightOrange
				}

				if distance > 100 {
					color = slightBlue
					targetLocation = dragStart
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

			e.Frame(gtx.Ops)

			window.Invalidate()
		}
	}
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

	// draw the circle using clip
	ellipse2 := clip.Ellipse{
		// drawing with center at the coordinates
		Min: image.Point{X: int(unit.Dp(x - radius - 5)), Y: int(unit.Dp(y - radius - 5))},
		Max: image.Point{X: int(unit.Dp(x + radius + 5)), Y: int(unit.Dp(y + radius + 5))},
	}

	paint.FillShape(gtx.Ops, color, ellipse2.Op(gtx.Ops))
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
