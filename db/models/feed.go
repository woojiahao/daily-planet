package models

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/woojiahao/daily-planet/db/helpers"
	"github.com/woojiahao/daily-planet/db/scanner"
	"github.com/woojiahao/daily-planet/db/transaction"
	"github.com/woojiahao/daily-planet/ds"
)

type (
	FeedID  int
	FeedKey ds.Pair[ConfigurationID, string]
)

type Feed struct {
	ID              FeedID
	ConfigurationID ConfigurationID
	URL             string
	FeedType        string
	CronSchedule    sql.NullString
	Disabled        bool
	CreatedAt       time.Time
}

type FeedModel struct {
	DB transaction.Transaction
}

type FeedInterface interface {
	// retrieve
	All() ([]Feed, error)
	AllEnabledByConfigurationID(id ConfigurationID) ([]Feed, error)
	AllByConfigurationID(configurationID ConfigurationID) ([]Feed, error)
	OneByID(id FeedID) (Feed, error)
	OneByKey(key FeedKey) (Feed, error)

	// insert
	InsertOne(configurationID ConfigurationID, url, feedType string) (Feed, error)
	InsertManyWithSameConfigurationID(configurationID ConfigurationID, urls, feedTypes []string) ([]Feed, error)

	// update
	DeleteOneByID(id FeedID) error

	// delete
	UpdateOneByID(id FeedID, disabled bool) error
}

func parseFeedRow(rows scanner.RowScanner) (Feed, error) {
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

func NewFeedKey(configurationID ConfigurationID, url string) FeedKey {
	return FeedKey(*ds.NewPair(configurationID, url))
}

func (f Feed) Key() FeedKey {
	return NewFeedKey(f.ConfigurationID, f.URL)
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

func (m FeedModel) AllEnabledByConfigurationID(id ConfigurationID) ([]Feed, error) {
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
		disabled = 0
		AND configuration_id = ?;`
	stmt, err := m.DB.Prepare(query)
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	rows, err := stmt.Query(id)
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

func (m FeedModel) AllByConfigurationID(configurationID ConfigurationID) ([]Feed, error) {
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
		configuration_id = ?;`
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

func (m FeedModel) OneByID(id FeedID) (Feed, error) {
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

func (m FeedModel) OneByKey(key FeedKey) (Feed, error) {
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
		AND url = ?
	LIMIT 1;`
	stmt, err := m.DB.Prepare(query)
	if err != nil {
		return Feed{}, err
	}

	defer stmt.Close()
	row := stmt.QueryRow(key.First, key.Second)

	feed, err := parseFeedRow(row)
	if err != nil {
		return Feed{}, err
	}

	return feed, nil
}

func (m FeedModel) InsertOne(configurationID ConfigurationID, url, feedType string) (Feed, error) {
	// Always default to nil cron_schedule and subsequent fetches will derive from the parent configuration
	query := `
	INSERT INTO feed  (
		configuration_id,
		url,
		feed_type,
		cron_schedule,
		disabled
	) VALUES (
		?,
		?,
		?,
		NULL,
		0
	) RETURNING 
		id,
		configuration_id,
		url,
		feed_type,
		cron_schedule,
		disabled,
		created_at;`
	stmt, err := m.DB.Prepare(query)
	if err != nil {
		return Feed{}, err
	}

	defer stmt.Close()

	row := stmt.QueryRow(configurationID, url, feedType)
	return parseFeedRow(row)
}

func (m FeedModel) InsertManyWithSameConfigurationID(
	configurationID ConfigurationID,
	urls []string,
	feedTypes []string,
) ([]Feed, error) {
	if len(urls) != len(feedTypes) {
		return nil, fmt.Errorf("urls and feedTypes must have same length")
	}

	if len(urls) == 0 {
		return []Feed{}, nil
	}

	valueStrings := make([]string, 0, len(urls))
	valueArgs := make([]any, 0, len(urls)*3)

	for i := range urls {
		valueStrings = append(valueStrings, "(?, ?, ?, NULL, 0)")
		valueArgs = append(valueArgs,
			configurationID,
			urls[i],
			feedTypes[i],
		)
	}

	query := fmt.Sprintf(`
		INSERT INTO feed (
			configuration_id,
			url,
			feed_type,
			cron_schedule,
			disabled
		) VALUES %s
		ON CONFLICT DO NOTHING
		RETURNING 
			id,
			configuration_id,
			url,
			feed_type,
			cron_schedule,
			disabled,
			created_at;
	`, strings.Join(valueStrings, ","))
	fmt.Printf("query is\n%s\n", query)
	fmt.Printf("values is %v\n", valueArgs)

	rows, err := m.DB.Query(query, valueArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var feeds []Feed
	for rows.Next() {
		f, err := parseFeedRow(rows)
		if err != nil {
			return nil, err
		}
		feeds = append(feeds, f)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return feeds, nil
}

func (m FeedModel) DeleteOneByID(id FeedID) error {
	query := `
	DELETE FROM feed
	WHERE id = ?;`
	stmt, err := m.DB.Prepare(query)
	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(id)
	return err
}

func (m FeedModel) UpdateOneByID(id FeedID, disabled bool) error {
	query := `
	UPDATE feed
	SET
		disabled = ?
	WHERE id = ?;`
	stmt, err := m.DB.Prepare(query)
	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(helpers.BoolToInt(disabled), id)
	return err
}
