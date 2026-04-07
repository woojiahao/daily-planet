package models

import "database/sql"

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
