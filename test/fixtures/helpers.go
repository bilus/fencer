package fixtures

import (
	"github.com/JamesMilnerUK/pip-go"
	"github.com/bilus/fencer/primitives"
)

func boundingBoxToRect(boundingRect pip.BoundingBox) (*primitives.Rect, error) {
	return primitives.NewRect(
		primitives.Point{
			boundingRect.BottomLeft.X,
			boundingRect.BottomLeft.Y,
		},
		boundingRect.TopRight.X-boundingRect.BottomLeft.X,
		boundingRect.TopRight.Y-boundingRect.BottomLeft.Y,
	)
}

func makeRect(points []primitives.Point) *primitives.Rect {
	p := points[0]
	op := points[2]
	rect, err := primitives.NewRect(p, op[0]-p[0], op[1]-p[1])
	if err != nil {
		panic(err)
	}
	return rect
}
