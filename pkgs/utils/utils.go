package utils

import (
	"image"
	"math"
)

// Min - returns min between two integer (actually int64) values
func Min(x, y int64) int64 {
	if x < y {
		return x
	}
	return y
}

// Max - returns max between two integer (actually int64) values
func Max(x, y int64) int64 {
	if x > y {
		return x
	}
	return y
}

// DistanceBetweenPoints - returns distance between to points (image.Point)
func DistanceBetweenPoints(p1 image.Point, p2 image.Point) float64 {
	intX := math.Abs(float64(p1.X - p2.X))
	intY := math.Abs(float64(p1.Y - p2.Y))
	return math.Sqrt(math.Pow(intX, 2) + math.Pow(intY, 2))
}
