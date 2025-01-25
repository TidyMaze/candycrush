package main

import (
	"fmt"
	"math/rand"
	"time"
)

/**
 * Candy crush engine (implemented only game logic)
 */

type Cell int

const (
	Empty Cell = iota
	Red
	Yellow
	Green
	Blue
	Purple
	Orange
)

type Board struct {
	Width  int
	Height int
	Cells  [][]Cell
}

type State struct {
	Board Board
	score int
}

func (s *State) clone() State {
	// deep copy
	newBoard := Board{
		Width:  s.Board.Width,
		Height: s.Board.Height,
		Cells:  make([][]Cell, s.Board.Height),
	}

	for i := 0; i < s.Board.Height; i++ {
		newBoard.Cells[i] = make([]Cell, s.Board.Width)
		for j := 0; j < s.Board.Width; j++ {
			newBoard.Cells[i][j] = s.Board.Cells[i][j]
		}
	}

	return State{
		Board: newBoard,
		score: s.score,
	}
}

type Engine struct {
	state                         State
	handleChangedAfterExplode     func(changed bool, exploded [][]bool)
	handleExplodeFinished         func(fallen [][]bool)
	handleExplodeFinishedNoChange func()
	handleFallFinished            func(newFilled [][]bool)
	handleAddMissingCandies       func()
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
			board.Cells[i][j] = Empty
		}
	}

	return State{
		Board: board,
		score: 0,
	}
}

func (e *Engine) InitRandom() {

	state := e.Init()

	for i := 0; i < state.Board.Height; i++ {
		for j := 0; j < state.Board.Width; j++ {
			state.Board.Cells[i][j] = e.randomCell()
		}
	}

	if state.Board.Width <= 0 || state.Board.Height <= 0 {
		panic("Invalid board size")
	}

	e.ExplodeAndFallUntilStableSync()

	state.score = 0
}

func (e *Engine) randomCell() Cell {
	return Cell(rand.Intn(6) + 1)
}

func (e *Engine) Swap(state State, x1, y1, x2, y2 int) State {
	if x1 < 0 || x1 >= state.Board.Width || y1 < 0 || y1 >= state.Board.Height {
		return state
	}

	if x2 < 0 || x2 >= state.Board.Width || y2 < 0 || y2 >= state.Board.Height {
		return state
	}

	state.Board.Cells[y1][x1], state.Board.Cells[y2][x2] = state.Board.Cells[y2][x2], state.Board.Cells[y1][x1]
	return state
}

type Coord struct {
	x int
	y int
}

func (e *Engine) findAllTouching(state State, x, y int) []Coord {
	explored := make([]bool, state.Board.Width*state.Board.Height)
	touching := make([]Coord, 0)

	q := make([]Coord, 0)

	q = append(q, Coord{x: x, y: y})

	for len(q) > 0 {
		current := q[0]
		q = q[1:]

		if current.x < 0 || current.x >= state.Board.Width || current.y < 0 || current.y >= state.Board.Height {
			continue
		}

		if explored[current.y*state.Board.Width+current.x] {
			continue
		}

		explored[current.y*state.Board.Width+current.x] = true

		if state.Board.Cells[current.y][current.x] == state.Board.Cells[y][x] {
			touching = append(touching, current)
			q = append(q, Coord{x: current.x - 1, y: current.y})
			q = append(q, Coord{x: current.x + 1, y: current.y})
			q = append(q, Coord{x: current.x, y: current.y - 1})
			q = append(q, Coord{x: current.x, y: current.y + 1})
		}

	}

	return touching
}

