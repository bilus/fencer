package index_test

import (
	"fmt"
	"github.com/JamesMilnerUK/pip-go"
	"github.com/bilus/fencer/feature"
	"github.com/bilus/fencer/geo"
	"github.com/bilus/fencer/index"
	"github.com/bilus/fencer/primitives"
	"github.com/bilus/fencer/query"
)

// CityID uniquely identifies a city.
type CityID string

func (id CityID) String() string {
	return string(id)
}

// City represents a city's area and its other properties.
type City struct {
	feature.Feature
	ID           CityID
	Name         string
	Population   int
	Boundaries   pip.Polygon
	BoundingRect *primitives.Rect
}

func NewCity(id CityID, name string, population int, boundaries pip.Polygon) (City, error) {
	bounds, err := makeRect(pip.GetBoundingBox(boundaries))
	if err != nil {
		return City{}, err
	}
	return City{
		ID:           id,
		Population:   population,
		Name:         name,
		Boundaries:   boundaries,
		BoundingRect: bounds,
	}, nil
}

// Contains returns true if a point lies within a city's area.
func (r *City) Contains(point primitives.Point) (bool, error) {
	return pip.PointInPolygon(
		pip.Point{
			X: point[0],
			Y: point[1],
		},
		r.Boundaries,
	), nil
}

// Bounds returns the bounding rectangle for a restaurant.
func (r *City) Bounds() *primitives.Rect {
	return r.BoundingRect
}

func (r *City) Key() feature.Key {
	return r.ID
}

func makeRect(boundingRect pip.BoundingBox) (*primitives.Rect, error) {
	return primitives.NewRect(
		primitives.Point{
			boundingRect.BottomLeft.X,
			boundingRect.BottomLeft.Y,
		},
		boundingRect.TopRight.X-boundingRect.BottomLeft.X,
		boundingRect.TopRight.Y-boundingRect.BottomLeft.Y,
	)
}

// This example uses an example spatial feature implementation.
// See https://github.com/bilus/fencer/blob/master/index/index_test.go for more details.
func ExampleIndex_Lookup() {
	wroclaw, _ := NewCity("wrocław", "Wrocław", 638384, pip.Polygon{Points: wroclawBoundaries})
	szczecin, _ := NewCity("szczecin", "Szczecin", 407811, pip.Polygon{Points: szczecinBoundaries})
	index, _ := index.New([]feature.Feature{&wroclaw, &szczecin})
	results, _ := index.Lookup(CityID("wrocław"))
	fmt.Println(len(results), "match")
	// Output: 1 match
}

// This example uses an example spatial feature implementation.
// See https://github.com/bilus/fencer/blob/master/index/index_test.go for more details.
func ExampleIndex_FindContaining() {
	wroclaw, _ := NewCity("wrocław", "Wrocław", 638384, pip.Polygon{Points: wroclawBoundaries})
	szczecin, _ := NewCity("szczecin", "Szczecin", 407811, pip.Polygon{Points: szczecinBoundaries})
	index, _ := index.New([]feature.Feature{&wroclaw, &szczecin})
	location := primitives.Point{14.499678611755371, 53.41209631751399}
	results, _ := index.FindContaining(location)
	fmt.Println(len(results), "result:", results[0].(*City).Name)
	// Output: 1 result: Szczecin
}

// This example uses an example spatial feature implementation.
// See https://github.com/bilus/fencer/blob/master/index/index_test.go for more details.
func ExampleIndex_Intersect() {
	wroclaw, _ := NewCity("wrocław", "Wrocław", 638384, pip.Polygon{Points: wroclawBoundaries})
	szczecin, _ := NewCity("szczecin", "Szczecin", 407811, pip.Polygon{Points: szczecinBoundaries})
	index, _ := index.New([]feature.Feature{&wroclaw, &szczecin})
	location := primitives.Point{14.499678611755371, 53.41209631751399}
	// A 1000kmx1000km bounding rectangle around the location so we match both cities.
	radius := 500000.0
	bounds, _ := geo.NewBoundsAround(location, radius)
	results, _ := index.Intersect(bounds)
	fmt.Println(len(results), "results")
	// Output: 2 results
}

type PopulationGreaterThan struct {
	Threshold int
}

func (c PopulationGreaterThan) IsMatch(feature feature.Feature) (bool, error) {
	return feature.(*City).Population > c.Threshold, nil
}

