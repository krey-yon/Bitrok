# Bitrok Server — Architecture & Request Flow

This document covers **only the server code** — the Go binary deployed via Coolify/Docker. It explains how the server receives requests and relays them to the CLI client and back.

---

## Overview

The Bitrok server is a single Go binary that does three things:

1. **Tunnel management** — a REST API (`/api/tunnels`) for creating/listing/updating/deleting tunnels (authenticated via JWT)
2. **WebSocket relay** — a persistent WS connection (`/tunnel/{id}/connect`) that the CLI dials to receive proxied requests
3. **HTTP proxy** — any non-API request is intercepted, matched to a tunnel by `Host` header, and forwarded over the tunnel's WS session to the CLI, which forwards it to `localhost`

---

## High-Level Architecture

```
                         ┌─────────────────────────────────────────┐
                         │            Coolify / Traefik             │
                         │  (terminates TLS, wildcard *.domain)      │
                         └───────────────────┬─────────────────────┘
                                             │ HTTP (port 8080)
                                             ▼
                         ┌─────────────────────────────────────────┐
                         │            bitrok-server                 │
                         │                                         │
   Public request ──────►│  ProxyMiddleware                        │
   GET /api/users        │  ├── /health, /api/*, /tunnel/* → API   │
   (tunnel subdomain)    │  └── everything else → ProxyHandler      │
                         │                                         │
                         │  AuthMiddleware (Bearer JWT)            │
                         │  ├── /api/tunnels      (CRUD)           │
                         │  └── /tunnel/{id}/connect (WebSocket)    │
                         │                                         │
                         │  Hub: tunnelID → *Session                │
                         │  └── session.Conn (gorilla/websocket)    │
                         │                                         │
                         │  Store: SQLite (tunnels, logs, uptime)  │
                         └───────────────────┬─────────────────────┘
                                             │ WebSocket
                                             │ (ProxyRequest / ProxyResponse frames)
                                             ▼
                         ┌─────────────────────────────────────────┐
                         │            CLI client (bitrok)            │
                         │                                         │
                         │  readLoop                               │
                         │  ├── TypePing → send Pong               │
                         │  └── TypeRequest → handleRequest        │
                         │      ├── decode base64 body             │
                         │      ├── forward to localhost:PORT      │
                         │      ├── read response                  │
                         │      └── send ProxyResponse over WS     │
                         └─────────────────────────────────────────┘
```

---

## The Request/Response Relay Flow

This is the core of the tunneling. When a visitor hits `https://myapp.bitrok.tech/`:

```
 Visitor          Coolify (TLS)       bitrok-server                  CLI client             localhost:3000
   │                  │                    │                            │                        │
   │  HTTPS GET /     │                    │                            │                        │
   │─────────────────►│                    │                            │                        │
   │                  │  HTTP GET /        │                            │                        │
   │                  │  Host: myapp...     │                            │                        │
   │                  │───────────────────►│                            │                        │
   │                  │                    │                            │                        │
   │                  │            ProxyMiddleware                     │                        │
   │                  │            path "/" not in /api,/tunnel,/health│                        │
   │                  │            → ProxyHandler.ServeHTTP()         │                        │
   │                  │                    │                            │                        │
   │                  │            Store.GetTunnelByHost("myapp...")   │                        │
   │                  │            → tun.ID = "abc123"                 │                        │
   │                  │                    │                            │                        │
   │                  │            Hub.Get("abc123")                   │                        │
   │                  │            → session (active WS conn)          │                        │
   │                  │                    │                            │                        │
   │                  │            Serialize request:                  │                        │
   │                  │            ProxyRequest{                       │                        │
   │                  │              reqID, method, path,              │                        │
   │                  │              headers, bodyB64                  │                        │
   │                  │            }                                  │                        │
   │                  │                    │                            │                        │
   │                  │            respChan := make(chan ProxyResponse)│                        │
   │                  │            waiters[reqID] = respChan          │                        │
   │                  │            (blocks here, waiting for response)│                        │
   │                  │                    │                            │                        │
   │                  │                    │  WS write: ProxyRequest    │                        │
   │                  │                    │───────────────────────────►│                        │
   │                  │                    │                            │                        │
   │                  │                    │                    readLoop receives                │
   │                  │                    │                    TypeRequest frame                │
   │                  │                    │                            │                        │
   │                  │                    │                    handleRequest():                │
   │                  │                    │                    ├── decode base64 body           │
   │                  │                    │                    ├── build http.Request           │
   │                  │                    │                    │   http://localhost:3000/        │
   │                  │                    │                    └── httpClient.Do(req)           │
   │                  │                    │                            │                        │
   │                  │                    │                            │  HTTP GET /            │
   │                  │                    │                            │───────────────────────►│
   │                  │                    │                            │                        │
   │                  │                    │                            │  200 OK + body         │
   │                  │                    │                            │◄───────────────────────│
   │                  │                    │                            │                        │
   │                  │                    │                    sendResponse():                 │
   │                  │                    │                    ├── base64 encode body           │
   │                  │                    │                    └── WS write: ProxyResponse      │
   │                  │                    │◄───────────────────────────│                        │
   │                  │                    │                            │                        │
   │                  │            HandleResponse(resp)                │                        │
   │                  │            ch, ok := waiters[resp.ReqID]        │                        │
   │                  │            ch <- resp  (unblocks ServeHTTP)    │                        │
   │                  │                    │                            │                        │
   │                  │            w.WriteHeader(resp.Status)          │                        │
   │                  │            w.Write(decoded body)               │                        │
   │                  │            Store.LogRequest(...)              │                        │
   │                  │                    │                            │                        │
   │                  │  200 OK + body     │                            │                        │
   │                  │◄───────────────────│                            │                        │
   │  HTTPS 200 OK    │                    │                            │                        │
   │◄─────────────────│                    │                            │                        │
   │                  │                    │                            │                        │
```

