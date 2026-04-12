package models

import (
	"database/sql"
	"strings"
	"time"

	"github.com/woojiahao/daily-planet/db/scanner"
)

type (
	CommandSource   string
	ConfigurationID int
)

const (
	CommandSourceDM     CommandSource = "DM"
	CommandSourceServer CommandSource = "server"
)

type Configuration struct {
	ID           ConfigurationID
	SnowflakeID  string
	Type         CommandSource
	CronSchedule string
	ShowStats    bool
	Disabled     bool
	CreatedAt    time.Time
}

type ConfigurationModel struct {
	DB *sql.DB
}

func parseConfigurationRow(rows scanner.RowScanner) (Configuration, error) {
	var configuration Configuration
	var showStatsInt int
	var disabledInt int
	var createdAtString string

	err := rows.Scan(
		&configuration.ID,
		&configuration.SnowflakeID,
		&configuration.Type,
		&configuration.CronSchedule,
		&showStatsInt,
		&disabledInt,
		&createdAtString,
	)
	if err != nil {
		return Configuration{}, err
	}

	configuration.Disabled = disabledInt == 1
	configuration.ShowStats = showStatsInt == 1
	createdAt, err := time.Parse("2006-01-02 15:04:05", createdAtString)
	if err != nil {
		return Configuration{}, err
	}
	configuration.CreatedAt = createdAt

	return configuration, nil
}

func (m ConfigurationModel) All() ([]Configuration, error) {
	query := `
	SELECT
		id,
		snowflake_id,
		type,
		cron_schedule,
		show_stats,
		disabled,
		created_at
	FROM
		configuration;`
	rows, err := m.DB.Query(query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var configurations []Configuration
	for rows.Next() {
		configuration, err := parseConfigurationRow(rows)
		if err != nil {
			return nil, err
		}
		configurations = append(configurations, configuration)
	}

	return configurations, nil
}

func (m ConfigurationModel) OneBySnowflakeID(snowflakeID string) (Configuration, error) {
	query := `
	SELECT
		id,
		snowflake_id,
		type,
		cron_schedule,
		show_stats,
		disabled,
		created_at
	FROM
		configuration
	WHERE
		snowflake_id = ?;`
	stmt, err := m.DB.Prepare(query)
	if err != nil {
		return Configuration{}, err
	}

	defer stmt.Close()

	row := stmt.QueryRow(snowflakeID)

	configuration, err := parseConfigurationRow(row)
	if err != nil {
		return Configuration{}, err
	}

	return configuration, nil
}

func (m ConfigurationModel) InsertOne(snowflakeID string, commandSource CommandSource) error {
	// Always default to a cron_schedule of once every 6 hours
	query := `
	INSERT INTO configuration  (
		snowflake_id,
		type,
		cron_schedule,
		show_stats,
		disabled
	) VALUES (
		?,
		?,
		'0 */6 * * *',
		0,
		0
	);`
	stmt, err := m.DB.Prepare(query)
	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(snowflakeID, commandSource)
	if err != nil {
		return err
	}

	return nil
}

func (m ConfigurationModel) UpdateOneByID(id ConfigurationID, cronSchedule *string, showStats, disabled *bool) error {
	if cronSchedule == nil && showStats == nil && disabled == nil {
		return nil
	}

	query := "UPDATE configuration SET "
	args := []any{}
	setClauses := []string{}

	if cronSchedule != nil {
		setClauses = append(setClauses, "cron_schedule = ?")
		args = append(args, *cronSchedule)
	}

	if showStats != nil {
		setClauses = append(setClauses, "show_stats = ?")
		args = append(args, *showStats)
	}

	if disabled != nil {
		setClauses = append(setClauses, "disabled = ?")
		args = append(args, *disabled)
	}

	query += strings.Join(setClauses, ", ") + " WHERE id = ?"
	args = append(args, id)

	_, err := m.DB.Exec(query, args...)
	return err
}
