package main

import (
	"database/sql"
	"github.com/bilus/fencer/store"
	_ "github.com/jackc/pgx/stdlib"
	_ "github.com/lib/pq"
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

func runExperiment(db *sql.DB) error {
	bs, err := store.Load(db)
	if err != nil {
		return err
	}
	point := store.Point{13.4, 52.52}
	// results, err := bs.FindBroadcasts(point)
	results, err := bs.FindClosestBroadcasts(point)
	if err != nil {
		return err
	}

	for _, result := range results {
		println(result.BroadcastId)
	}
	log.Printf("%v result(s).", len(results))
	return nil
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
