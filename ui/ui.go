package ui

import (
	"candycrush/engine"
	"candycrush/utils"
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
	"os"
	"time"
)

// ui constants
const cellSizeDp = unit.Dp(75)
const AnimationSleepMs = 200

var theme = material.NewTheme()

const UseStateAsBackgroundColor = true

func BuildUI(state *engine.State) *UI {

	ui := UI{
		animationStep:  Idle,
		AnimationSince: time.Now(),
		Destroyed:      nil,
		Filled:         nil,
		Fallen:         nil,
	}

	ui.Delay = func() {
		println(fmt.Sprintf("Sleeping for %d ms", AnimationSleepMs))
		time.Sleep(AnimationSleepMs * time.Millisecond)
	}

	ui.state = state

	return &ui
}

type UI struct {
	animationStep      AnimationStep
	AnimationSince     time.Time
	Destroyed          [][]bool
	Filled             [][]bool
	Fallen             [][]bool
	lastFramesDuration []time.Duration
	lastFrameTime      time.Time
	clickables         []widget.Clickable
	OnSwap             func(fromX, fromY, toX, toY int)
	Delay              func()
	OnSwapFinished     func()
	score              int
	state              *engine.State
}

func (ui *UI) onDragFar(dragStart, dragEnd f32.Point, gtx layout.Context) {
	println(fmt.Sprintf("Dragged far at %f, %f", dragStart.X, dragStart.Y))

	// find the cell at the dragStart
	cellX := int(gtx.Metric.PxToDp(int(dragStart.X)) / cellSizeDp)
	cellY := int(gtx.Metric.PxToDp(int(dragStart.Y)) / cellSizeDp)

	println(fmt.Sprintf("Cell at %d, %d", cellX, cellY))

	// find the main drag direction (up, down, left, right)

	verticalDiff := float64(dragEnd.Y - dragStart.Y)
	horizontalDiff := float64(dragEnd.X - dragStart.X)

	dir := engine.Direction(-1)

	if math.Abs(verticalDiff) > math.Abs(horizontalDiff) {
		// vertical drag
		if verticalDiff > 0 {
			println("Down")
			dir = engine.Down
		} else {
			println("Up")
			dir = engine.Up
		}
	} else {
		// horizontal drag
		if horizontalDiff > 0 {
			println("Right")
			dir = engine.Right
		} else {
			println("Left")
			dir = engine.Left
		}
	}

	if dir == -1 {
		panic("Invalid direction")
	}

	// convert dir to offset
	offset := f32.Point{X: 0, Y: 0}

	switch dir {
	case engine.Up:
		offset = f32.Point{X: 0, Y: -1}
	case engine.Down:
		offset = f32.Point{X: 0, Y: 1}
	case engine.Left:
		offset = f32.Point{X: -1, Y: 0}
	case engine.Right:
		offset = f32.Point{X: 1, Y: 0}
	}

	ui.SetAnimStep(Swap)

	destX := cellX + int(offset.X)
	destY := cellY + int(offset.Y)

	// swap the 2 cells in state
	ui.OnSwap(cellX, cellY, destX, destY)

	// schedule onSwapFinished for later (1s)
	go func() {
		if ui.Delay != nil {
			ui.Delay()
		}
		ui.OnSwapFinished()
	}()
}

func (ui *UI) SetScore(score int) {
	println(fmt.Sprintf("Setting score to %d", score))
	ui.score = score
}

func (ui *UI) draw(window *app.Window) error {
	var ops op.Ops

	tag := new(bool)

	var mouseLocation f32.Point

	pressed := false

	dragStart := f32.Point{X: -1, Y: -1}

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
				distance := utils.Distance(dragStart, mouseLocation)

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
				}
			} else {
				drawCircle(int(mouseLocation.X), int(mouseLocation.Y), gtx, slightRed, 10)
			}

			// draw the score with size
			material.Label(theme, unit.Sp(24), fmt.Sprintf("Score: %d", ui.score)).Layout(gtx)

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

