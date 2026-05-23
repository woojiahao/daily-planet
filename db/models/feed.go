package models

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/woojiahao/daily-planet/apperrors"
	"github.com/woojiahao/daily-planet/common"
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

type FeedInsert struct {
	ConfigurationID ConfigurationID
	URL             string
	FeedType        string
}

type FeedInterface interface {
	// retrieve
	All() ([]Feed, error)
	AllEnabledByConfigurationID(id ConfigurationID) ([]Feed, error)
	AllByConfigurationID(configurationID ConfigurationID) ([]Feed, error)
	OneByID(id FeedID) (Feed, error)
	OneByKey(key FeedKey) (Feed, error)

	// insert
	Insert(feedInserts ...FeedInsert) ([]Feed, error)

	// delete
	DeleteOneByID(id FeedID) error

	// update
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

func scanFeedRows(rows *sql.Rows) ([]Feed, error) {
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
		return nil, common.WrapError(apperrors.ErrFeedDBError, err)
	}

	feeds, err := scanFeedRows(rows)
	if err != nil {
		return nil, common.WrapError(apperrors.ErrFeedDBError, err)
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
	rows, err := m.DB.Query(query, id)
	if err != nil {
		return nil, common.WrapError(apperrors.ErrFeedDBError, err)
	}

	feeds, err := scanFeedRows(rows)
	if err != nil {
		return nil, common.WrapError(apperrors.ErrFeedDBError, err)
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
	rows, err := m.DB.Query(query, configurationID)
	if err != nil {
		return nil, common.WrapError(apperrors.ErrFeedDBError, err)
	}

	feeds, err := scanFeedRows(rows)
	if err != nil {
		return nil, common.WrapError(apperrors.ErrFeedDBError, err)
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

	row := m.DB.QueryRow(query, id)
	feed, err := parseFeedRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Feed{}, common.WrapError(apperrors.ErrFeedNotFound, err)
		}
		return Feed{}, common.WrapError(apperrors.ErrFeedDBError, err)
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
	row := m.DB.QueryRow(query, key.First, key.Second)

	feed, err := parseFeedRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Feed{}, common.WrapError(apperrors.ErrFeedNotFound, err)
		}
		return Feed{}, common.WrapError(apperrors.ErrFeedDBError, err)
	}

	return feed, nil
}

func (m FeedModel) Insert(feedInserts ...FeedInsert) ([]Feed, error) {
	if len(feedInserts) == 0 {
		return []Feed{}, nil
	}

	n := len(feedInserts)
	valueStrings := make([]string, 0, n)
	valueArgs := make([]any, 0, n*3)

	for i := range n {
		valueStrings = append(valueStrings, "(?, ?, ?, NULL, 0)")
		valueArgs = append(valueArgs,
			feedInserts[i].ConfigurationID,
			feedInserts[i].URL,
			feedInserts[i].FeedType,
		)
	}

	// we update the id to the same value just so that conflicts will still return the
	// corresponding feed row
	query := fmt.Sprintf(`
		INSERT INTO feed (
			configuration_id,
			url,
			feed_type,
			cron_schedule,
			disabled
		) VALUES %s
		ON CONFLICT (configuration_id, url)
		DO UPDATE SET 
			id = id
		RETURNING 
			id,
			configuration_id,
			url,
			feed_type,
			cron_schedule,
			disabled,
			created_at;
	`, strings.Join(valueStrings, ","))

	rows, err := m.DB.Query(query, valueArgs...)
	if err != nil {
		return nil, common.WrapError(apperrors.ErrFeedDBError, err)
	}
	feeds, err := scanFeedRows(rows)
	if err != nil {
		return nil, common.WrapError(apperrors.ErrFeedDBError, err)
	}

	return feeds, nil
}

func (m FeedModel) DeleteOneByID(id FeedID) error {
	query := `
	DELETE FROM feed
	WHERE id = ?;`

	_, err := m.DB.Exec(query, id)
	if err != nil {
		return common.WrapError(apperrors.ErrFeedDBError, err)
	}

	return nil
}

func (m FeedModel) UpdateOneByID(id FeedID, disabled bool) error {
	query := `
	UPDATE feed
	SET
		disabled = ?
	WHERE id = ?;`

	_, err := m.DB.Exec(query, helpers.BoolToInt(disabled), id)
	if err != nil {
		return common.WrapError(apperrors.ErrFeedUpdateFailed, err)
	}

	return nil
}
