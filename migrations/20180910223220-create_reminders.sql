
-- +migrate Up
CREATE TABLE IF NOT EXISTS reminders(
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  content TEXT,
  chat_id INTEGER,
  created DATETIME,
  due_dt DATETIME,
  due_day VARCHAR(255)
);

-- +migrate Down
DROP TABLE reminders;
