CREATE TABLE cache (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  configuration_id INTEGER NOT NULL,
  feed_id INTEGER NOT NULL,
  article_key TEXT NOT NULL,
  FOREIGN KEY (configuration_id) REFERENCES configuration(id) ON DELETE CASCADE,
  FOREIGN KEY (feed_id) REFERENCES feed(id) ON DELETE CASCADE
);
