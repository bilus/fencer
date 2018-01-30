package store

import (
	"database/sql"
	"errors"
	pq "github.com/mc2soft/pq-types"
	geom "github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/geojson"
)

var MissingDataError = errors.New("Missing broadcast data")

func NewBroadcastFromRow(rows *sql.Rows) (*Broadcast, error) {
	var id int64
	var broadcastType sql.NullString
	var baselineData sql.NullString
	var boundingBox pq.PostGISBox2D
	var geoJson sql.NullString
	var freq sql.NullInt64

	if err := rows.Scan(&id, &broadcastType, &baselineData, &boundingBox, &geoJson, &freq); err != nil {
		return nil, err
	}
	if geoJson.Valid && broadcastType.Valid && baselineData.Valid {
		var covArea geom.T
		if err := geojson.Unmarshal([]byte(geoJson.String), &covArea); err != nil {
			return nil, err
		}
		var freqVal *int64
		if freq.Valid {
			freqVal = &freq.Int64
		}
		if broadcast, err := NewBroadcast(id, broadcastType.String, baselineData.String, freqVal, boundingBox, covArea); err != nil {
			return nil, err
		} else {
			return broadcast, nil
		}
	} else {
		return nil, MissingDataError
	}
}
