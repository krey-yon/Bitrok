# Production Readiness Audit

**Assessment date:** 2026-07-20
**Scope:** Product behavior, Go relay and CLI, Next.js dashboard, authentication and authorization, storage, tunnel protocol, installers, release automation, container runtime, performance boundaries, operations, and recovery.

## Executive assessment

Bitrok's code and release artifacts are suitable for an initial production launch after the remediations in this audit. The core product is internally consistent: GitHub authentication leads to one immutable username namespace, CLI credentials are scoped to that identity, tunnel ownership is enforced at every control-plane boundary, and the relay has explicit resource limits.

No known critical or high-severity code vulnerability remains from this review. The remaining release blockers are operational controls that cannot be proven from the repository. Public production traffic is therefore a **conditional no-go** until every external gate below has evidence from the real environment.

**Decision:** Code/build **GO**. Public production traffic **CONDITIONAL NO-GO** pending operational gates.

## Subsystem grades

| Area | Grade | Assessment |
|---|---:|---|
| Product and feature integrity | B+ | The supported launch workflow is coherent; unfinished commands and inaccurate WebSocket/logging/uptime claims were removed. Account deletion/export and token management remain absent. |
| Architecture | B | Clear web, relay, CLI, protocol, and persistence boundaries. A single relay replica is mandatory because active sessions and SQLite are process-local. |
| Security | A- | Strict JWT claims and algorithms, tenant authorization, immutable namespaces, origin checks, bounded protocol input, header sanitization, secure installers, and hardened container runtime are implemented. Production secret, proxy, DNS, TLS, OAuth, and Redis configuration still require external verification. |
| Reliability and recovery | B | CLI reconnects, transactional SQLite migrations, WAL, timeouts, health checks, and documented recovery exist. External monitoring and successful restores of both databases are still release gates. |
| Code quality and testing | A- | Go race tests, vet, targeted security tests, TypeScript, ESLint, production builds, release builds, vulnerability scans, and CI pass. There is no full browser end-to-end suite or measured coverage threshold. |
| Performance and scalability | B | Request/frame sizes, tunnel concurrency, per-user quotas, timeouts, retention, and static page generation bound normal operation. The buffered 10 MiB protocol, SQLite write path, in-memory sessions, and lack of production load/RUM data define the current ceiling. |
| Operations | C+ | Container hardening, health checks, rollback guidance, and backup procedures are present. Real alerts, retained logs, error tracking, restore evidence, and rollback rehearsal must be supplied by the deployment environment. |

## Critical and high-risk findings remediated

- Restricted relay JWT validation to HS256 and required subject, expiry, issuer, and audience.
- Enforced user ownership for tunnel CRUD and required platform hosts to use the authenticated user's immutable username suffix.
- Made usernames immutable in application logic and PostgreSQL, including a constraint and trigger that protect against direct writes and generic user-update paths.
- Reduced launch authentication to GitHub OAuth and required a verified GitHub email.
- Required valid origins on authenticated state-changing dashboard routes.
- Replaced reusable CLI browser callbacks with one-time database state and validated loopback return addresses.
- Bound relay responses to their tunnel session, closing cross-tunnel response injection.
- Bounded HTTP bodies, proxied responses, and WebSocket messages; stripped hop-by-hop, connection-nominated, injected forwarding, and response length headers.
- Added concurrent request limits, per-user tunnel quotas, SQLite busy handling, transactional migrations, retention cleanup, and restrictive file permissions.
- Made Redis-backed CLI token issuance fail closed when configured Redis storage is unavailable.
- Added reconnect behavior, atomic CLI configuration writes, process ownership checks, archive checksum verification, and current GoReleaser manifests.
- Removed unimplemented CLI commands and unsupported product claims instead of exposing misleading launch features.
- Hardened the container to run non-root with a read-only root filesystem, dropped capabilities, and `no-new-privileges`.
- Removed unrelated analytics and remote font dependencies; added security headers, legal/security contacts, canonical metadata, Open Graph output, and framework header removal.