---

## Key Components

### 1. Router & Middleware (`server/internal/api/routes.go`)

Chi router with middleware stack applied in order:

```
 Request → Recoverer → RequestID → RequestLogger → RateLimiter → ProxyMiddleware → [AuthMiddleware] → Handler
```

**ProxyMiddleware** (the key piece) runs **before** route matching. It checks the path:

| Path prefix | Routed to |
|---|---|
| `/health` | Health check handler (public) |
| `/api/*` | Tunnel CRUD API (JWT-authenticated) |
| `/tunnel/*` | WebSocket endpoint (JWT-authenticated) |
| `/.well-known/*` | Pass-through (public) |
| **anything else** | **ProxyHandler** (tunnel traffic) |

This means any subdomain request like `myapp.bitrok.tech/users` hits the ProxyHandler, not a 404.

### 2. ProxyHandler (`server/internal/relay/proxy.go`)

The traffic relay engine. For each incoming public request:

1. **Lookup tunnel** by `Host` header (e.g. `myapp.bitrok.tech`) via `Store.GetTunnelByHost()`
2. **Find active session** — `Hub.Get(tunnelID)` returns the WebSocket connection the CLI is holding
3. **Serialize** the HTTP request into a `ProxyRequest` frame:
   - method, path, host, headers (hop-by-hop stripped), body (base64-encoded)
   - adds `X-Forwarded-For`, `X-Forwarded-Proto` (preserved from reverse proxy), `X-Forwarded-Host`
4. **Register a waiter** — creates a buffered channel `respChan` and stores it in `waiters[reqID]`
5. **Send** the frame over WebSocket: `session.Conn.WriteJSON(frame)`
6. **Block** on `select` waiting for either:
   - `<-respChan` — the CLI's response arrived (within 30s timeout)
   - `<-ctx.Done()` — gateway timeout (504)
7. **Write response** — decode base64 body, set headers (Set-Cookie multi-values handled), write status + body to the original `ResponseWriter`
8. **Log** — `Store.LogRequest()` records method, path, status, latency, bytes in/out

### 3. WebSocket Session (`server/internal/api/ws.go`)

The endpoint `/tunnel/{id}/connect` is where the CLI connects:

1. **AuthMiddleware** validates the `Authorization: Bearer <JWT>` header first (checks `aud: "bitrok-cli"`, `iss: "bitrok"`, valid signature + expiry)
2. **Upgrade** HTTP → WebSocket
3. **Hello handshake** — CLI sends a `Hello` frame with its token; server validates against static tokens if configured (defense-in-depth, legacy)
4. **Register session** — `Hub.Register(session)` makes this tunnel's WS connection discoverable by the ProxyHandler
5. **Read loop** (goroutine) — listens for incoming frames:
   - `TypePong` → keepalive, ignore
   - `TypeResponse` → call `relay.HandleResponse(resp)` which routes it to the waiting `respChan`
6. **Write loop** (main goroutine) — sends `Ping` frames every 30s to keep the connection alive; exits when the read loop ends (disconnect)

### 4. Hub (`server/internal/relay/hub.go`)

A thread-safe in-memory map of active tunnels:

```
Hub
├── sessions: map[tunnelID]*Session    (protected by sync.RWMutex)
├── Register(s *Session)               — adds a session
├── Unregister(s *Session)             — removes only if it's still the same session
│                                        (prevents a stale disconnect from removing
│                                         a newer reconnect)
└── Get(tunnelID string) *Session      — O(1) lookup used by ProxyHandler
```

**Note:** Sessions are in-memory only. If the server restarts, all active tunnels drop and CLIs must reconnect (they have exponential backoff built in).

### 5. Store (`server/internal/store/sqlite.go`)

SQLite-backed persistence for:

| Table | Purpose |
|---|---|
| `tunnels` | Tunnel config (name, host, port, user_id) — created via `/api/tunnels` |
| `tunnel_logs` | Per-request logs (method, path, status, latency, bytes) — shown in dashboard |
| `uptime_checks` | Background health pings every 30s — used for status page |