// This example uses an example spatial feature implementation.
// See https://github.com/bilus/fencer/blob/master/index/index_test.go for more details.
func ExampleIndex_Query_preconditions() {
	wroclaw, _ := NewCity("wrocław", "Wrocław", 638384, pip.Polygon{Points: wroclawBoundaries})
	szczecin, _ := NewCity("szczecin", "Szczecin", 407811, pip.Polygon{Points: szczecinBoundaries})
	index, _ := index.New([]feature.Feature{&wroclaw, &szczecin})
	location := primitives.Point{14.499678611755371, 53.41209631751399}
	// A 1000kmx1000km bounding rectangle around the location so we match both cities.
	radius := 500000.0
	bounds, _ := geo.NewBoundsAround(location, radius)
	results, _ := index.Query(bounds, query.Build().Precondition(PopulationGreaterThan{500000}).Query())
	fmt.Println(len(results), "result")
	// Output: 1 result
}

// Rough outlines of two example cities.
var (
	wroclawBoundaries = []pip.Point{
		{X: 16.806, Y: 51.135}, {X: 16.804, Y: 51.136}, {X: 16.803, Y: 51.14}, {X: 16.81, Y: 51.153},
		{X: 16.814, Y: 51.154}, {X: 16.814, Y: 51.156}, {X: 16.816, Y: 51.157}, {X: 16.824, Y: 51.156},
		{X: 16.823, Y: 51.161}, {X: 16.817, Y: 51.169}, {X: 16.817, Y: 51.173}, {X: 16.82, Y: 51.176},
		{X: 16.819, Y: 51.177}, {X: 16.821, Y: 51.182}, {X: 16.827, Y: 51.188}, {X: 16.83, Y: 51.187},
		{X: 16.838, Y: 51.188}, {X: 16.839, Y: 51.186}, {X: 16.844, Y: 51.187}, {X: 16.853, Y: 51.182},
		{X: 16.86, Y: 51.182}, {X: 16.862, Y: 51.184}, {X: 16.87, Y: 51.183}, {X: 16.873, Y: 51.185},
		{X: 16.875, Y: 51.189}, {X: 16.88, Y: 51.192}, {X: 16.883, Y: 51.192}, {X: 16.881, Y: 51.198},
		{X: 16.878, Y: 51.2}, {X: 16.878, Y: 51.205}, {X: 16.886, Y: 51.21}, {X: 16.89, Y: 51.21},
		{X: 16.893, Y: 51.208}, {X: 16.898, Y: 51.213}, {X: 16.906, Y: 51.214}, {X: 16.909, Y: 51.211},
		{X: 16.91, Y: 51.207}, {X: 16.914, Y: 51.207}, {X: 16.921, Y: 51.202}, {X: 16.925, Y: 51.201},
		{X: 16.932, Y: 51.202}, {X: 16.935, Y: 51.204}, {X: 16.936, Y: 51.203}, {X: 16.943, Y: 51.207},
		{X: 16.948, Y: 51.207}, {X: 16.954, Y: 51.213}, {X: 16.963, Y: 51.213}, {X: 16.969, Y: 51.211},
		{X: 16.97, Y: 51.209}, {X: 16.983, Y: 51.209}, {X: 16.992, Y: 51.207}, {X: 16.997, Y: 51.203},
		{X: 17.003, Y: 51.201}, {X: 17.005, Y: 51.199}, {X: 17.005, Y: 51.191}, {X: 17.003, Y: 51.186},
		{X: 17.009, Y: 51.186}, {X: 17.011, Y: 51.19}, {X: 17.014, Y: 51.192}, {X: 17.024, Y: 51.19},
		{X: 17.029, Y: 51.185}, {X: 17.03, Y: 51.179}, {X: 17.043, Y: 51.178}, {X: 17.066, Y: 51.169},
		{X: 17.074, Y: 51.17}, {X: 17.077, Y: 51.178}, {X: 17.084, Y: 51.181}, {X: 17.092, Y: 51.179},
		{X: 17.114, Y: 51.179}, {X: 17.12, Y: 51.176}, {X: 17.124, Y: 51.176}, {X: 17.126, Y: 51.177},
		{X: 17.127, Y: 51.181}, {X: 17.13, Y: 51.184}, {X: 17.16, Y: 51.182}, {X: 17.163, Y: 51.177},
		{X: 17.159, Y: 51.172}, {X: 17.159, Y: 51.16}, {X: 17.156, Y: 51.158}, {X: 17.157, Y: 51.15},
		{X: 17.154, Y: 51.142}, {X: 17.144, Y: 51.128}, {X: 17.151, Y: 51.128}, {X: 17.157, Y: 51.123},
		{X: 17.161, Y: 51.123}, {X: 17.166, Y: 51.12}, {X: 17.168, Y: 51.116}, {X: 17.175, Y: 51.115},
		{X: 17.18, Y: 51.109}, {X: 17.18, Y: 51.105}, {X: 17.175, Y: 51.1}, {X: 17.176, Y: 51.095},
		{X: 17.173, Y: 51.092}, {X: 17.167, Y: 51.092}, {X: 17.166, Y: 51.087}, {X: 17.164, Y: 51.085},
		{X: 17.159, Y: 51.084}, {X: 17.15, Y: 51.077}, {X: 17.142, Y: 51.078}, {X: 17.129, Y: 51.075},
		{X: 17.126, Y: 51.078}, {X: 17.121, Y: 51.078}, {X: 17.118, Y: 51.076}, {X: 17.114, Y: 51.076},
		{X: 17.113, Y: 51.072}, {X: 17.105, Y: 51.064}, {X: 17.104, Y: 51.059}, {X: 17.105, Y: 51.057},
		{X: 17.109, Y: 51.056}, {X: 17.111, Y: 51.049}, {X: 17.104, Y: 51.041}, {X: 17.083, Y: 51.043},
		{X: 17.08, Y: 51.046}, {X: 17.076, Y: 51.047}, {X: 17.07, Y: 51.053}, {X: 17.07, Y: 51.049},
		{X: 17.067, Y: 51.046}, {X: 17.065, Y: 51.046}, {X: 17.065, Y: 51.044}, {X: 17.062, Y: 51.041},
		{X: 17.054, Y: 51.042}, {X: 17.05, Y: 51.04}, {X: 17.039, Y: 51.045}, {X: 17.033, Y: 51.045},
		{X: 17.033, Y: 51.041}, {X: 17.031, Y: 51.039}, {X: 17.026, Y: 51.039}, {X: 17.008, Y: 51.043},
		{X: 17.006, Y: 51.045}, {X: 17.006, Y: 51.049}, {X: 17.003, Y: 51.046}, {X: 16.999, Y: 51.045},
		{X: 16.99, Y: 51.048}, {X: 16.989, Y: 51.05}, {X: 16.978, Y: 51.049}, {X: 16.973, Y: 51.051},
		{X: 16.968, Y: 51.045}, {X: 16.962, Y: 51.046}, {X: 16.951, Y: 51.057}, {X: 16.951, Y: 51.062},
		{X: 16.953, Y: 51.065}, {X: 16.936, Y: 51.074}, {X: 16.935, Y: 51.078}, {X: 16.92, Y: 51.076},
		{X: 16.916, Y: 51.077}, {X: 16.913, Y: 51.08}, {X: 16.919, Y: 51.093}, {X: 16.901, Y: 51.093},
		{X: 16.892, Y: 51.095}, {X: 16.891, Y: 51.091}, {X: 16.887, Y: 51.089}, {X: 16.874, Y: 51.091},
		{X: 16.858, Y: 51.096}, {X: 16.85, Y: 51.106}, {X: 16.837, Y: 51.106}, {X: 16.829, Y: 51.108},
		{X: 16.826, Y: 51.111}, {X: 16.826, Y: 51.114}, {X: 16.832, Y: 51.118}, {X: 16.829, Y: 51.121},
		{X: 16.815, Y: 51.123}, {X: 16.813, Y: 51.125}, {X: 16.815, Y: 51.133}, {X: 16.809, Y: 51.133},
		{X: 16.806, Y: 51.135},
	}

	szczecinBoundaries = []pip.Point{
		{X: 14.43, Y: 53.487}, {X: 14.431, Y: 53.49}, {X: 14.43, Y: 53.497}, {X: 14.432, Y: 53.503}, {X: 14.436, Y: 53.507},
		{X: 14.443, Y: 53.51}, {X: 14.468, Y: 53.511}, {X: 14.472, Y: 53.509}, {X: 14.481, Y: 53.51}, {X: 14.499, Y: 53.506},
		{X: 14.502, Y: 53.509}, {X: 14.512, Y: 53.511}, {X: 14.529, Y: 53.51}, {X: 14.537, Y: 53.513}, {X: 14.543, Y: 53.513},
		{X: 14.546, Y: 53.511}, {X: 14.55, Y: 53.512}, {X: 14.545, Y: 53.515}, {X: 14.541, Y: 53.521}, {X: 14.539, Y: 53.529},
		{X: 14.541, Y: 53.537}, {X: 14.546, Y: 53.544}, {X: 14.565, Y: 53.551}, {X: 14.571, Y: 53.552}, {X: 14.594, Y: 53.547},
		{X: 14.6, Y: 53.551}, {X: 14.612, Y: 53.551}, {X: 14.629, Y: 53.547}, {X: 14.639, Y: 53.548}, {X: 14.645, Y: 53.545},
		{X: 14.65, Y: 53.548}, {X: 14.664, Y: 53.551}, {X: 14.681, Y: 53.55}, {X: 14.694, Y: 53.54}, {X: 14.697, Y: 53.533},
		{X: 14.696, Y: 53.525}, {X: 14.709, Y: 53.519}, {X: 14.714, Y: 53.511}, {X: 14.716, Y: 53.503}, {X: 14.715, Y: 53.493},
		{X: 14.712, Y: 53.485}, {X: 14.713, Y: 53.481}, {X: 14.721, Y: 53.473}, {X: 14.724, Y: 53.466}, {X: 14.726, Y: 53.448},
		{X: 14.731, Y: 53.449}, {X: 14.736, Y: 53.455}, {X: 14.744, Y: 53.458}, {X: 14.761, Y: 53.456}, {X: 14.77, Y: 53.453},
		{X: 14.774, Y: 53.449}, {X: 14.776, Y: 53.444}, {X: 14.776, Y: 53.438}, {X: 14.772, Y: 53.431}, {X: 14.766, Y: 53.427},
		{X: 14.766, Y: 53.421}, {X: 14.789, Y: 53.415}, {X: 14.798, Y: 53.41}, {X: 14.802, Y: 53.406}, {X: 14.805, Y: 53.398},
		{X: 14.804, Y: 53.39}, {X: 14.794, Y: 53.38}, {X: 14.789, Y: 53.379}, {X: 14.788, Y: 53.372}, {X: 14.782, Y: 53.365},
		{X: 14.79, Y: 53.365}, {X: 14.796, Y: 53.362}, {X: 14.806, Y: 53.35}, {X: 14.816, Y: 53.342}, {X: 14.819, Y: 53.336},
		{X: 14.819, Y: 53.328}, {X: 14.816, Y: 53.32}, {X: 14.811, Y: 53.315}, {X: 14.805, Y: 53.312}, {X: 14.797, Y: 53.312},
		{X: 14.792, Y: 53.315}, {X: 14.789, Y: 53.319}, {X: 14.781, Y: 53.323}, {X: 14.77, Y: 53.316}, {X: 14.761, Y: 53.314},
		{X: 14.749, Y: 53.314}, {X: 14.721, Y: 53.324}, {X: 14.714, Y: 53.329}, {X: 14.708, Y: 53.33}, {X: 14.695, Y: 53.343},
		{X: 14.681, Y: 53.339}, {X: 14.675, Y: 53.339}, {X: 14.661, Y: 53.342}, {X: 14.645, Y: 53.35}, {X: 14.633, Y: 53.35},
		{X: 14.627, Y: 53.342}, {X: 14.623, Y: 53.34}, {X: 14.61, Y: 53.339}, {X: 14.607, Y: 53.336}, {X: 14.606, Y: 53.329},
		{X: 14.599, Y: 53.324}, {X: 14.596, Y: 53.315}, {X: 14.586, Y: 53.31}, {X: 14.57, Y: 53.308}, {X: 14.563, Y: 53.31},
		{X: 14.555, Y: 53.317}, {X: 14.553, Y: 53.316}, {X: 14.526, Y: 53.32}, {X: 14.518, Y: 53.327}, {X: 14.517, Y: 53.338},
		{X: 14.523, Y: 53.347}, {X: 14.537, Y: 53.358}, {X: 14.526, Y: 53.361}, {X: 14.52, Y: 53.367}, {X: 14.519, Y: 53.373},
		{X: 14.514, Y: 53.373}, {X: 14.493, Y: 53.379}, {X: 14.489, Y: 53.377}, {X: 14.483, Y: 53.377}, {X: 14.473, Y: 53.381},
		{X: 14.467, Y: 53.387}, {X: 14.462, Y: 53.388}, {X: 14.456, Y: 53.393}, {X: 14.454, Y: 53.401}, {X: 14.455, Y: 53.406},
		{X: 14.458, Y: 53.41}, {X: 14.455, Y: 53.415}, {X: 14.455, Y: 53.42}, {X: 14.459, Y: 53.428}, {X: 14.44, Y: 53.431},
		{X: 14.434, Y: 53.437}, {X: 14.432, Y: 53.443}, {X: 14.434, Y: 53.452}, {X: 14.443, Y: 53.461}, {X: 14.449, Y: 53.463},
		{X: 14.445, Y: 53.47}, {X: 14.436, Y: 53.473}, {X: 14.432, Y: 53.477}, {X: 14.43, Y: 53.487},
	}
)
