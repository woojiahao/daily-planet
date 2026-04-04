// Package db contains the database connection to SQLite3
package db

import "database/sql"

type Database struct {
	db *sql.DB
}

func NewDatabase() (*Database, error) {
	db, err := sql.Open("sqlite3", "sqlite3://daily_planet.db")
	if err != nil {
		return nil, err
	}

	return &Database{db: db}, nil
}
