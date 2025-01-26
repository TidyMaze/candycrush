package engine

import "gioui.org/f32"

type Direction int

const (
	Up Direction = iota
	Down
	Left
	Right
)

func DirToOffset(dir Direction) f32.Point {
	// convert dir to offset
	offset := f32.Point{X: 0, Y: 0}

	switch dir {
	case Up:
		offset = f32.Point{X: 0, Y: -1}
	case Down:
		offset = f32.Point{X: 0, Y: 1}
	case Left:
		offset = f32.Point{X: -1, Y: 0}
	case Right:
		offset = f32.Point{X: 1, Y: 0}
	}
	return offset
}
