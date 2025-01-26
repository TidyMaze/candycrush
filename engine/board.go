package engine

type Board struct {
	Width  int
	Height int
	Cells  [][]Cell
}

func (b *Board) SetCell(coord Coord, cell Cell) {
	b.Cells[coord.Y][coord.X] = cell
}

func (b *Board) GetCell(coord Coord) Cell {
	return b.Cells[coord.Y][coord.X]
}
