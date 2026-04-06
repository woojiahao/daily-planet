// Package db contains the database connection to SQLite3
package db

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
	"github.com/woojiahao/daily-planet/db/models"
)

type Database struct {
	Configuration models.ConfigurationModel
	Feed          models.FeedModel
}

func New() (*Database, error) {
	db, err := sql.Open("sqlite3", "sqlite3://daily_planet.db")
	if err != nil {
		return nil, err
	}

	database := Database{
		Configuration: models.ConfigurationModel{DB: db},
		Feed:          models.FeedModel{DB: db},
	}
	return &database, nil
}
