package models

import (
	"database/sql"
	"strings"
	"time"

	"github.com/woojiahao/daily-planet/db/helpers"
	"github.com/woojiahao/daily-planet/db/scanner"
)

type (
	CommandSource   string
	ConfigurationID int
)

const (
	CommandSourceDM     CommandSource = "direct_message"
	CommandSourceServer CommandSource = "server"
)

type Configuration struct {
	ID           ConfigurationID
	SnowflakeID  string
	Type         CommandSource
	CronSchedule string
	ShowStats    bool
	Disabled     bool
	ChannelID    sql.NullString
	CreatedAt    time.Time
}

type ConfigurationModel struct {
	DB *sql.DB
}

type ConfigurationInterface interface {
	// fetch
	All() ([]Configuration, error)
	OneByID(id ConfigurationID) (Configuration, error)
	OneBySnowflakeID(snowflakeID string) (Configuration, error)

	// insert
	InsertOne(snowflakeID string, channelID *string, commandSource CommandSource) error

	// update
	// TODO(woojiahao): make this a struct for the inputs
	UpdateOneByID(id ConfigurationID, cronSchedule, channelID *string, showStats, disabled *bool) (Configuration, error)
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
		&configuration.ChannelID,
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
		channel_id,
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

func (m ConfigurationModel) OneByID(id ConfigurationID) (Configuration, error) {
	query := `
	SELECT
		id,
		snowflake_id,
		type,
		cron_schedule,
		show_stats,
		disabled,
		channel_id,
		created_at
	FROM
		configuration
	WHERE
		id = ?
	LIMIT 1;`
	stmt, err := m.DB.Prepare(query)
	if err != nil {
		return Configuration{}, err
	}
	defer stmt.Close()

	row := stmt.QueryRow(id)

	configuration, err := parseConfigurationRow(row)
	if err != nil {
		return Configuration{}, err
	}

	return configuration, nil
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
		channel_id,
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

func (m ConfigurationModel) InsertOne(snowflakeID string, channelID *string, commandSource CommandSource) error {
	// Always default to a cron_schedule of once every 6 hour starting at 12am (with seconds)
	query := `
	INSERT INTO configuration  (
		snowflake_id,
		type,
		cron_schedule,
		show_stats,
		disabled,
		channel_id
	) VALUES (
		?,
		?,
		'* 0 0,6,12,18 * * *',
		0,
		0,
		?
	);`
	stmt, err := m.DB.Prepare(query)
	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(snowflakeID, commandSource, channelID)
	if err != nil {
		return err
	}

	return nil
}

func (m ConfigurationModel) UpdateOneByID(id ConfigurationID, cronSchedule, channelID *string, showStats, disabled *bool) (Configuration, error) {
	if cronSchedule == nil && channelID == nil && showStats == nil && disabled == nil {
		return Configuration{}, nil
	}

	query := "UPDATE configuration SET "
	args := []any{}
	setClauses := []string{}

	if cronSchedule != nil {
		setClauses = append(setClauses, "cron_schedule = ?")
		args = append(args, *cronSchedule)
	}

	if channelID != nil {
		setClauses = append(setClauses, "channel_id = ?")
		args = append(args, *channelID)
	}

	if showStats != nil {
		setClauses = append(setClauses, "show_stats = ?")
		showStatsInt := helpers.BoolToInt(*showStats)
		args = append(args, showStatsInt)
	}

	if disabled != nil {
		setClauses = append(setClauses, "disabled = ?")
		disabledInt := helpers.BoolToInt(*disabled)
		args = append(args, disabledInt)
	}

	query += strings.Join(setClauses, ", ") + " WHERE id = ? RETURNING id, snowflake_id, type, cron_schedule, show_stats, disabled, channel_id, created_at"
	args = append(args, id)

	row := m.DB.QueryRow(query, args...)
	return parseConfigurationRow(row)
}