The DB file lives at `/data/bitrok.db` (a Docker volume, persists across container restarts).

---

## Authentication Flow

```
 CLI                           bitrok-server                  Web Dashboard
  │                                 │                              │
  │  bitrok login                   │                              │
  │  (opens browser)                │                              │
  │─────────────────────────────────────────────────────────────────►│
  │                                 │                       user signs in
  │                                 │                       (GitHub OAuth / email)
  │                                 │                              │
  │                                 │              clicks "Generate CLI Token"
  │                                 │◄──────────────────────────────│
  │                                 │  POST /api/cli-auth/generate  │
  │                                 │  signs JWT with BITROK_JWT_SECRET
  │                                 │  claims: {sub, email,         │
  │                                 │   aud:"bitrok-cli",           │
  │                                 │   iss:"bitrok", exp:30d}      │
  │                                 │──────────────────────────────►│
  │  user pastes token into CLI     │                              │
  │◄──────────────────────────────  │                              │
  │                                 │                              │
  │  bitrok create / up             │                              │
  │  Authorization: Bearer <JWT>    │                              │
  │────────────────────────────────►│                              │
  │                                 │  AuthMiddleware:              │
  │                                 │  validateJWT(token, secret,   │
  │                                 │    "bitrok-cli", "bitrok")    │
  │                                 │  ✓ signature, exp, aud, iss    │
  │                                 │                              │
  │  201 Created / WS upgraded      │                              │
  │◄────────────────────────────────│                              │
```

The same `BITROK_JWT_SECRET` must be set on **both** the web dashboard and the server, or tokens are rejected.

---

## Environment Variables

### Required

| Variable | Example | Purpose |
|---|---|---|
| `BITROK_JWT_SECRET` | `5ea6a404...` (64 hex chars) | Signs/validates CLI tokens. **Must match the web dashboard's value.** |
| `BITROK_DOMAIN` | `bitrok.tech` | Used for WebSocket origin validation (browser CSRF check) |

### Set by Dockerfile (don't override unless needed)

| Variable | Default | Purpose |
|---|---|---|
| `BITROK_DB_PATH` | `/data/bitrok.db` | SQLite file location (Docker volume) |

### Optional tuning

| Variable | Default | Purpose |
|---|---|---|
| `BITROK_PORT` | `8080` | Listen port |
| `BITROK_LOG_LEVEL` | `info` | `debug` shows tunnel connect/disconnect + forwarded requests |
| `BITROK_RATE_LIMIT_CAPACITY` | `100` | Max requests per IP per window |
| `BITROK_RATE_LIMIT_WINDOW` | `60` (sec) | Rate limit window |
| `BITROK_WS_MAX_MESSAGE_SIZE_MB` | `10` | Max WebSocket frame size |
| `BITROK_WS_HELLO_TIMEOUT_SEC` | `10` | Time to receive Hello frame after upgrade |
| `BITROK_WS_PING_INTERVAL_SEC` | `30` | Keepalive ping interval |
| `BITROK_WS_READ_TIMEOUT_SEC` | `60` | Read deadline (resets on each message) |
| `BITROK_MAX_REQUEST_BODY_BYTES` | `1048576` (1 MB) | Max body for `/api/tunnels` CRUD (proxy body limit is 10 MB, hardcoded) |
| `BITROK_ALLOW_INSECURE_WS` | `false` | Skip Origin check entirely (dev only) |

---

## WebSocket Message Protocol (`pkg/protocol/protocol.go`)

All frames are JSON with a `type` field:

```
Hello          → { type: "hello", token, tunnel_id }          CLI → Server (auth)
ProxyRequest   → { type: "req", req_id, method, path,        Server → CLI
                    host, headers, body_b64 }
ProxyResponse  → { type: "res", req_id, status, headers,     CLI → Server
                    body_b64 }
Ping           → { type: "ping" }                              Server → CLI (keepalive)
Pong           → { type: "pong" }                              CLI → Server (keepalive reply)
Error          → { type: "error", error: "msg" }              Server → CLI (on failure)
```

Body payloads are base64-encoded to safely transport binary over JSON text frames.

---

## Deployment Checklist

- [ ] Wildcard DNS: `*.bitrok.tech` → server IP
- [ ] Coolify resource created from `krey-yon/Bitrok`, branch `main`
- [ ] Domain set to `*.bitrok.tech` (wildcard cert issued)
- [ ] `BITROK_JWT_SECRET` set (matches web dashboard)
- [ ] `BITROK_DOMAIN` set to your domain
- [ ] Deploy succeeded, logs show `server listening`
- [ ] `curl https://api.bitrok.tech/health` returns `{"status":"ok"}`
- [ ] Web dashboard's `BITROK_JWT_SECRET` matches the server's
- [ ] CLI: `BITROK_SERVER=https://bitrok.tech bitrok login` works
- [ ] Tunnel test: `bitrok up` → visit subdomain → traffic reaches localhost
