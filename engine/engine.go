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
	OnScoreUpdated                func(score int)
}

func (e *Engine) FindValidMoves(state State) []Action {
	var validMoves []Action

	for i := 0; i < state.Height(); i++ {
		for j := 0; j < state.Width(); j++ {
			if i > 0 {
				validMoves = append(validMoves, Action{From: Coord{X: j, Y: i}, To: Coord{X: j, Y: i - 1}})
			}
			if i < state.Height()-1 {
				validMoves = append(validMoves, Action{From: Coord{X: j, Y: i}, To: Coord{X: j, Y: i + 1}})
			}
			if j > 0 {
				validMoves = append(validMoves, Action{From: Coord{X: j, Y: i}, To: Coord{X: j - 1, Y: i}})
			}
			if j < state.Width()-1 {
				validMoves = append(validMoves, Action{From: Coord{X: j, Y: i}, To: Coord{X: j + 1, Y: i}})
			}
		}
	}

	return validMoves
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
			c := Coord{X: j, Y: i}
			board.SetCell(c, Empty)
		}
	}

	return State{
		Board: board,
		Score: 0,
	}
}

func (e *Engine) InitRandom() {

	e.State = e.Init()

	for i := 0; i < e.State.Height(); i++ {
		for j := 0; j < e.State.Width(); j++ {
			c := Coord{X: j, Y: i}
			e.State.SetCell(c, e.randomCell())
		}
	}

	if e.State.Board.Width <= 0 || e.State.Board.Height <= 0 {
		panic("Invalid board size")
	}

	e.ExplodeAndFallUntilStableSync()

	e.State.Score = 0
}

func (e *Engine) randomCell() Cell {
	return Cell(rand.Intn(6) + 1)
}

func (e *Engine) isValidAction(action Action) error {
	if action.From.X < 0 || action.From.X >= e.State.Width() || action.From.Y < 0 || action.From.Y >= e.State.Height() {
		return fmt.Errorf("invalid action: %v: out of bounds (from)", action)
	}

	if action.To.X < 0 || action.To.X >= e.State.Width() || action.To.Y < 0 || action.To.Y >= e.State.Height() {
		return fmt.Errorf("invalid action: %v: out of bounds (to)", action)
	}

	if action.From.X == action.To.X && action.From.Y == action.To.Y {
		return fmt.Errorf("invalid action: %v: same cell", action)
	}

	if action.From.X != action.To.X && action.From.Y != action.To.Y {
		return fmt.Errorf("invalid action: %v: not adjacent", action)
	}

	if e.State.GetCell(action.From) == Empty || e.State.GetCell(action.To) == Empty {
		return fmt.Errorf("invalid action: %v: empty cell", action)
	}

	return nil
}

func (e *Engine) Swap(action Action) State {
	if err := e.isValidAction(action); err != nil {
		println(fmt.Sprintf("Invalid action: %v: %v", action, err))
		return e.State
	}

	state := e.State.clone()
	state.SwapCells(action.From, action.To)

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
			c := Coord{X: j, Y: i}
			c2 := Coord{X: j + 1, Y: i}
			c3 := Coord{X: j + 2, Y: i}
			if state.GetCell(c) != Empty && state.GetCell(c) == state.GetCell(c2) && state.GetCell(c) == state.GetCell(c3) {
				exploding[i][j] = true
				exploding[i][j+1] = true
				exploding[i][j+2] = true
			}
		}
	}

	// Explode columns
	for i := 0; i < state.Board.Height-2; i++ {
		for j := 0; j < state.Board.Width; j++ {
			c := Coord{X: j, Y: i}
			c2 := Coord{X: j, Y: i + 1}
			c3 := Coord{X: j, Y: i + 2}
			if state.GetCell(c) != Empty && state.GetCell(c) == state.GetCell(c2) && state.GetCell(c) == state.GetCell(c3) {
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
			c := Coord{X: j, Y: i}
			if exploding[i][j] {
				newState.SetCell(c, Empty)
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
			c := Coord{X: j, Y: i}
			if newState.GetCell(c) == Empty {
				for k := i - 1; k >= 0; k-- {
					c2 := Coord{X: j, Y: k}
					if newState.GetCell(c2) != Empty {
						newState.SetCell(c, newState.GetCell(c2))
						fallen[i][j] = true
						newState.SetCell(c2, Empty)
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
			e.Delay()
			e.State = newGameState
			e.OnScoreUpdated(e.State.Score)
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
			e.Delay()
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
	e.HandleFallFinished(newFilled)

	go func() {
		e.Delay()
		e.onAddMissingCandiesFinished()
	}()
}

func (e *Engine) onAddMissingCandiesFinished() {
	println("Add missing candies finished")

	go func() {
		e.Delay()
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
			c := Coord{X: j, Y: i}
			if newState.GetCell(c) == Empty {
				newState.SetCell(c, e.randomCell())

				newFilledCellsTmp[i][j] = true
			}
		}
	}

	return newState, newFilledCellsTmp
}
