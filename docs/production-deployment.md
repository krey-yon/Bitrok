# Production Deployment

This document is the release gate for Bitrok. It describes the deployment shape the code supports today; changing any of these assumptions requires an architecture change and a new audit.

## Required topology

- One relay replica only. SQLite and the in-memory WebSocket hub are process-local; two replicas will split tunnel registrations and can route a visitor to a replica that has no CLI session.
- A reverse proxy terminates TLS, forwards `X-Forwarded-Host`, `X-Forwarded-Proto`, and the client IP, and supports the CLI control WebSocket upgrade. Set `BITROK_TRUST_PROXY_HEADERS=true` only when the relay is reachable exclusively through that trusted proxy network.
- Wildcard DNS and a wildcard certificate cover `*.bitrok.tech` (or the value of `BITROK_DOMAIN`). The relay API should have a separate stable hostname such as `api.bitrok.tech`.
- The dashboard is deployed separately to Vercel with PostgreSQL. `BITROK_JWT_SECRET` must match exactly between Vercel and the relay.
- Configure the same `BITROK_REDIS_URL` on both projects if opaque CLI tokens or distributed web rate limiting are enabled. Token issuance fails closed when a configured Redis service is unavailable. Without Redis, the dashboard uses signed 30-day JWT CLI tokens that cannot be individually revoked, plus best-effort per-instance rate limiting.

## Secrets and environment

Set every value in `.env.example`. Generate secrets with `openssl rand -hex 32`; do not use the example values. Configure real `SECURITY_CONTACT_EMAIL`, `PRIVACY_CONTACT_EMAIL`, and `NEXT_PUBLIC_SUPPORT_EMAIL` mailboxes before publishing the legal/security pages.

For the relay, require `BITROK_JWT_SECRET`, set `BITROK_DOMAIN`, keep `BITROK_ALLOW_INSECURE_WS=false`, and leave `BITROK_AUTH_TOKENS` empty unless migrating legacy clients. Keep `BITROK_DB_PATH=/data/bitrok.db` on a persistent volume.

## Database backup and restore

Bitrok has two independent data stores and both must be recoverable before release:

- Relay SQLite stores tunnels and retained request metadata.
- Dashboard PostgreSQL stores users, sessions, OAuth accounts, usernames, and one-time CLI authentication callbacks.

Back up both stores before every schema deployment. A backup is not considered valid until it has been restored into an isolated environment and the relevant integrity or smoke checks pass.

### Relay SQLite

Back up the SQLite database while the server is running with SQLite's online backup API:

```bash
docker compose exec -T bitrok-server \
  sqlite3 /data/bitrok.db ".backup '/data/bitrok.backup.db'"
docker compose cp bitrok-server:/data/bitrok.backup.db ./bitrok-$(date -u +%Y%m%dT%H%M%SZ).db
```

Verify the artifact before relying on it:

```bash
sqlite3 ./bitrok-YYYYMMDDTHHMMSSZ.db "PRAGMA integrity_check;"
```

To restore, stop the relay, preserve the existing volume, replace `/data/bitrok.db` with a verified backup, and start exactly one replica. Run the smoke tests below before reconnecting DNS traffic.

### Dashboard PostgreSQL

Enable the managed PostgreSQL provider's automated backups and point-in-time recovery where available. Also take an application-controlled logical backup before applying Prisma migrations:

```bash
pg_dump --format=custom --no-owner --no-acl \
  --file=bitrok-web-YYYYMMDDTHHMMSSZ.dump "$DATABASE_URL"
pg_restore --list bitrok-web-YYYYMMDDTHHMMSSZ.dump >/dev/null
```

Test recovery using a new, empty, isolated database. Never use a production connection string for a restore test:

```bash
pg_restore --exit-on-error --no-owner --no-acl \
  --dbname="$RESTORE_TEST_DATABASE_URL" bitrok-web-YYYYMMDDTHHMMSSZ.dump
DATABASE_URL="$RESTORE_TEST_DATABASE_URL" npx prisma migrate status --schema web/prisma/schema.prisma
```

After restore, verify that a test user can sign in, that existing usernames remain immutable, and that expired CLI callbacks cannot be exchanged. Delete the isolated restore database after recording the result.

For a production schema release, take both backups, run `npm run db:deploy` from `web`, verify `npx prisma migrate status`, deploy the matching dashboard and relay versions, and complete the smoke test below. Prisma migrations in this repository are forward-only; an application rollback that is incompatible with the migrated schema requires restoring the matching PostgreSQL backup.

Keep at least seven days of encrypted, access-controlled SQLite and PostgreSQL backups. Alert on backup age and failure, and test recovery for both stores at least monthly.

## Monitoring and alerting

`GET /health` only confirms that the process can reach its local SQLite database. It is a readiness check, not an uptime report. Configure an external monitor from outside the deployment that checks `/health`, the CLI control WebSocket upgrade, and a real HTTP tunnel request. Alert on consecutive failures, 5xx rate, control-channel disconnects, disk usage, backup age, and certificate expiry. Forward structured JSON logs to a retained log system; the relay intentionally does not include a vendor-specific error tracker.

## Release smoke test

1. `curl -fsS https://api.bitrok.tech/health` returns `status=ok` and the expected version.
2. `curl -fsS https://bitrok.tech/install | sh -s -- --version <tag>` verifies the archive checksum and installs the expected CLI version.
3. Sign in, claim a username, generate a CLI token, and run `bitrok login` against `https://api.bitrok.tech`.
4. Start `bitrok smoke 3000`, then exercise GET, POST, a redirect, and a cookie. Confirm an oversized response is rejected at the documented 10 MiB boundary.
5. Confirm the dashboard lists the tunnel and retained request logs, then run `bitrok stop smoke` and verify the tunnel is inactive.
6. Confirm the external monitor sees the same tunnel during a controlled relay restart and that the CLI reconnects.

## Rollback

Keep the previous container image, dashboard deployment, and CLI release tag available. Roll back the relay and dashboard together when JWT claims, protocol frames, or schema migrations change. Never run an older application against an incompatible forward-migrated SQLite or PostgreSQL schema; restore the matching verified backup when the rollback requires it.
