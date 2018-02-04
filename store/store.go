package store

import (
	"database/sql"
	"github.com/bilus/fencer/index"
)

func LoadFromSQL(db *sql.DB) (*index.Index, int, error) {
	broadcasts, numSkipped, err := LoadBroadcastsFromSQL(db)
	if err != nil {
		return nil, 0, err
	}
	index, err := index.New(broadcasts)
	return index, numSkipped, err
}
