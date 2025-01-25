package engine

type State struct {
	Board Board
	Score int
}

func (s *State) SetCell(x, y int, cell Cell) {
	s.Board.SetCell(x, y, cell)
}

func (s *State) GetCell(x, y int) Cell {
	return s.Board.GetCell(x, y)
}

func (s *State) Width() int {
	return s.Board.Width
}

func (s *State) Height() int {
	return s.Board.Height
}

func (s *State) clone() State {
	// deep copy
	newBoard := Board{
		Width:  s.Width(),
		Height: s.Height(),
		Cells:  make([][]Cell, s.Height()),
	}

	for i := 0; i < s.Height(); i++ {
		newBoard.Cells[i] = make([]Cell, s.Width())
		for j := 0; j < s.Width(); j++ {
			newBoard.SetCell(j, i, s.GetCell(j, i))
		}
	}

	return State{
		Board: newBoard,
		Score: s.Score,
	}
}
