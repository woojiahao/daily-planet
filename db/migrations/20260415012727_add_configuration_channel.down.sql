-- Need to do this elaborate dance because SQLite3 doesn't support dropping columns
BEGIN TRANSACTION;

CREATE TABLE configuration_new (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  snowflake_id TEXT NOT NULL,
  type TEXT NOT NULL CHECK (type IN ('server', 'direct_message')),
  cron_schedule TEXT NOT NULL,
  show_stats INTEGER NOT NULL DEFAULT 0 CHECK (show_stats IN (0, 1)),
  disabled INTEGER NOT NULL DEFAULT 0 CHECK (disabled IN (0, 1)),
  channel_id TEXT NULL,
  created_at TEXT DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO configuration_new (
  id,
  snowflake_id,
  type,
  cron_schedule,
  show_stats,
  disabled,
  channel_id,
  created_at
)
SELECT
  id,
  snowflake_id,
  type,
  cron_schedule,
  show_stats,
  disabled,
  channel_id,
  created_at
FROM configuration;

DROP TABLE configuration;

ALTER TABLE configuration_new RENAME TO configuration;

COMMIT;
