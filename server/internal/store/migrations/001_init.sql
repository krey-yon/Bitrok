-- 001_init.sql

CREATE TABLE IF NOT EXISTS tunnels (
  id          TEXT PRIMARY KEY,
  name        TEXT NOT NULL UNIQUE,
  host        TEXT NOT NULL UNIQUE,
  port        INTEGER NOT NULL DEFAULT 0,
  created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS tunnel_logs (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  tunnel_id   TEXT NOT NULL,
  method      TEXT,
  path        TEXT,
  status      INTEGER,
  latency_ms  INTEGER,
  bytes_in    INTEGER,
  bytes_out   INTEGER,
  ts          DATETIME DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (tunnel_id) REFERENCES tunnels(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_logs_tunnel ON tunnel_logs(tunnel_id);
CREATE INDEX IF NOT EXISTS idx_logs_ts ON tunnel_logs(ts);
CREATE INDEX IF NOT EXISTS idx_tunnels_host ON tunnels(host);
