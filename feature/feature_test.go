package feature_test

import (
	"fmt"

	"github.com/bilus/fencer/primitives"
	"github.com/bilus/fencer/testutil"
)

// RestaurantID uniquely identifies an example spatial feature.
type RestaurantID string

func (id RestaurantID) String() string {
	return string(id)
}

// Restaurant implements Feature interface to represents a restaurant's
// area and its other properties.
type Restaurant struct {
	ID   RestaurantID
	Name string
	Area *primitives.Rect
}

// Contains a true if a point lies within a restaurant's area.
func (r *Restaurant) Contains(point primitives.Point) (bool, error) {
	return testutil.Contains(*r.Bounds(), point), nil
}

// Bounds returns the bounding rectangle for a restaurant.
func (r *Restaurant) Bounds() *primitives.Rect {
	return r.Area
}

func (r *Restaurant) Key() RestaurantID {
	return r.ID
}

func Example() {
	area, _ := primitives.NewRect(primitives.Point{0, 0}, 100, 100)
	restaurant := Restaurant{
		ID:   "milliways",
		Name: "Restaurant at the End of Universe",
		Area: area,
	}
	contains, _ := restaurant.Contains(primitives.Point{50, 50})
	fmt.Println("Contains 50,50:", contains)
	// Output: Contains 50,50: true
}
