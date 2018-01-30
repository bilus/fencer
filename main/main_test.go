package main

import (
	"database/sql"
	"github.com/bilus/fencer/store"
	_ "github.com/lib/pq"
	"testing"
)

// func BenchmarkFindBroadcasts(b *testing.B) {
// 	db, err := sql.Open("postgres", "user=postgres password=mysecretpassword host=localhost port=5432 dbname=broadcasts sslmode=disable")
// 	if err != nil {
// 		b.Fatal(err)
// 	}
// 	broadcastsStore, err := store.Load(db)
// 	if err != nil {
// 		b.Fatal(err)
// 	}
// 	point := store.Point{13.4, 52.52}
// 	b.Run("Find point", func(b *testing.B) {
// 		for n := 0; n < b.N; n++ {
// 			_, err := broadcastsStore.FindBroadcasts(point)
// 			if err != nil {
// 				b.Fatal(err)
// 			}
// 		}
// 	})
// }

func BenchmarkFindClosestBroadcasts(b *testing.B) {
	db, err := sql.Open("postgres", "user=postgres password=mysecretpassword host=localhost port=5432 dbname=broadcasts sslmode=disable")
	if err != nil {
		b.Fatal(err)
	}
	broadcastsStore, err := store.LoadFromSQL(db)
	if err != nil {
		b.Fatal(err)
	}
	// point := store.Point{13.4, 52.52} // Berlin
	// freqs := []Freq{100100, 100200}
	point := store.Point{-74.0059413, 40.7127837} // New York
	freqs := []Freq{1520, 1310}
	b.Run("Find point", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			radius := 58000.0
			_, err := broadcastsStore.FindClosestBroadcasts(point, radius, MatchAnyFrequency{freqs})
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
