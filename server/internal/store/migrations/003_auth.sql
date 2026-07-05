-- 003_auth.sql

ALTER TABLE tunnels ADD COLUMN user_id TEXT DEFAULT 'legacy';
CREATE INDEX idx_tunnels_user_id ON tunnels(user_id);
