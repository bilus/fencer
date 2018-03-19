package query

import (
	"github.com/bilus/fencer/feature"
)

type ResultKey interface{}

type ResultEntry struct {
	Features []feature.Feature
	Meta     interface{}
}

type Result struct {
	entries map[ResultKey]*ResultEntry
}

type UpdateFunc func(entry *ResultEntry) error

func (result *Result) Update(key ResultKey, f UpdateFunc) error {
	entry := result.getSafe(key)
	return f(entry)
}

func (result *Result) Replace(match *Match) error {
	for _, key := range match.ResultKeys {
		entry := result.getSafe(key)
		entry.Features = []feature.Feature{match.Feature}
	}
	return nil
}

func (result *Result) getSafe(key ResultKey) *ResultEntry {
	m := result.entries
	entry, exists := m[key]
	if !exists {
		entry = &ResultEntry{}
		m[key] = entry
	}
	return entry
}

func (result *Result) distinct() []feature.Feature {
	features := make([]feature.Feature, 0)
	matched := make(map[feature.Key]struct{})
	for _, entry := range result.entries {
		for _, feature := range entry.Features {
			key := feature.Key()
			_, isMatched := matched[key]
			if !isMatched {
				features = append(features, feature)
				matched[key] = struct{}{}
			}
		}
	}
	return features
}
