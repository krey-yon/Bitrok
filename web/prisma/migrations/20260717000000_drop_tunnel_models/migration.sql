-- Drop tunnel tables: the Go relay server (SQLite) is now the single source of
-- truth for tunnel configs and request logs. The dashboard proxies to its
-- REST API instead of keeping its own Postgres copy.
--
-- Order matters: drop tunnel_log first (it has an FK into tunnel), then tunnel.
-- Indexes and constraints on these tables are dropped automatically with the
-- tables.

DROP TABLE IF EXISTS "tunnel_log";
DROP TABLE IF EXISTS "tunnel";