func (ui *UI) drawGrid(gtx layout.Context) {
	defaultSizePct := 0.95

	//destroyedSizePct := 0.5

	for i := 0; i < ui.Height(); i++ {
		for j := 0; j < ui.Width(); j++ {
			sizePct := defaultSizePct

			switch ui.animationStep {
			case Explode:
				if ui.Destroyed != nil && ui.Destroyed[i][j] {
					// linear interpolation
					sizePct = utils.Lerp(defaultSizePct, 0, 0, float64(AnimationSleepMs), float64(time.Since(ui.AnimationSince).Milliseconds()))
					sizePct = math.Max(0, sizePct)
				}
			case Refill:
				if ui.Filled != nil && ui.Filled[i][j] {
					sizePct = utils.Lerp(0, defaultSizePct, 0, float64(AnimationSleepMs), float64(time.Since(ui.AnimationSince).Milliseconds()))
				}
			}

			fallPct := float64(1)

			if ui.animationStep == Fall {
				if ui.Fallen != nil && ui.Fallen[i][j] {
					fallPct = utils.Lerp(0, 1, 0, float64(AnimationSleepMs), float64(time.Since(ui.AnimationSince).Milliseconds()))
				}
			}

			ui.drawCell(cellSizeDp, gtx, j, i, ui.state.GetCell(j, i), float32(sizePct), fallPct)
		}
	}
	//print(".")
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

func (ui *UI) drawCell(cellSize unit.Dp, gtx layout.Context, cellX int, cellY int, cell engine.Cell, sizePct float32, fallPct float64) {

	if cellX < 0 || cellY < 0 {
		panic(fmt.Sprintf("Invalid negative cell position: %d, %d", cellX, cellY))
	}

	clickable := &ui.clickables[cellY*ui.Width()+cellX]

	// offset based on the fallPct (0 is 1 cell up, 1 is the normal position)
	fallOffset := (1 - fallPct) * float64(gtx.Dp(cellSizeDp))

	// size offset
	emptySize := float32(gtx.Dp(cellSizeDp)) * (1 - sizePct)

	cellGlobalX := cellX*gtx.Dp(cellSize) + int(emptySize/2)
	cellGlobalY := cellY*gtx.Dp(cellSize) - int(fallOffset) + int(emptySize/2)

	stack := op.Offset(image.Point{X: cellGlobalX, Y: cellGlobalY}).Push(gtx.Ops)

	defer stack.Pop()

	// draw the square
	clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
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

func getColor(cell engine.Cell) color.NRGBA {
	switch cell {
	case engine.Empty:
		return emptyColor
	case engine.Red:
		return redColor
	case engine.Yellow:
		return yellowColor
	case engine.Green:
		return greenColor
	case engine.Blue:
		return blueColor
	case engine.Purple:
		return purpleColor
	case engine.Orange:
		return orangeColor
	default:
		panic("Invalid cell")
	}

}

func (ui *UI) run(window *app.Window) error {
	ui.draw(window)
	return nil
}

func (ui *UI) SetAnimStart() {
	println("Setting animation start")
	ui.AnimationSince = time.Now()
}

func (ui *UI) SetAnimStep(step AnimationStep) {
	println(fmt.Sprintf("Setting animation step to %s", showAnimationStep(step)))
	ui.animationStep = step
}

func (ui *UI) Width() int {
	return ui.state.Width()
}

func (ui *UI) Height() int {
	return ui.state.Height()
}

func showAnimationStep(step AnimationStep) string {
	switch step {
	case Idle:
		return "Idle"
	case Swap:
		return "Swap"
	case Explode:
		return "Explode"
	case Fall:
		return "Fall"
	case Refill:
		return "Refill"
	default:
		panic(fmt.Sprintf("Invalid animation step: %d", step))
	}
}

func RunUI(ui *UI) {
	if ui.Width() <= 0 || ui.Height() <= 0 {
		panic(fmt.Sprintf("Invalid board dimensions: %d, %d", ui.Width(), ui.Height()))
	}

	go func() {
		window := new(app.Window)

		window.Option(app.Size(
			unit.Dp(ui.Width())*cellSizeDp,
			unit.Dp(ui.Height())*cellSizeDp,
		))

		// create clickables
		ui.clickables = make([]widget.Clickable, ui.Width()*ui.Height())

		err := ui.run(window)
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}
