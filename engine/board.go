package engine

type Board struct {
	Width  int
	Height int
	Cells  [][]Cell
}

func (b *Board) SetCell(x, y int, cell Cell) {
	b.Cells[y][x] = cell
}

func (b *Board) GetCell(x, y int) Cell {
	return b.Cells[y][x]
}
