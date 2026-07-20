# Bitrok

Bitrok is a self-hosted HTTP tunnel relay. The Go server owns tunnel registrations and SQLite persistence; the Go CLI keeps an authenticated WebSocket control channel open to forward HTTP traffic to a local service; the Next.js dashboard handles account sessions and CLI credentials.

## Local development

Run the relay and dashboard in separate terminals:

```bash
cp .env.example .env
go run ./server/cmd/bitrok-server
cd web && npm ci && npm run dev
```

For the dashboard, use a PostgreSQL `DATABASE_URL`, a 32-byte-plus `BETTER_AUTH_SECRET`, GitHub OAuth credentials, and the same `BITROK_JWT_SECRET` as the relay. The local relay listens on `http://localhost:8080`; the dashboard uses `BITROK_SERVER_URL=http://localhost:8080` when `NEXT_PUBLIC_USE_LOCALHOST=true`.

Build and test everything with:

```bash
go test ./...
go test -race ./...
go vet ./...
cd web && npm run lint && npx tsc --noEmit && npm run build
```

## User flow

1. Open the dashboard and sign in with GitHub.
2. Claim the permanent username namespace in Settings.
3. Generate a CLI token and run `bitrok login`, or configure the token with `bitrok auth`.
4. Start a tunnel with `bitrok myapp 3000`.

The public host is deterministic: `myapp-<username>.<BITROK_DOMAIN>`. Usernames are immutable after claim because CLI credentials carry the namespace claim.

## Production architecture

The relay is intentionally a single replica: SQLite and the in-memory WebSocket hub are not shared across processes. Put a reverse proxy in front of it with a wildcard certificate and wildcard DNS for `*.bitrok.tech` (or your configured domain), and route both HTTP traffic and the CLI control WebSocket upgrades to the same replica. Deploy the dashboard separately to Vercel with PostgreSQL.

See [docs/server-architecture.md](docs/server-architecture.md) for protocol details, [docs/production-deployment.md](docs/production-deployment.md) for the release checklist and recovery procedure, and [docs/production-readiness-audit.md](docs/production-readiness-audit.md) for the current CTO-level release assessment.

## Security model

- Relay JWTs accept only HS256 and require `sub`, `exp`, issuer, and audience.
- Tunnel CRUD is scoped by authenticated user ID; platform hosts must end in the caller's immutable username.
- Proxy requests and responses are bounded at 10 MiB and hop-by-hop/header-injection vectors are stripped.
- SQLite is opened with WAL, foreign keys, busy timeout, and a `0600` database file.
- The CLI stores credentials and run-state files with `0600` permissions and validates process ownership before stopping a tunnel.
- Installers verify GoReleaser SHA-256 checksums before extraction.

## License

See the repository license file before distributing the software.
