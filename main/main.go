package main

import (
	"database/sql"
	"fmt"
	"github.com/dhconnelly/rtreego"
	_ "github.com/jackc/pgx/stdlib"
	_ "github.com/lib/pq"
	pq "github.com/mc2soft/pq-types"
	"github.com/paulsmith/gogeos/geos"
	geom "github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/geojson"
	"log"
)

func main() {
	db, err := sql.Open("postgres", "user=postgres password=mysecretpassword host=localhost port=5432 dbname=broadcasts sslmode=disable")
	if err != nil {
		log.Fatal(err)
		return
	}

	err = runExperiment(db)
	if err != nil {
		log.Fatal(err)
	}
}

type Broadcast struct {
	broadcastId  int64
	bounds       *rtreego.Rect
	coverageArea []*geos.Geometry
}

func NewBroadcast(id int64, bounds pq.PostGISBox2D, coverageArea []*geos.Geometry) (*Broadcast, error) {
	// log.Println(bounds, lengths(bounds))
	rtBounds, err := rtreego.NewRect(
		rtreego.Point{bounds.Min.Lon, bounds.Min.Lat},
		lengths(bounds),
	)
	if err != nil {
		return nil, err
	}
	return &Broadcast{
		id,
		rtBounds,
		coverageArea,
	}, nil
}

func (b *Broadcast) Bounds() *rtreego.Rect {
	return b.bounds
}

func lengths(bounds pq.PostGISBox2D) []float64 {
	return []float64{
		bounds.Max.Lon - bounds.Min.Lon,
		bounds.Max.Lat - bounds.Min.Lat,
	}
}

func (b *Broadcast) Contains(point *geos.Geometry) bool {
	for _, geometry := range b.coverageArea {
		inter, err := geometry.Prepare().Covers(point) // TODO: Prepare before contains
		if err != nil {
			log.Println(geometry)
			log.Println(point)
			panic(fmt.Sprintf("Ooops: %v", err))
		}
		if inter {
			return true
		}
	}
	return false
}

func build(db *sql.DB) (*rtreego.Rtree, error) {
	rt := rtreego.NewTree(2, 25, 50)

	rows, err := db.Query("SELECT id, ST_Extent(coverage_area::geometry)::box2d, ST_AsGeoJson(coverage_area::geometry) FROM broadcasts GROUP BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var id int64
	var boundingBox pq.PostGISBox2D
	var geoJson sql.NullString
	for rows.Next() {
		if err := rows.Scan(&id, &boundingBox, &geoJson); err != nil {
			return nil, err
		}
		if geoJson.Valid {
			var covArea geom.T
			if err = geojson.Unmarshal([]byte(geoJson.String), &covArea); err != nil {
				return nil, err
			}
			multiPoly := covArea.(*geom.MultiPolygon)
			geometries := make([]*geos.Geometry, multiPoly.NumPolygons())
			for i := 0; i < multiPoly.NumPolygons(); i++ {
				geometries[i] = polygonToGeometry(multiPoly.Polygon(i))
			}

			if broadcast, err := NewBroadcast(id, boundingBox, geometries); err != nil {
				log.Printf("Skipping %v: %v", id, err)
			} else {
				rt.Insert(broadcast)
			}
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return rt, nil
}

func FilterContaining(point *geos.Geometry) rtreego.Filter {
	return func(results []rtreego.Spatial, object rtreego.Spatial) (refuse, abort bool) {
		if object.(*Broadcast).Contains(point) {
			return false, false
		} else {
			return true, false
		}
	}
}

func query(rt *rtreego.Rtree, point rtreego.Point) ([]rtreego.Spatial, error) {
	geosPoint, err := geos.NewPoint(geos.NewCoord(point[0], point[1]))
	if err != nil {
		return nil, err
	}

	p, err := rtreego.NewRect(point, []float64{0.01, 0.01})
	if err != nil {
		return nil, err
	}
	// result := make([]int64, 0)
	broadcasts := rt.SearchIntersect(p, FilterContaining(geosPoint))
	return broadcasts, nil
	// for _, broadcast := range broadcasts {
	// 	if broadcast.(*Broadcast).Contains(&point) {
	// 		result = append(result, broadcast.(*Broadcast).broadcastId)
	// 	}
	// }
	// return result, nil
}

func runExperiment(db *sql.DB) error {
	rt, err := build(db)
	if err != nil {
		return err
	}
	point := rtreego.Point{13.4, 52.52}
	results, err := query(rt, point)
	if err != nil {
		return err
	}

	for _, result := range results {
		println(result.(*Broadcast).broadcastId)
	}
	log.Printf("%v result(s).", len(results))
	return nil
}

func polygonToGeometry(geofence *geom.Polygon) *geos.Geometry {
	// Convert the outer shell to geos format.
	shell := geofence.LinearRing(0).Coords()
	shellGeos := geomToGeosCoords(shell)

	// TODO: Holes!
	// Convert each hole to geos format.
	// numHoles := geofence.NumLinearRings() - 1
	// holes := make([][]geos.Coord, numHoles)
	// for i := 0; i < numHoles; i++ {
	// 	holes[i] = geomToGeosCoords(geofence.LinearRing(i).Coords())
	// }

	return geos.Must(geos.NewPolygon(shellGeos)) //, holes...))
}

func geomToGeosCoord(coord geom.Coord) geos.Coord {
	return geos.Coord{
		X: coord.X(),
		Y: coord.Y(),
	}
}

func geomToGeosCoords(coords []geom.Coord) []geos.Coord {
	out := make([]geos.Coord, len(coords))
	for i := 0; i < len(coords); i++ {
		out[i] = geomToGeosCoord(coords[i])
	}
	return out
}

// Use geohashes:
// package main

// import (
// 	"github.com/Willyham/hashfill"
// 	geom "github.com/twpayne/go-geom"
// 	"github.com/twpayne/go-geom/encoding/geojson"
// 	"io/ioutil"
// )

// func main() {
// 	var err error
// 	path := "first.json"
// 	var data []byte
// 	if data, err = ioutil.ReadFile(path); err != nil {
// 		println("Error reading:", err)
// 		return
// 	}
// 	// var poly geom.T
// 	var feature geojson.Feature
// 	if err = feature.UnmarshalJSON(data); err != nil {
// 		println("Error unmarshalling:", err)
// 		return
// 	}
// 	filler := hashfill.NewRecursiveFiller(
// 		hashfill.WithMaxPrecision(9),
// 		// hashfill.WithFixedPrecision(),
// 	)

// 	poly := feature.Geometry.(*geom.MultiPolygon)

// 	var hashes []string
// 	if hashes, err = filler.Fill(poly.Polygon(0), hashfill.FillIntersects); err != nil {
// 		println("Error hashing:", err)
// 		return
// 	}

// 	println(len(hashes))
// 	if len(hashes) > 0 {
// 		println(hashes[0])
// 	}
// }
