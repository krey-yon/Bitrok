-- Tunnel names identify an app within an account. The original schema made
-- name globally unique, allowing one account to block a common name for every
-- other account even though public hosts are username-scoped.

ALTER TABLE tunnel_logs RENAME TO tunnel_logs_legacy_004;
ALTER TABLE tunnels RENAME TO tunnels_legacy_004;

DROP INDEX IF EXISTS idx_logs_tunnel;
DROP INDEX IF EXISTS idx_logs_ts;
DROP INDEX IF EXISTS idx_tunnels_host;
DROP INDEX IF EXISTS idx_tunnels_user_id;

CREATE TABLE tunnels (
  id          TEXT PRIMARY KEY,
  name        TEXT NOT NULL,
  host        TEXT NOT NULL UNIQUE,
  port        INTEGER NOT NULL DEFAULT 0,
  created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
  user_id     TEXT NOT NULL DEFAULT 'legacy'
);

INSERT INTO tunnels (id, name, host, port, created_at, updated_at, user_id)
SELECT id, name, host, port, created_at, updated_at, user_id
FROM tunnels_legacy_004;

CREATE TABLE tunnel_logs (
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

INSERT INTO tunnel_logs (id, tunnel_id, method, path, status, latency_ms, bytes_in, bytes_out, ts)
SELECT id, tunnel_id, method, path, status, latency_ms, bytes_in, bytes_out, ts
FROM tunnel_logs_legacy_004;

DROP TABLE tunnel_logs_legacy_004;
DROP TABLE tunnels_legacy_004;

CREATE UNIQUE INDEX idx_tunnels_user_name_unique ON tunnels(user_id, name);
CREATE INDEX idx_tunnels_user_id ON tunnels(user_id);
CREATE INDEX idx_tunnels_host ON tunnels(host);
CREATE INDEX idx_logs_tunnel ON tunnel_logs(tunnel_id);
CREATE INDEX idx_logs_ts ON tunnel_logs(ts);
