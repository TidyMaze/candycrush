package main

import "math/rand"

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
