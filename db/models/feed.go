package models

import (
	"database/sql"
	"time"

	"github.com/woojiahao/daily-planet/db"
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

func parseFeedRow(rows db.RowScanner) (Feed, error) {
	var feed Feed
	var disabledInt int
	var createdAtString string
	err := rows.Scan(
		&feed.ID,
		&feed.ConfigurationID,
		&feed.URL,
		&feed.FeedType,
		&feed.CronSchedule,
		&disabledInt,
		&createdAtString,
	)
	if err != nil {
		return Feed{}, err
	}

	feed.Disabled = disabledInt == 1
	createdAt, err := time.Parse("2006-01-02 15:04:05", createdAtString)
	if err != nil {
		return Feed{}, err
	}
	feed.CreatedAt = createdAt

	return feed, nil
}

func (m FeedModel) All() ([]Feed, error) {
	query := `
	SELECT
		id,
		configuration_id,
		url,
		feed_type,
		cron_schedule,
		disabled,
		created_at
	FROM
		feed;`
	rows, err := m.DB.Query(query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var feeds []Feed
	for rows.Next() {
		feed, err := parseFeedRow(rows)
		if err != nil {
			return nil, err
		}
		feeds = append(feeds, feed)
	}

	return feeds, nil
}

func (m FeedModel) OneByID(id string) (Feed, error) {
	query := `
	SELECT
		id,
		configuration_id,
		url,
		feed_type,
		cron_schedule,
		disabled,
		created_at
	FROM
		feed
	WHERE
		id = ?
	LIMIT 1;`
	stmt, err := m.DB.Prepare(query)
	if err != nil {
		return Feed{}, err
	}

	defer stmt.Close()
	row := stmt.QueryRow(id)

	feed, err := parseFeedRow(row)
	if err != nil {
		return Feed{}, err
	}

	return feed, nil
}

func (m FeedModel) AllByConfigurationID(configurationID string) ([]Feed, error) {
	query := `
	SELECT
		id,
		configuration_id,
		url,
		feed_type,
		cron_schedule,
		disabled,
		created_at
	FROM
		feed
	WHERE
		configuration_id = ?
	LIMIT 1;`
	stmt, err := m.DB.Prepare(query)
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	rows, err := stmt.Query(configurationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var feeds []Feed
	for rows.Next() {
		feed, err := parseFeedRow(rows)
		if err != nil {
			return nil, err
		}
		feeds = append(feeds, feed)
	}

	return feeds, nil
}
