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

func buildUI() UI {

	engine := Engine{}

	state := engine.InitRandom()

	engine.state = state

	ui := UI{
		animationStep:  Idle,
		animationSince: time.Now(),
		destroyed:      nil,
		filled:         nil,
		fallen:         nil,
		engine:         engine,
	}

	engine.handleChangedAfterExplode = func(changed bool, exploded [][]bool) {
		if changed {
			ui.animationStep = Explode
			ui.setAnimStart()
			ui.destroyed = exploded

			println(fmt.Sprintf("Setting destroying to true, animationSince: %s", ui.animationSince))
		} else {
			println("Explode and fall until stable finished")
			ui.animationStep = Idle
		}
	}

	engine.handleExplodeFinished = func(fallen [][]bool) {
		ui.animationStep = Fall
		ui.setAnimStart()
		ui.fallen = fallen
	}

	engine.handleExplodeFinishedNoChange = func() {
		ui.animationStep = Idle
	}

	engine.handleFallFinished = func(newFilled [][]bool) {
		ui.filled = newFilled
		ui.setAnimStart()
	}

	engine.handleAddMissingCandies = func() {
		ui.animationStep = Refill
	}

	return ui
}

// ui constants
const cellSizeDp = unit.Dp(75)
const ANIMATION_SLEEP_MS = 200

var theme = material.NewTheme()

const UseStateAsBackgroundColor = false

type AnimationStep int

const (
	Idle AnimationStep = iota
	Swap
	Explode
	Fall
	Refill
)

