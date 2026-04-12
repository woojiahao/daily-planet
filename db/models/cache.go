package models

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/woojiahao/daily-planet/db/scanner"
)

type Cache struct {
	ID              int
	ConfigurationID int
	FeedID          int
	ArticleKey      string
	CreatedAt       time.Time
}

type CacheModel struct {
	DB *sql.DB
}

func parseCacheRow(rows scanner.RowScanner) (Cache, error) {
	var cache Cache
	var createdAtString string
	err := rows.Scan(
		&cache.ID,
		&cache.ConfigurationID,
		&cache.FeedID,
		&cache.ArticleKey,
		&createdAtString,
	)
	if err != nil {
		return Cache{}, err
	}

	createdAt, err := time.Parse("2006-01-02 15:04:05", createdAtString)
	if err != nil {
		return Cache{}, err
	}
	cache.CreatedAt = createdAt

	return cache, nil
}

func (m CacheModel) InsertOne(configurationID, feedID int, articleKey string) error {
	query := `
	INSERT INTO cache (
		configuration_id,
		feed_id,
		article_key
	) VALUES (
		?,
		?,
		?
	);`
	stmt, err := m.DB.Prepare(query)
	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(configurationID, feedID, articleKey)
	if err != nil {
		return err
	}

	return nil
}

func (m CacheModel) InsertMany(configurationIDs, feedIDs []int, articleKeys []string) error {
	if len(configurationIDs) != len(feedIDs) || len(configurationIDs) != len(articleKeys) || len(feedIDs) != len(articleKeys) {
		return fmt.Errorf("input should be of equal length")
	}

	var placeholderValues []any
	var queryPlaceholders []string

	for i := range len(configurationIDs) {
		placeholderValues = append(placeholderValues, configurationIDs[i], feedIDs[i], articleKeys[i])
		queryPlaceholders = append(queryPlaceholders, "(?, ?, ?)")
	}

	query := fmt.Sprintf(`
	INSERT INTO cache (
		configuration_id,
		feed_id,
		article_key
	) VALUES 
		%s
	;`, strings.Join(queryPlaceholders, ",\n"))
	stmt, err := m.DB.Prepare(query)
	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(placeholderValues...)
	if err != nil {
		return err
	}

	return nil
}

func (m CacheModel) InsertManyWithSameConfigurationIDAndFeedID(configurationID, feedID int, articleKeys []string) error {
	var placeholderValues []any
	var queryPlaceholders []string

	for i := range len(articleKeys) {
		placeholderValues = append(placeholderValues, configurationID, feedID, articleKeys[i])
		queryPlaceholders = append(queryPlaceholders, "(?, ?, ?)")
	}

	query := fmt.Sprintf(`
	INSERT INTO cache (
		configuration_id,
		feed_id,
		article_key
	) VALUES 
		%s
	;`, strings.Join(queryPlaceholders, ",\n"))
	stmt, err := m.DB.Prepare(query)
	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(placeholderValues...)
	if err != nil {
		return err
	}

	return nil
}

func (m CacheModel) AllByConfigurationIDAndFeedID(configurationID, feedID int) ([]Cache, error) {
	query := `
	SELECT
		id,
		configuration_id,
		feed_id,
		article_key,
		created_at
	FROM
		cache
	WHERE
		configuration_id = ?
		AND feed_id = ?;`
	stmt, err := m.DB.Prepare(query)
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	rows, err := stmt.Query(configurationID, feedID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var caches []Cache
	for rows.Next() {
		cache, err := parseCacheRow(rows)
		if err != nil {
			return nil, err
		}
		caches = append(caches, cache)
	}

	return caches, nil
}
