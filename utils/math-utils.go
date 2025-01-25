package utils

import (
	"gioui.org/f32"
	"math"
)

func Distance(a, b f32.Point) float64 {
	return math.Sqrt(math.Pow(float64(a.X-b.X), 2) + math.Pow(float64(a.Y-b.Y), 2))
}

func Lerp(outputRangeStart, outputRangeEnd, inputRangeStart, inputRangeEnd, inputRangePosition float64) float64 {
	minDest := math.Min(outputRangeStart, outputRangeEnd)
	maxDest := math.Max(outputRangeStart, outputRangeEnd)

	pct := (inputRangePosition - inputRangeStart) / (inputRangeEnd - inputRangeStart)
	rescaled := outputRangeStart + pct*(outputRangeEnd-outputRangeStart)

	return math.Max(minDest, math.Min(maxDest, rescaled))
}
