package primitives

import (
	"github.com/bilus/rtreego"
)

type Rect rtreego.Rect

func NewRect(p Point, lengths ...float64) (*Rect, error) {
	r, err := rtreego.NewRect(rtreego.Point(p), lengths)
	if err != nil {
		return nil, err
	}
	return (*Rect)(r), nil
}

type Point rtreego.Point

func (p Point) MinDist(r *Rect) float64 {
	return rtreego.Point(p).MinDist((*rtreego.Rect)(r))
}
