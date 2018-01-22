package main

import (
	"database/sql"
	"github.com/dhconnelly/rtreego"
	_ "github.com/lib/pq"
	"testing"
)

func BenchmarkQuery(b *testing.B) {
	db, err := sql.Open("postgres", "user=postgres password=mysecretpassword host=localhost port=5432 dbname=broadcasts sslmode=disable")
	if err != nil {
		b.Fatal(err)
	}
	rt, err := build(db)
	if err != nil {
		b.Fatal(err)
	}
	point := rtreego.Point{13.4, 52.52}
	b.Run("Find point", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			_, err := query(rt, point)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
