package testutil

import "github.com/bilus/fencer/primitives"

func Contains(rect primitives.Rect, point primitives.Point) bool {
	return point[0] >= rect.Min[0] && point[0] <= rect.Max[0] &&
		point[1] >= rect.Min[1] && point[1] <= rect.Max[1]
}
