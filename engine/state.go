package engine

type State struct {
	Board Board
	Score int
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
		Score: s.Score,
	}
}
