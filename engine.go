package main

import (
	"fmt"
	"math/rand"
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

type Engine struct{}

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

func (e *Engine) InitRandom() State {

	state := e.Init()

	for i := 0; i < state.Board.Height; i++ {
		for j := 0; j < state.Board.Width; j++ {
			state.Board.Cells[i][j] = e.randomCell()
		}
	}

	state = e.ExplodeAndFallUntilStable(state)

	return state
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

/*
 * Explode candies (if there are 3 or more in a row or column)
 */
func (e *Engine) explode(state State) State {
	score := 0

	// Explode rows
	for i := 0; i < state.Board.Height; i++ {
		for j := 0; j < state.Board.Width-2; j++ {
			if state.Board.Cells[i][j] != Empty && state.Board.Cells[i][j] == state.Board.Cells[i][j+1] && state.Board.Cells[i][j] == state.Board.Cells[i][j+2] {
				state.Board.Cells[i][j] = Empty
				state.Board.Cells[i][j+1] = Empty
				state.Board.Cells[i][j+2] = Empty
				score += 3
			}
		}
	}

	// Explode columns
	for i := 0; i < state.Board.Height-2; i++ {
		for j := 0; j < state.Board.Width; j++ {
			if state.Board.Cells[i][j] != Empty && state.Board.Cells[i][j] == state.Board.Cells[i+1][j] && state.Board.Cells[i][j] == state.Board.Cells[i+2][j] {
				state.Board.Cells[i][j] = Empty
				state.Board.Cells[i+1][j] = Empty
				state.Board.Cells[i+2][j] = Empty
				score += 3
			}
		}
	}

	// Update score
	state.score += score

	return state
}

func (e *Engine) ExplodeWhilePossible(state State) (State, bool) {
	engine := Engine{}

	changed := false

	for {
		newState := engine.explode(state)
		if newState.score == state.score {
			break
		}

		changed = true
		state = newState
		println(fmt.Sprintf("Score: %d", state.score))
	}

	return state, changed
}

/*
Fall candies: move candies down to fill empty cells
*/
func (e *Engine) Fall(state State) (State, bool) {

	changed := false

	for j := 0; j < state.Board.Width; j++ {
		for i := state.Board.Height - 1; i >= 0; i-- {
			if state.Board.Cells[i][j] == Empty {
				for k := i - 1; k >= 0; k-- {
					if state.Board.Cells[k][j] != Empty {
						state.Board.Cells[i][j] = state.Board.Cells[k][j]
						state.Board.Cells[k][j] = Empty
						changed = true
						break
					}
				}
			}
		}
	}

	return state, changed
}

func (e *Engine) ExplodeAndFallUntilStable(gameState State) State {
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