func (e *Engine) findAllExploding(state State) [][]bool {
	exploding := make([][]bool, state.Board.Height)

	for i := 0; i < state.Board.Height; i++ {
		exploding[i] = make([]bool, state.Board.Width)
	}

	// Explode rows
	for i := 0; i < state.Board.Height; i++ {
		for j := 0; j < state.Board.Width-2; j++ {
			if state.Board.Cells[i][j] != Empty && state.Board.Cells[i][j] == state.Board.Cells[i][j+1] && state.Board.Cells[i][j] == state.Board.Cells[i][j+2] {
				exploding[i][j] = true
				exploding[i][j+1] = true
				exploding[i][j+2] = true
			}
		}
	}

	// Explode columns
	for i := 0; i < state.Board.Height-2; i++ {
		for j := 0; j < state.Board.Width; j++ {
			if state.Board.Cells[i][j] != Empty && state.Board.Cells[i][j] == state.Board.Cells[i+1][j] && state.Board.Cells[i][j] == state.Board.Cells[i+2][j] {
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
	for i := 0; i < newState.Board.Height; i++ {
		for j := 0; j < newState.Board.Width; j++ {
			if exploding[i][j] {
				newState.Board.Cells[i][j] = Empty
				score++
			}
		}
	}

	// Update score
	newState.score += score

	return newState, exploding
}

func (e *Engine) ExplodeAndScore(state State) (State, bool, [][]bool) {
	changed := false

	newState, exploded := e.explode(state)
	if newState.score != state.score {
		changed = true
		state = newState
	}

	println(fmt.Sprintf("Score: %d", state.score))

	return state, changed, exploded
}

/*
Fall candies: move candies down to fill empty cells
*/
func (e *Engine) Fall(state State) (State, [][]bool) {

	newState := state.clone()

	fallen := make([][]bool, newState.Board.Height)
	for i := 0; i < newState.Board.Height; i++ {
		fallen[i] = make([]bool, newState.Board.Width)
	}

	for j := 0; j < newState.Board.Width; j++ {
		for i := newState.Board.Height - 1; i >= 0; i-- {
			if newState.Board.Cells[i][j] == Empty {
				for k := i - 1; k >= 0; k-- {
					if newState.Board.Cells[k][j] != Empty {
						newState.Board.Cells[i][j] = newState.Board.Cells[k][j]
						fallen[i][j] = true
						newState.Board.Cells[k][j] = Empty
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
	newGameState, changed, exploded := e.ExplodeAndScore(e.state)

	if e.handleChangedAfterExplode != nil {
		e.handleChangedAfterExplode(changed, exploded)
	}

	if changed {
		go func() {
			time.Sleep(ANIMATION_SLEEP_MS * time.Millisecond)
			e.state = newGameState
			e.onExplodeFinished(changed)
		}()
	}
}

func (e *Engine) ExplodeAndFallUntilStableSync() {
	// explode while possible
	for {
		newGameState, changed, _ := e.ExplodeAndScore(e.state)
		e.state = newGameState

		if changed {
			newGameState, _ := e.Fall(e.state)
			e.state = newGameState

			// add missing candies
			newGameState2, _ := e.AddMissingCandies(e.state)
			e.state = newGameState2
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
		newGameState, fallen := e.Fall(e.state)

		if e.handleExplodeFinished != nil {
			e.handleExplodeFinished(fallen)
		}

		go func() {
			e.state = newGameState

			time.Sleep(ANIMATION_SLEEP_MS * time.Millisecond)
			e.onFallFinished()
		}()
	} else {
		e.handleExplodeFinishedNoChange()
	}
}

func (e *Engine) onFallFinished() {
	println("Fall finished")

	// add missing candies
	newGameState, newFilled := e.AddMissingCandies(e.state)

	e.state = newGameState

	if e.handleFallFinished != nil {
		e.handleFallFinished(newFilled)
	}

	go func() {
		time.Sleep(ANIMATION_SLEEP_MS * time.Millisecond)
		e.onAddMissingCandiesFinished()
	}()
}

func (e *Engine) onAddMissingCandiesFinished() {
	println("Add missing candies finished")

	go func() {
		time.Sleep(ANIMATION_SLEEP_MS * time.Millisecond)
		// explode while possible
		e.ExplodeAndFallUntilStable()
	}()
}

func (e *Engine) AddMissingCandies(state State) (State, [][]bool) {
	if e.handleAddMissingCandies != nil {
		e.handleAddMissingCandies()
	}

	newState := state.clone()

	newFilledCellsTmp := make([][]bool, newState.Board.Height)
	for i := 0; i < newState.Board.Height; i++ {
		newFilledCellsTmp[i] = make([]bool, newState.Board.Width)
	}

	for j := 0; j < newState.Board.Width; j++ {
		for i := 0; i < newState.Board.Height; i++ {
			if newState.Board.Cells[i][j] == Empty {
				newState.Board.Cells[i][j] = e.randomCell()

				newFilledCellsTmp[i][j] = true
			}
		}
	}

	return newState, newFilledCellsTmp
}
