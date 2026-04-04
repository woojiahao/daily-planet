package models

import (
	"database/sql"
	"time"
)

type Configuration struct {
	ID           int
	SnowflakeID  string
	Type         string
	CronSchedule string
	ShowStats    bool
	Disabled     bool
	CreatedAt    time.Time
}

type ConfigurationModel struct {
	DB *sql.DB
}
