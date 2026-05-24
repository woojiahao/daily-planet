// Package db contains the database connection to SQLite3
package db

import (
	"database/sql"

	"github.com/woojiahao/daily-planet/common"
	"github.com/woojiahao/daily-planet/db/models"
	_ "modernc.org/sqlite"
)

type Database struct {
	DB *sql.DB

	Configuration models.ConfigurationInterface
	Feed          models.FeedInterface
	Cache         models.CacheInterface
}

func New() (*Database, error) {
	db, err := sql.Open("sqlite", common.DBRoot()+"daily_planet.db")
	if err != nil {
		return nil, err
	}

	// we use several FK constraints so we want to ensure that it's enabled here
	_, err = db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		panic(err)
	}

	database := Database{
		DB:            db,
		Configuration: models.ConfigurationModel{DB: db},
		Feed:          models.FeedModel{DB: db},
		Cache:         models.CacheModel{DB: db},
	}
	return &database, nil
}

func (d *Database) WithTransaction(fn func(tx Database) error) error {
	tx, err := d.DB.Begin()
	if err != nil {
		return err
	}

	txDB := Database{
		DB:            d.DB,
		Configuration: models.ConfigurationModel{DB: tx},
		Feed:          models.FeedModel{DB: tx},
		Cache:         models.CacheModel{DB: tx},
	}

	if err := fn(txDB); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}
