// Package helpers contains helpers for the sources
package helpers

type EngineType string

const (
	EngineTypeRSS  EngineType = "rss"
	EngineTypeAtom EngineType = "atom"
	EngineTypeJSON EngineType = "json"
)
