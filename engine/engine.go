package engine

import (
	"fmt"
	"math/rand"
)

/**
 * Candy crush engine (implemented only game logic)
 */

type Engine struct {
	State                         State
	HandleChangedAfterExplode     func(changed bool, exploded [][]bool)
	HandleExplodeFinished         func(fallen [][]bool)
	HandleExplodeFinishedNoChange func()
	HandleFallFinished            func(newFilled [][]bool)
	HandleAddMissingCandies       func()
	Delay                         func()
}

func (e *Engine) GetCell(x, y int) Cell {
	return e.State.GetCell(x, y)
}

func (e *Engine) SetCell(x, y int, cell Cell) {
	e.State.SetCell(x, y, cell)
}

func (e *Engine) Width() int {
	return e.State.Width()
}

func (e *Engine) Height() int {
	return e.State.Height()
}

func (e *Engine) Init() State {

	width := 9
	height := 9

	board := Board{
		Width:  width,
		Height: height,
		Cells:  make([][]Cell, height),
	}

	for i := 0; i < height; i++ {
		board.Cells[i] = make([]Cell, width)
		for j := 0; j < width; j++ {
			board.SetCell(j, i, Empty)
		}
	}

	return State{
		Board: board,
		Score: 0,
	}
}

func (e *Engine) InitRandom() {

	e.State = e.Init()

	for i := 0; i < e.State.Board.Height; i++ {
		for j := 0; j < e.State.Board.Width; j++ {
			e.setCell(j, i, e.randomCell())
		}
	}

	if e.State.Board.Width <= 0 || e.State.Board.Height <= 0 {
		panic("Invalid board size")
	}

	e.ExplodeAndFallUntilStableSync()

	e.State.Score = 0
}

func (e *Engine) setCell(x, y int, cell Cell) {
	e.State.SetCell(x, y, cell)
}

func (e *Engine) getCell(x, y int) Cell {
	return e.State.GetCell(x, y)
}

func (e *Engine) randomCell() Cell {
	return Cell(rand.Intn(6) + 1)
}

func (e *Engine) Swap(oldState State, x1, y1, x2, y2 int) State {
	state := oldState.clone()

	if x1 < 0 || x1 >= state.Board.Width || y1 < 0 || y1 >= state.Height() {
		return state
	}

	if x2 < 0 || x2 >= state.Board.Width || y2 < 0 || y2 >= state.Height() {
		return state
	}

	old1 := state.GetCell(x1, y1)
	old2 := state.GetCell(x2, y2)

	state.SetCell(x1, y1, old2)
	state.SetCell(x2, y2, old1)

	return state
}

func (e *Engine) findAllExploding(state State) [][]bool {
	exploding := make([][]bool, state.Height())

	for i := 0; i < state.Height(); i++ {
		exploding[i] = make([]bool, state.Width())
	}

	// Explode rows
	for i := 0; i < state.Height(); i++ {
		for j := 0; j < state.Width()-2; j++ {
			if state.GetCell(j, i) != Empty && state.GetCell(j, i) == state.GetCell(j+1, i) && state.GetCell(j, i) == state.GetCell(j+2, i) {
				exploding[i][j] = true
				exploding[i][j+1] = true
				exploding[i][j+2] = true
			}
		}
	}

	// Explode columns
	for i := 0; i < state.Board.Height-2; i++ {
		for j := 0; j < state.Board.Width; j++ {
			if state.GetCell(j, i) != Empty && state.GetCell(j, i) == state.GetCell(j, i+1) && state.GetCell(j, i) == state.GetCell(j, i+2) {
				exploding[i][j] = true
				exploding[i+1][j] = true
				exploding[i+2][j] = true
			}
		}
	}

	return exploding
}

/*
 * Explode candies (if there are 3 or more in a row or column)
 */
func (e *Engine) explode(state State) (State, [][]bool) {
	newState := state.clone()

	score := 0

	exploding := e.findAllExploding(newState)

	// Explode candies
	for i := 0; i < newState.Height(); i++ {
		for j := 0; j < newState.Width(); j++ {
			if exploding[i][j] {
				newState.SetCell(j, i, Empty)
				score++
			}
		}
	}

	// Update score
	newState.Score += score

	return newState, exploding
}

