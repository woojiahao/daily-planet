package models

import (
	"database/sql"
	"fmt"
	"strings"
)

type Cache struct {
	ID              int
	ConfigurationID int
	FeedID          int
	ArticleKey      string
}

type CacheModel struct {
	DB *sql.DB
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
