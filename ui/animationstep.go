package ui

type AnimationStep int

const (
	Idle AnimationStep = iota
	Swap
	Explode
	Fall
	Refill
)
