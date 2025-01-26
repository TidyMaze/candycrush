package engine

type State struct {
	Board Board
	Score int
}

func (s *State) SwapCells(from, to Coord) {
	old := s.GetCell(from)
	s.SetCell(from, s.GetCell(to))
	s.SetCell(to, old)
}

func (s *State) SetCell(coord Coord, cell Cell) {
	s.Board.SetCell(coord, cell)
}

func (s *State) GetCell(coord Coord) Cell {
	return s.Board.GetCell(coord)
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
			c := Coord{X: j, Y: i}
			newBoard.SetCell(c, s.GetCell(c))
		}
	}

	return State{
		Board: newBoard,
		Score: s.Score,
	}
}
