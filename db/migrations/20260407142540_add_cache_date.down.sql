-- Need to do this elaborate dance because SQLite3 doesn't support dropping columns
BEGIN TRANSACTION;

CREATE TABLE cache_new (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  configuration_id INTEGER NOT NULL,
  feed_id INTEGER NOT NULL,
  article_key TEXT NOT NULL,
  FOREIGN KEY (configuration_id) REFERENCES configuration(id) ON DELETE CASCADE,
  FOREIGN KEY (feed_id) REFERENCES feed(id) ON DELETE CASCADE
);

INSERT INTO cache_new (
  id,
  configuration_id,
  feed_id,
  article_key
)
SELECT
  id,
  configuration_id,
  feed_id,
  article_key
FROM cache;

DROP TABLE cache;

ALTER TABLE cache_new RENAME TO cache;

COMMIT;
