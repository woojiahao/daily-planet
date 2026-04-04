CREATE TABLE configuration (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  snowflake_id TEXT NOT NULL,
  type TEXT NOT NULL CHECK (type IN ('server', 'direct_message')),
  cron_schedule TEXT NOT NULL,
  show_stats INTEGER NOT NULL DEFAULT 0 CHECK (show_stats IN (0, 1)),
  disabled INTEGER NOT NULL DEFAULT 0 CHECK (disabled IN (0, 1)),
  created_at TEXT DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE feed (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  configuration_id INTEGER NOT NULL,
  url TEXT NOT NULL,
  feed_type TEXT NOT NULL CHECK (feed_type IN ('rss', 'atom', 'json')),
  cron_schedule TEXT,
  disabled INTEGER NOT NULL DEFAULT 0 CHECK (disabled IN (0, 1)),
  created_at TEXT DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (configuration_id) REFERENCES configuration(id) ON DELETE CASCADE
);

CREATE INDEX idx_configuration_snowflake_id ON configuration(snowflake_id);
CREATE INDEX idx_feed_configuration_id ON feed(configuration_id);


