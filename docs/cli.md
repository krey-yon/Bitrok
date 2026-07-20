# Bitrok CLI Reference

The CLI creates deterministic HTTP tunnel registrations and keeps an authenticated WebSocket control channel open to the relay. Visitor WebSocket upgrades are not supported by the current buffered request/response protocol.

## Install and authenticate

```bash
curl -fsSL https://bitrok.tech/install | sh
bitrok login
```

`bitrok login` opens the dashboard, verifies the generated credential against the Go relay, and writes configuration to `~/.config/bitrok/config.json` with mode `0600`.

## Start a tunnel

```bash
bitrok myapp 3000
```

The default public host is `myapp-<username>.bitrok.tech`. The foreground command displays live HTTP requests. Use `--detach` to keep a tunnel running in the background:

```bash
bitrok myapp 3000 --detach
bitrok list
bitrok status myapp
bitrok stop myapp
```

Common start flags:

| Flag | Purpose |
|---|---|
| `-d`, `--detach` | Run in the background |
| `--open` | Open the public URL in a browser |
| `--qr` | Print a QR code |
| `--no-anim` | Disable animated terminal output |
| `--allow-ip <CIDR>` | Restrict traffic by client CIDR; repeatable |
| `-H`, `--host <host>` | Override the derived hostname |

`bitrok http 3000 myapp` is an interactive-compatible alias for `bitrok myapp 3000`.

## Multi-tunnel configuration

Create `bitrok.yml`:

```yaml
tunnels:
  api:
    port: 3000
    subdomain: myapp-api
  web:
    port: 5173
    subdomain: myapp-web
```

Then run:

```bash
bitrok up
bitrok up api
bitrok up --detach
```

Do not put a token in `bitrok.yml`; use `bitrok login` to write the protected CLI configuration.

## Process commands

| Command | Behavior |
|---|---|
| `bitrok list` | Show active local tunnel processes |
| `bitrok list --server` | Show the account's relay registrations |
| `bitrok status [name]` | Show local process uptime and traffic counters |
| `bitrok stop <name>` | Stop the recorded Bitrok process after verifying process ownership |
| `bitrok stop --all` | Stop all active local tunnel processes |
| `bitrok down` | Alias for `stop` |

Detached process metadata and logs live under `~/.config/bitrok/run/`. PID ownership is validated before a process is signaled.

## Registration commands

The direct start flow creates and removes ephemeral registrations automatically. These commands are available for manual registration management:

```bash
bitrok create --name myapp --host myapp-you.bitrok.tech --port 3000
bitrok inspect myapp
bitrok update myapp --port 4000
bitrok delete myapp
```

Tunnel names are unique within an account; public hosts are unique across the relay.

## Configuration

`~/.config/bitrok/config.json` contains:

```json
{
  "server_url": "https://api.bitrok.tech",
  "web_url": "https://bitrok.tech",
  "token": "<credential>",
  "username": "you",
  "default_domain": "bitrok.tech"
}
```

The file and its containing directory use owner-only permissions. Supported environment overrides include `BITROK_SERVER` and `BITROK_WEB`. Plain HTTP service URLs are rejected unless they point to a loopback address.

## Protocol limits

- HTTP request and response bodies are buffered and limited to 10 MiB.
- The relay permits up to 50 concurrent requests per tunnel.
- A public request waits up to 30 seconds for the local service response.
- Hop-by-hop headers and connection-nominated headers are not forwarded.
- After the first successful connection, the CLI reconnects indefinitely with capped exponential backoff.