func (e *Engine) ExplodeAndScore(state State) (State, bool, [][]bool) {
	changed := false

	newState, exploded := e.explode(state)
	if newState.Score != state.Score {
		changed = true
		state = newState
	}

	println(fmt.Sprintf("Score: %d", state.Score))

	return state, changed, exploded
}

/*
Fall candies: move candies down to fill empty cells
*/
func (e *Engine) Fall(state State) (State, [][]bool) {

	newState := state.clone()

	fallen := make([][]bool, newState.Height())
	for i := 0; i < newState.Height(); i++ {
		fallen[i] = make([]bool, newState.Width())
	}

	for j := 0; j < newState.Width(); j++ {
		for i := newState.Height() - 1; i >= 0; i-- {
			if newState.GetCell(j, i) == Empty {
				for k := i - 1; k >= 0; k-- {
					if newState.GetCell(j, k) != Empty {
						newState.SetCell(j, i, newState.GetCell(j, k))
						fallen[i][j] = true
						newState.SetCell(j, k, Empty)
						break
					}
				}
			}
		}
	}

	return newState, fallen
}

func (e *Engine) ExplodeAndFallUntilStable() {
	// explode while possible
	newGameState, changed, exploded := e.ExplodeAndScore(e.State)

	if e.HandleChangedAfterExplode != nil {
		e.HandleChangedAfterExplode(changed, exploded)
	}

	if changed {
		go func() {
			if e.Delay != nil {
				e.Delay()
			}
			e.State = newGameState
			e.onExplodeFinished(changed)
		}()
	}
}

func (e *Engine) ExplodeAndFallUntilStableSync() {
	// explode while possible
	for {
		newGameState, changed, _ := e.ExplodeAndScore(e.State)
		e.State = newGameState

		if changed {
			newGameState, _ := e.Fall(e.State)
			e.State = newGameState

			// add missing candies
			newGameState2, _ := e.AddMissingCandies(e.State)
			e.State = newGameState2
		} else {
			println("No more explosions for this loop")
			break
		}
	}

	println("Explode and fall until stable finished")
}

func (e *Engine) onExplodeFinished(explodedChanged bool) {
	println("Explode finished")

	if explodedChanged {
		newGameState, fallen := e.Fall(e.State)

		if e.HandleExplodeFinished != nil {
			e.HandleExplodeFinished(fallen)
		}

		go func() {
			e.State = newGameState
			if e.Delay != nil {
				e.Delay()
			}
			e.onFallFinished()
		}()
	} else {
		e.HandleExplodeFinishedNoChange()
	}
}

func (e *Engine) onFallFinished() {
	println("Fall finished")

	// add missing candies
	newGameState, newFilled := e.AddMissingCandies(e.State)

	e.State = newGameState

	if e.HandleFallFinished != nil {
		e.HandleFallFinished(newFilled)
	}

	go func() {
		if e.Delay != nil {
			e.Delay()
		}
		e.onAddMissingCandiesFinished()
	}()
}

func (e *Engine) onAddMissingCandiesFinished() {
	println("Add missing candies finished")

	go func() {
		if e.Delay != nil {
			e.Delay()
		}
		// explode while possible
		e.ExplodeAndFallUntilStable()
	}()
}

func (e *Engine) AddMissingCandies(state State) (State, [][]bool) {
	if e.HandleAddMissingCandies != nil {
		e.HandleAddMissingCandies()
	}

	newState := state.clone()

	newFilledCellsTmp := make([][]bool, newState.Height())
	for i := 0; i < newState.Height(); i++ {
		newFilledCellsTmp[i] = make([]bool, newState.Width())
	}

	for j := 0; j < newState.Width(); j++ {
		for i := 0; i < newState.Height(); i++ {
			if newState.GetCell(j, i) == Empty {
				newState.SetCell(j, i, e.randomCell())

				newFilledCellsTmp[i][j] = true
			}
		}
	}

	return newState, newFilledCellsTmp
}
