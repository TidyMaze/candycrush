package engine

type Coord struct {
	X int
	Y int
}

func GetNeighbor(dir Direction, from Coord) Coord {
	offset := DirToOffset(dir)
	destX := from.X + offset.X
	destY := from.Y + offset.Y
	dest := Coord{X: destX, Y: destY}
	return dest
}
