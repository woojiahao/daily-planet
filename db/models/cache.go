package models

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/woojiahao/daily-planet/db/scanner"
	"github.com/woojiahao/daily-planet/ds"
)

type (
	CacheID  int
	CacheKey ds.Pair[ConfigurationID, FeedID]
)

type Cache struct {
	ID              CacheID
	ConfigurationID ConfigurationID
	FeedID          FeedID
	ArticleKey      string
	CreatedAt       time.Time
}

type CacheModel struct {
	DB *sql.DB
}

type CacheInterface interface {
	AllByKey(cacheKey CacheKey) ([]Cache, error)
	AllByKeys(cacheKeys []CacheKey) ([]Cache, error)

	InsertOne(cacheKey CacheKey, articleKey string) error
	InsertMany(cacheKeys []CacheKey, articleKeys []string) error
	InsertManyWithSameKey(cacheKey CacheKey, articleKeys []string) error
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

func NewCacheKey(configurationID ConfigurationID, feedID FeedID) CacheKey {
	return CacheKey(*ds.NewPair(configurationID, feedID))
}

func (c Cache) Key() CacheKey {
	return NewCacheKey(c.ConfigurationID, c.FeedID)
}

func (m CacheModel) InsertOne(cacheKey CacheKey, articleKey string) error {
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

	_, err = stmt.Exec(cacheKey.First, cacheKey.Second, articleKey)
	if err != nil {
		return err
	}

	return nil
}

func (m CacheModel) InsertMany(cacheKeys []CacheKey, articleKeys []string) error {
	if len(cacheKeys) != len(articleKeys) {
		return fmt.Errorf("input should be of equal length")
	}

	var placeholderValues []any
	var queryPlaceholders []string

	for i := range len(cacheKeys) {
		placeholderValues = append(placeholderValues, cacheKeys[i].First, cacheKeys[i].Second, articleKeys[i])
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

func (m CacheModel) InsertManyWithSameKey(cacheKey CacheKey, articleKeys []string) error {
	var placeholderValues []any
	var queryPlaceholders []string

	for i := range len(articleKeys) {
		placeholderValues = append(placeholderValues, cacheKey.First, cacheKey.Second, articleKeys[i])
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

func (m CacheModel) AllByKey(cacheKey CacheKey) ([]Cache, error) {
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

	rows, err := stmt.Query(cacheKey.First, cacheKey.Second)
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

func (m CacheModel) AllByKeys(keys []CacheKey) ([]Cache, error) {
	var preparedQueryPlaceholders []string
	for range len(keys) {
		preparedQueryPlaceholders = append(preparedQueryPlaceholders, "?")
	}
	preparedQueryString := strings.Join(preparedQueryPlaceholders, ", ")

	query := fmt.Sprintf(`
	SELECT
		id,
		configuration_id,
		feed_id,
		article_key,
		created_at
	FROM
		cache
	WHERE
		configuration_id IN (%s)
		AND feed_id IN (%s);`, preparedQueryString, preparedQueryString)
	stmt, err := m.DB.Prepare(query)
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	values := make([]any, len(keys)*2)
	for i, key := range keys {
		values[i] = int(key.First)
		values[i+len(keys)] = int(key.Second)
	}
	rows, err := stmt.Query(values...)
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
