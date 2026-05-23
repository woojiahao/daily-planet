package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/woojiahao/daily-planet/apperrors"
	"github.com/woojiahao/daily-planet/common"
	"github.com/woojiahao/daily-planet/db/scanner"
	"github.com/woojiahao/daily-planet/db/transaction"
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
	DB transaction.Transaction
}

type CacheInsert struct {
	CacheKey   CacheKey
	ArticleKey string
}

type CacheInterface interface {
	// retrieve
	All(cacheKeys ...CacheKey) ([]Cache, error)

	// insert
	Insert(cacheInserts ...CacheInsert) error
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

func (m CacheModel) Insert(cacheInserts ...CacheInsert) error {
	if len(cacheInserts) == 0 {
		return nil
	}

	var placeholderValues []any
	var queryPlaceholders []string

	for _, cacheInsert := range cacheInserts {
		placeholderValues = append(
			placeholderValues,
			cacheInsert.CacheKey.First,
			cacheInsert.CacheKey.Second,
			cacheInsert.ArticleKey,
		)
		queryPlaceholders = append(queryPlaceholders, "(?, ?, ?)")
	}

	fmt.Println(placeholderValues...)

	query := fmt.Sprintf(`
	INSERT INTO cache (
		configuration_id,
		feed_id,
		article_key
	) VALUES 
		%s
	ON CONFLICT DO NOTHING;`, strings.Join(queryPlaceholders, ",\n"))
	_, err := m.DB.Exec(query, placeholderValues...)
	if err != nil {
		return common.WrapError(apperrors.ErrCacheDBError, err)
	}

	return nil
}

func (m CacheModel) All(keys ...CacheKey) ([]Cache, error) {
	var preparedQueryPlaceholders []string
	for range len(keys) {
		preparedQueryPlaceholders = append(preparedQueryPlaceholders, "?")
	}
	preparedQueryString := strings.Join(preparedQueryPlaceholders, ", ")

	values := make([]any, len(keys)*2)
	for i, key := range keys {
		values[i] = int(key.First)
		values[i+len(keys)] = int(key.Second)
	}

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
	rows, err := m.DB.Query(query, values...)
	if err != nil {
		return nil, common.WrapError(apperrors.ErrCacheDBError, err)
	}

	defer rows.Close()

	var caches []Cache
	for rows.Next() {
		cache, err := parseCacheRow(rows)
		if err != nil {
			return nil, common.WrapError(apperrors.ErrCacheDBError, err)
		}
		caches = append(caches, cache)
	}

	return caches, nil
}
