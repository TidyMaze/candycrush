package engine

type Direction int

const (
	Up Direction = iota
	Down
	Left
	Right
)

func DirToOffset(dir Direction) Coord {
	// convert dir to offset
	offset := Coord{X: 0, Y: 0}

	switch dir {
	case Up:
		offset = Coord{X: 0, Y: -1}
	case Down:
		offset = Coord{X: 0, Y: 1}
	case Left:
		offset = Coord{X: -1, Y: 0}
	case Right:
		offset = Coord{X: 1, Y: 0}
	}
	return offset
}
