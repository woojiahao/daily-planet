package models

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/woojiahao/daily-planet/apperrors"
	"github.com/woojiahao/daily-planet/common"
	"github.com/woojiahao/daily-planet/db/helpers"
	"github.com/woojiahao/daily-planet/db/scanner"
	"github.com/woojiahao/daily-planet/db/transaction"
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
	DB transaction.Transaction
}

type ConfigurationInsert struct {
	SnowflakeID   string
	ChannelID     *string
	CommandSource CommandSource
}

type ConfigurationUpdate struct {
	CronSchedule *string
	ChannelID    *string
	ShowStats    *bool
	Disabled     *bool
}

func (cu ConfigurationUpdate) hasUpdate() bool {
	return cu.CronSchedule != nil || cu.ChannelID != nil || cu.ShowStats != nil || cu.Disabled != nil
}

type ConfigurationInterface interface {
	// fetch
	All() ([]Configuration, error)
	OneByID(id ConfigurationID) (Configuration, error)
	OneBySnowflakeID(snowflakeID string) (Configuration, error)

	// insert
	InsertOne(configurationInsert ConfigurationInsert) error

	// update
	UpdateOneByID(id ConfigurationID, configurationUpdate ConfigurationUpdate) (Configuration, error)
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
			return nil, common.WrapError(apperrors.ErrConfigurationDBError, err)
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
	row := m.DB.QueryRow(query, id)

	configuration, err := parseConfigurationRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Configuration{}, common.WrapError(apperrors.ErrConfigurationNotFound, err)
		}
		return Configuration{}, common.WrapError(apperrors.ErrConfigurationDBError, err)
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
	row := m.DB.QueryRow(query, snowflakeID)

	configuration, err := parseConfigurationRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Configuration{}, common.WrapError(apperrors.ErrConfigurationNotFound, err)
		}
		return Configuration{}, common.WrapError(apperrors.ErrConfigurationDBError, err)
	}

	return configuration, nil
}

func (m ConfigurationModel) InsertOne(configurationInsert ConfigurationInsert) error {
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
	_, err := m.DB.Exec(
		query,
		configurationInsert.SnowflakeID,
		configurationInsert.CommandSource,
		configurationInsert.ChannelID,
	)
	if err != nil {
		return common.WrapError(apperrors.ErrConfigurationDBError, err)
	}

	return nil
}

func (m ConfigurationModel) UpdateOneByID(id ConfigurationID, configurationUpdate ConfigurationUpdate) (Configuration, error) {
	if !configurationUpdate.hasUpdate() {
		return m.OneByID(id)
	}

	query := "UPDATE configuration SET "
	args := []any{}
	setClauses := []string{}

	if configurationUpdate.CronSchedule != nil {
		setClauses = append(setClauses, "cron_schedule = ?")
		args = append(args, *configurationUpdate.CronSchedule)
	}

	if configurationUpdate.ChannelID != nil {
		setClauses = append(setClauses, "channel_id = ?")
		args = append(args, *configurationUpdate.ChannelID)
	}

	if configurationUpdate.ShowStats != nil {
		setClauses = append(setClauses, "show_stats = ?")
		showStatsInt := helpers.BoolToInt(*configurationUpdate.ShowStats)
		args = append(args, showStatsInt)
	}

	if configurationUpdate.Disabled != nil {
		setClauses = append(setClauses, "disabled = ?")
		disabledInt := helpers.BoolToInt(*configurationUpdate.Disabled)
		args = append(args, disabledInt)
	}

	query += strings.Join(setClauses, ", ") + " WHERE id = ? RETURNING id, snowflake_id, type, cron_schedule, show_stats, disabled, channel_id, created_at"
	args = append(args, id)

	row := m.DB.QueryRow(query, args...)
	configuration, err := parseConfigurationRow(row)
	if err != nil {
		return Configuration{}, common.WrapError(apperrors.ErrConfigurationUpdateFailed, err)
	}

	return configuration, nil
}
