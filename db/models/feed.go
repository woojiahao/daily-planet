package models

import (
	"database/sql"
	"time"
)

type Feed struct {
	ID              int
	ConfigurationID int
	URL             string
	FeedType        string
	CronSchedule    string
	Disabled        bool
	CreatedAt       time.Time
}

type FeedModel struct {
	DB *sql.DB
}
