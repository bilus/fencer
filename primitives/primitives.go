// Package primitives contains basic spatial types.
package primitives

// Point represents a 2D point.
type Point = [2]float64

// Rect represents a bounding rectangle.
type Rect struct {
	Min, Max Point
}

func NewRect(minPoint Point, width, height float64) (*Rect, error) {
	return &Rect{
		Min: Point{minPoint[0], minPoint[1]},
		Max: Point{minPoint[0] + width, minPoint[1] + height},
	}, nil
}