## Verification evidence

The following passed on the audited worktree:

- `go test ./...`, `go test -race ./...`, and `go vet ./...`.
- Windows AMD64 CLI cross-build and six-platform GoReleaser CLI snapshot with checksums.
- `govulncheck ./...` with no reachable Go vulnerabilities.
- `npm test` with 8 focused tests, `npm run lint`, and `npx tsc --noEmit`.
- `npm audit` including development dependencies with zero reported vulnerabilities.
- Next.js 16 production build with 25 routes generated successfully.
- Prisma schema validation and all migrations applied in order to disposable PostgreSQL 17.
- Database enforcement test: first username claim succeeded; a later rename failed; a reserved direct username write failed.
- Shell installer syntax and served PowerShell security-regex verification.
- Docker image build and runtime validation as UID 100, read-only root, all capabilities dropped, and `no-new-privileges`.
- Relay `/health`, SQLite mode `0600`, online SQLite backup, and backup integrity workflow.
- Isolated production web smoke test for home, login, register, robots, sitemap, Open Graph image, icon, both installers, and `security.txt`; all returned HTTP 200 with expected security headers and no `X-Powered-By`.
- `git diff --check`.

## Known limitations accepted for initial launch

- Exactly one relay replica can serve traffic. Horizontal scaling requires shared tunnel state, shared durable storage, and cross-replica request routing.
- SQLite and the process-local WebSocket hub impose a throughput and availability ceiling.
- Proxy traffic is buffered and base64 encoded rather than streamed, with a 10 MiB body boundary.
- Visitor WebSockets are not supported; only HTTP tunneling and the CLI control WebSocket are supported.
- Without Redis, signed CLI JWTs cannot be individually revoked before expiry.
- Users do not have a token listing or revocation interface.
- Without Redis, dashboard rate limiting is best-effort and local to each web instance.
- The current Next.js content security policy requires `'unsafe-inline'` for styles.
- There is no user-facing account deletion or data export workflow.
- No external monitoring, retained log service, error tracker, browser telemetry, or production load-test result is configured in this repository.

These constraints are acceptable for a controlled initial launch only when traffic expectations fit one relay and incident response is staffed. Do not market visitor WebSocket support, high availability, zero-downtime relay deploys, or horizontally scalable throughput.

## Required external release gates

Do not direct public traffic to the production environment until all items have timestamped evidence:

- Real production secrets are unique, sufficiently random, and stored in the deployment secret manager; no example value is deployed.
- GitHub OAuth uses the exact production callback and the full sign-in flow succeeds.
- `prisma migrate deploy` has completed against production PostgreSQL and migration status is clean.
- Web and relay use the same JWT secret and, when Redis is enabled, the same reachable Redis service.
- Wildcard DNS and TLS cover the production domain; the relay is isolated behind the trusted proxy before proxy headers are enabled.
- An external health check, CLI control-channel check, and real HTTP tunnel monitor are alerting from outside the host.
- Encrypted SQLite and PostgreSQL backups exist, and each has passed an isolated restore test.
- The production flow passes end to end: GitHub sign-in, username claim, token generation, CLI login, tunnel creation, HTTP request/response, retained activity, stop, and reconnect after relay restart.
- `security@bitrok.tech`, `privacy@bitrok.tech`, and `support@bitrok.tech` are deliverable and monitored mailboxes.
- The previous relay image, dashboard deployment, database backups, and CLI release remain available for rollback.

## Post-launch priorities

1. Add synthetic tunnel monitoring, centralized logs, error reporting, and browser Core Web Vitals telemetry before increasing traffic.
2. Add token inventory/revocation and account deletion/export workflows.
3. Run repeatable relay load tests to set concurrency, latency, memory, and SQLite saturation thresholds.
4. Add a browser end-to-end suite for authentication, username claim, CLI callback, and dashboard activity.
5. Design shared session routing and durable metadata storage before running more than one relay replica.
