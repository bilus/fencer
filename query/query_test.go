package query_test

import (
	"fmt"
	"github.com/bilus/fencer/feature"
	"github.com/bilus/fencer/primitives"
	"github.com/bilus/fencer/query"
	"strconv"
	"strings"
	// "testing"
)

type CountryID int

func (id CountryID) String() string {
	return strconv.Itoa(int(id))
}

type Country struct {
	ID           CountryID
	Name         string
	Population   int
	Region       string
	BoundingRect *primitives.Rect
}

func (c *Country) Key() feature.Key {
	return feature.Key(c.ID)
}

func (c *Country) Contains(p primitives.Point) (bool, error) {
	return p.MinDist(c.BoundingRect) == 0, nil
}

func (c *Country) Bounds() *primitives.Rect {
	return c.BoundingRect
}

var countries = []*Country{
	{1, "Vatican City", 800, "Europe", makeRect(bounds[0])},
	{2, "Tokelau", 1300, "Polynesia", makeRect(bounds[0])},
	{3, "Niue", 1600, "Polynesia", makeRect(bounds[0])},
	{4, "Tuvalu", 11200, "Oceania", makeRect(bounds[0])},
	{5, "Nauru", 11300, "Oceania", makeRect(bounds[0])},
}

// This example uses an example spatial feature implementation.
// See https://github.com/bilus/fencer/blob/master/query/query_test.go for more details.
func ExampleBuild_preconditionsUsingPredicates() {
	query := query.Build().Precondition(
		query.Pred(func(feature feature.Feature) (bool, error) {
			return feature.(*Country).Population > 10000, nil
		}),
	).Query()
	for _, country := range countries {
		query.Scan(country)
	}
	fmt.Println("Countries with population > 10000:", len(query.Distinct()))
	// Output: Countries with population > 10000: 2
}

type PopulationGreaterThan struct {
	threshold int
}

func (p PopulationGreaterThan) IsMatch(feature feature.Feature) (bool, error) {
	return feature.(*Country).Population > p.threshold, nil
}

// This example uses an example spatial feature implementation.
// See https://github.com/bilus/fencer/blob/master/query/query_test.go for more details.
func ExampleBuild_preconditionsUsingStructs() {
	// This is how you implement struct conditions:
	//
	//   type PopulationGreaterThan struct {
	//   	  threshold int
	//   }

	//   func (p PopulationGreaterThan) IsMatch(feature feature.Feature) (bool, error) {
	//   	  return feature.(*Country).Population > p.threshold, nil
	//   }

	query := query.Build().Precondition(PopulationGreaterThan{10000}).Query()
	for _, country := range countries {
		query.Scan(country)
	}
	fmt.Println("Countries with population > 10000:", len(query.Distinct()))
	// Output: Countries with population > 10000: 2
}

// This example uses an example spatial feature implementation.
// See https://github.com/bilus/fencer/blob/master/query/query_test.go for more details.
func ExampleBuild_conjunction() {
	// Both preconditions must match.
	q := query.Build()
	q.Precondition(
		query.Pred(func(feature feature.Feature) (bool, error) {
			return feature.(*Country).Population > 10000, nil
		}))
	q.Precondition(
		query.Pred(func(feature feature.Feature) (bool, error) {
			return strings.HasPrefix(feature.(*Country).Name, "T"), nil
		}),
	)
	query := q.Query()
	for _, country := range countries {
		query.Scan(country)
	}
	fmt.Println("Countries with population > 10000 with name beginning with T:", len(query.Distinct()))
	// Output: Countries with population > 10000 with name beginning with T: 1
}

type MostPopulatedByRegion struct{}

func (MostPopulatedByRegion) Map(match *query.Match) (*query.Match, error) {
	match.Replace(match.Feature.(*Country).Region)
	return match, nil
}

func (MostPopulatedByRegion) Reduce(result *query.Result, match *query.Match) error {
	for _, key := range match.ResultKeys {
		err := result.Update(key, func(entry *query.ResultEntry) error {
			if len(entry.Features) == 0 {
				entry.Features = []feature.Feature{match.Feature}
				return nil
			}

			// In production you'd probably want to use something more robust :>
			existingCountry := entry.Features[0].(*Country)
			currentCountry := match.Feature.(*Country)
			if existingCountry.Population < currentCountry.Population {
				entry.Features = []feature.Feature{match.Feature}
			}
			return nil
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// This example uses an example spatial feature implementation.
// See https://github.com/bilus/fencer/blob/master/query/query_test.go for more details.
func ExampleBuild_groupingResults() {
	query := query.Build().Aggregate(MostPopulatedByRegion{}).Query() // Map(MostPopulatedByRegion{}).Query()
	for _, country := range countries {
		query.Scan(country)
	}
	fmt.Println("Most populated countries per region:", len(query.Distinct()))
	// Output: Most populated countries per region: 3
}

// Really, really rough boundaries generated by  following steps in: https://github.com/JamesChevalier/cities, importing into geojson.io and drawing bounding boxes around the polygons.

var bounds = [][]primitives.Point{
	// Vatican City
	{
		{12.44450569152832, 41.89978557507729},
		{12.459547519683836, 41.89978557507729},
		{12.459547519683836, 41.907946360630994},
		{12.44450569152832, 41.907946360630994},
	},
	// Tokelau
	{
		{-172.7874755859375, -9.66573839518868},
		{-170.947265625, -9.66573839518868},
		{-170.947265625, -8.303905908124174},
		{-172.7874755859375, -8.303905908124174},
	},
	// Niue
	{
		{-170.13702392578125, -19.265776189877485},
		{-169.5849609375, -19.265776189877485},
		{-169.5849609375, -18.818567424622376},
		{-170.13702392578125, -18.818567424622376},
	},
	// Tuvalu
	{
		{174.74853515625, -11.059820828563412},
		{180.296630859375, -11.059820828563412},
		{180.296630859375, -5.397273407690904},
		{174.74853515625, -5.397273407690904},
	},
	// Nauru
	{
		{166.79443359375, -0.6227752122036241},
		{167.07595825195312, -0.6227752122036241},
		{167.07595825195312, -0.4051174740026618},
		{166.79443359375, -0.4051174740026618},
	},
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
