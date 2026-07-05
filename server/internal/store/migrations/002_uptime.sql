-- 002_uptime.sql

CREATE TABLE IF NOT EXISTS uptime_checks (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  ts          DATETIME DEFAULT CURRENT_TIMESTAMP,
  status      INTEGER NOT NULL,  -- HTTP status code from /health
  latency_ms  INTEGER,           -- response time in milliseconds
  error       TEXT               -- error message if status != 200
);

CREATE INDEX IF NOT EXISTS idx_uptime_ts ON uptime_checks(ts);