type UI struct {
	animationStep      AnimationStep
	animationSince     time.Time
	destroyed          [][]bool
	filled             [][]bool
	fallen             [][]bool
	engine             Engine
	lastFramesDuration []time.Duration
	lastFrameTime      time.Time
	clickables         []widget.Clickable
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

func (ui *UI) onDragFar(dragStart, dragEnd f32.Point, gtx layout.Context) {
	println(fmt.Sprintf("Dragged far at %f, %f", dragStart.X, dragStart.Y))

	// find the cell at the dragStart
	cellX := int(gtx.Metric.PxToDp(int(dragStart.X)) / cellSizeDp)
	cellY := int(gtx.Metric.PxToDp(int(dragStart.Y)) / cellSizeDp)

	println(fmt.Sprintf("Cell at %d, %d", cellX, cellY))

	if ui.engine.state.Board.Cells[cellY][cellX] == Empty {
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

	ui.animationStep = Swap

	// swap the 2 cells in state
	ui.engine.state = ui.engine.Swap(ui.engine.state, cellX, cellY, cellX+int(offset.X), cellY+int(offset.Y))

	// schedule onSwapFinished for later (1s)
	go func() {
		// sleep for 1s
		time.Sleep(ANIMATION_SLEEP_MS * time.Millisecond)

		ui.onSwapFinished()
	}()
}

func (ui *UI) onSwapFinished() {
	println("Swap finished")
	ui.engine.ExplodeAndFallUntilStable()
}

func (ui *UI) draw(window *app.Window) error {
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

	for {
		switch e := window.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)

			//println(fmt.Sprintf("Drawing frame %d", displayedTick))

			// draw the background (same dimensions as the window), either white or black (destroying)
			backgroundColor := ui.getBackgroundColor()

			drawRect(gtx, 0, 0, int(gtx.Constraints.Max.X), int(gtx.Constraints.Max.Y), backgroundColor)

			ui.drawGrid(gtx)

			//drawCircle(0, 0, gtx, redColor, 50)

			// print the mouse position
			event.Op(&ops, tag)

			source := e.Source

			mouseLocation, pressed, dragStart, alreadySwapped = handleEvents(source, tag, mouseLocation, pressed, dragStart, alreadySwapped)

			//println(fmt.Sprintf("Mouse location: %+v", mouseLocation))

			// draw circle at the drag start location
			if dragStart.X != -1 && dragStart.Y != -1 {
				drawCircle(int(dragStart.X), int(dragStart.Y), gtx, redColor, 10)
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
					println(fmt.Sprintf("Drag threshold reached: %f at %f, %f", distance, mouseLocation.X, mouseLocation.Y))

					if !alreadySwapped {
						ui.onDragFar(dragStart, mouseLocation, gtx)
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
			material.Label(theme, unit.Sp(24), fmt.Sprintf("Score: %d", ui.engine.state.score)).Layout(gtx)

			// draw FPS counter
			fps := computeFPS(ui.lastFramesDuration)

			// add offset for the FPS counter, to the stack
			stack := op.Offset(image.Point{X: 500, Y: 0}).Push(gtx.Ops)

			material.Label(theme, unit.Sp(24), fmt.Sprintf("FPS: %d", fps)).Layout(gtx)

			stack.Pop()

			e.Frame(gtx.Ops)

			// update the FPS counter
			ui.lastFramesDuration = append(ui.lastFramesDuration, time.Since(ui.lastFrameTime))
			keepFrames := 120

			if len(ui.lastFramesDuration) > keepFrames {
				ui.lastFramesDuration = ui.lastFramesDuration[len(ui.lastFramesDuration)-keepFrames:]
			}

			ui.lastFrameTime = time.Now()

			window.Invalidate()
		}
	}
}

func computeFPS(lastFramesDuration []time.Duration) int {
	if len(lastFramesDuration) == 0 {
		return 0
	}

	sum := time.Duration(0)

	for _, duration := range lastFramesDuration {
		sum += duration
	}

	avg := sum / time.Duration(len(lastFramesDuration))

	if avg == 0 {
		return 0
	}

	return int(time.Second / avg)
}

func (ui *UI) getBackgroundColor() color.NRGBA {
	if !UseStateAsBackgroundColor {
		return ccBackgroundColor
	}

	switch ui.animationStep {
	case Idle:
		return ccBackgroundColor
	case Explode:
		return darkRedColor
	case Swap:
		return darkBlueColor
	case Fall:
		return darkGreenColor
	case Refill:
		return darkPurpleColor
	default:
		panic(fmt.Sprintf("Invalid animation step: %d", ui.animationStep))
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

func lerp(outputRangeStart, outputRangeEnd, inputRangeStart, inputRangeEnd, inputRangePosition float64) float64 {
	minDest := math.Min(outputRangeStart, outputRangeEnd)
	maxDest := math.Max(outputRangeStart, outputRangeEnd)

	pct := (inputRangePosition - inputRangeStart) / (inputRangeEnd - inputRangeStart)
	rescaled := outputRangeStart + pct*(outputRangeEnd-outputRangeStart)

	return math.Max(minDest, math.Min(maxDest, rescaled))
}

func (ui *UI) drawGrid(gtx layout.Context) {
	defaultSizePct := 0.95

	//destroyedSizePct := 0.5

	for i := 0; i < ui.engine.state.Board.Height; i++ {
		for j := 0; j < ui.engine.state.Board.Width; j++ {
			sizePct := defaultSizePct

			switch ui.animationStep {
			case Explode:
				if ui.destroyed != nil && ui.destroyed[i][j] {
					// linear interpolation
					sizePct = lerp(defaultSizePct, 0, 0, float64(ANIMATION_SLEEP_MS), float64(time.Since(ui.animationSince).Milliseconds()))
					sizePct = math.Max(0, sizePct)
				}
			case Refill:
				if ui.filled != nil && ui.filled[i][j] {
					sizePct = lerp(0, defaultSizePct, 0, float64(ANIMATION_SLEEP_MS), float64(time.Since(ui.animationSince).Milliseconds()))
				}
			}

			fallPct := float64(1)

			if ui.animationStep == Fall {
				if ui.fallen != nil && ui.fallen[i][j] {
					fallPct = lerp(0, 1, 0, float64(ANIMATION_SLEEP_MS), float64(time.Since(ui.animationSince).Milliseconds()))
				}
			}

			ui.drawCell(cellSizeDp, gtx, j, i, ui.engine.state.Board.Cells[i][j], float32(sizePct), fallPct)
		}
	}
	//print(".")
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

func randRange(min, max int) int {
	return rand.Intn(max-min) + min
}

func toRad(degrees float32) float32 {
	return degrees * math.Pi / 180
}

func (ui *UI) drawCell(cellSize unit.Dp, gtx layout.Context, cellX int, cellY int, cell Cell, sizePct float32, fallPct float64) {

	if cellX < 0 || cellY < 0 {
		panic(fmt.Sprintf("Invalid negative cell position: %d, %d", cellX, cellY))
	}

	cellWidget := CellWidget{
		X:         cellX,
		Y:         cellY,
		Cell:      cell,
		cellSize:  cellSize,
		clickable: &ui.clickables[cellY*ui.engine.state.Board.Width+cellX],
	}

	// offset based on the fallPct (0 is 1 cell up, 1 is the normal position)
	fallOffset := (1 - fallPct) * float64(gtx.Dp(cellSizeDp))

	// size offset
	emptySize := float32(gtx.Dp(cellSizeDp)) * (1 - sizePct)

	cellGlobalX := cellX*gtx.Dp(cellWidget.cellSize) + int(emptySize/2)
	cellGlobalY := cellY*gtx.Dp(cellWidget.cellSize) - int(fallOffset) + int(emptySize/2)

	stack := op.Offset(image.Point{X: cellGlobalX, Y: cellGlobalY}).Push(gtx.Ops)

	defer stack.Pop()

	// draw the square
	cellWidget.clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		// draw the square
		paint.Fill(gtx.Ops, getColor(cell))

		return layout.Dimensions{
			Size: image.Point{
				X: int(float32(gtx.Dp(cellSizeDp)) * sizePct),
				Y: int(float32(gtx.Dp(cellSizeDp)) * sizePct),
			},
		}
	})
}

func drawRect(gtx layout.Context, x, y, width, height int, color color.NRGBA) {
	if width < 0 || height < 0 {
		panic("Invalid negative width or height")
	}

	if width == 0 || height == 0 {
		return
	}

	// offset
	stack := op.Offset(image.Point{X: x, Y: y}).Push(gtx.Ops)

	defer stack.Pop()

	// draw the rect
	paint.Fill(gtx.Ops, color)
}

var ccBackgroundColor = color.NRGBA{R: 45, G: 109, B: 162, A: 255}
var emptyColor = color.NRGBA{R: 0, G: 0, B: 0, A: 0}
var whiteColor = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
var redColor = color.NRGBA{R: 255, G: 0, B: 0, A: 255}
var darkRedColor = color.NRGBA{R: 127, G: 0, B: 0, A: 255}
var yellowColor = color.NRGBA{R: 255, G: 255, B: 0, A: 255}
var greenColor = color.NRGBA{R: 0, G: 220, B: 0, A: 255}
var darkGreenColor = color.NRGBA{R: 0, G: 127, B: 0, A: 255}
var blueColor = color.NRGBA{R: 0, G: 0, B: 255, A: 255}
var darkBlueColor = color.NRGBA{R: 0, G: 0, B: 127, A: 255}
var purpleColor = color.NRGBA{R: 150, G: 0, B: 150, A: 255}
var darkPurpleColor = color.NRGBA{R: 75, G: 0, B: 75, A: 255}
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

func (ui *UI) run(window *app.Window) error {

	ui.draw(window)

	return nil
}

func (ui *UI) setAnimStart() {
	println("Setting animation start")
	ui.animationSince = time.Now()
}

func runUI() {
	ui := buildUI()

	go func() {
		window := new(app.Window)

		window.Option(app.Size(
			unit.Dp(ui.engine.state.Board.Width)*cellSizeDp,
			unit.Dp(ui.engine.state.Board.Height)*cellSizeDp,
		))

		// create clickables
		ui.clickables = make([]widget.Clickable, ui.engine.state.Board.Width*ui.engine.state.Board.Height)

		err := ui.run(window)
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}
