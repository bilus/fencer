package index

import (
	"github.com/bilus/fencer/feature"
	"github.com/bilus/fencer/query"
	"github.com/bilus/rtreego"
	"github.com/paulmach/go.geo"
)

type Point = rtreego.Point

type Index struct {
	*rtreego.Rtree
}

func New(features []feature.Feature) (*Index, error) {
	index := Index{rtreego.NewTree(2, 5, 20)}
	for _, feature := range features {
		index.Insert(feature)
	}
	return &index, nil
}

type MatchContaining struct {
	Point
}

func (mc MatchContaining) IsMatch(feature feature.Feature) (bool, error) {
	return feature.Contains(mc.Point)
}

func (index *Index) FindIntersecting(point Point) ([]feature.Feature, error) {
	p, err := rtreego.NewRect(rtreego.Point(point), []float64{0.01, 0.01})
	if err != nil {
		return nil, err
	}
	candidates := index.SearchIntersect(p)
	if len(candidates) == 0 {
		return nil, nil
	}
	condition := MatchContaining{point}
	features := make([]feature.Feature, 0, len(candidates))
	for _, candidate := range candidates {
		feature := candidate.(feature.Feature)
		match, err := condition.IsMatch(feature)
		if err != nil {
			return nil, err
		}
		if match {
			features = append(features, feature)
		}
	}
	return features, nil
}

func (index *Index) Find(point Point, radiusMeters float64, preconditions []query.Condition, filters []query.Filter, reducer query.Reducer) ([]feature.Feature, error) {
	bounds, err := geomBoundsAround(point, radiusMeters)
	if err != nil {
		return nil, err
	}

	candidates := index.SearchIntersect(bounds)
	if len(candidates) == 0 {
		return nil, nil
	}

	query := query.New(preconditions, filters, reducer)
	for _, candidate := range candidates {
		err := query.Scan(candidate.(feature.Feature))
		if err != nil {
			return nil, err
		}
	}
	return query.MatchingFeatures(), nil
}

func geomBoundsAround(point Point, radiusMeters float64) (*rtreego.Rect, error) {
	bound := geo.NewGeoBoundAroundPoint(geo.NewPoint(point[0], point[1]), radiusMeters)
	tl := bound.SouthWest()
	return rtreego.NewRect(rtreego.Point{tl[0], tl[1]}, []float64{bound.Width(), bound.Height()})
}
