# TLS Termination & Reverse Proxy

Node-Pulse's Pulse server listens in plaintext HTTP on port 6532 (it embeds
the SPA + API on one origin). For any non-loopback deployment you **must**
put a TLS-terminating reverse proxy in front of it — exposing the plaintext
port directly leaks session cookies, JWTs in transit, and Beacon heartbeat
payloads.

This document closes the D-G1 gap from `docs/user-journey.md` §23.2: the
README/AGENTS.md previously claimed "TLS 1.2+ enforced" while Pulse itself
ships no TLS. The enforcement happens at the proxy layer.

Two reference configurations are provided under `deploy/reverse-proxy/`:

| File | Stack | Use when |
|------|-------|----------|
| `nginx.conf` | nginx + certbot | You want maximal control or already run nginx |
| `Caddyfile` | Caddy | You want automatic HTTPS with the least ceremony |

## Common requirements (either proxy)

Whichever proxy you choose, it must:

1. **Terminate TLS** with a valid cert (Let's Encrypt via certbot, or
   Caddy's automatic HTTPS, or your own CA).
2. **Forward to `127.0.0.1:6532`** (or the Pulse container's address).
3. **Preserve the original client IP** via `X-Forwarded-For` /
   `X-Forwarded-Proto` — Pulse's `c.ClientIP()` and audit log rely on it.
   > Note: configuring trusted proxies in Pulse itself is the O-G6 gap
   > (not yet exposed); for now ensure only your proxy can reach Pulse.
4. **Upgrade WebSocket** connections — Node-Pulse uses `/api/v1/realtime`
   for live alert/node events. Without `Upgrade`/`Connection` headers the
   WS handshake fails and the dashboard falls back to polling.
5. **Not buffer SSE/streaming** responses.

## Quick start: Caddy (recommended for new deployments)

Caddy obtains and renews the certificate automatically. With
`deploy/reverse-proxy/Caddyfile`:

```bash
# 1. Point your domain's DNS A record at this host.
# 2. Run Caddy alongside the Pulse stack.
docker compose -f deploy/docker/docker-compose.prod.yml up -d
caddy run --config deploy/reverse-proxy/Caddyfile
```

Edit the Caddyfile to replace `pulse.example.com` with your domain.

## Quick start: nginx + certbot

```bash
# 1. Install nginx + certbot, point DNS at the host.
# 2. Copy the example config and edit server_name.
sudo cp deploy/reverse-proxy/nginx.conf /etc/nginx/sites-available/pulse
sudo ln -s /etc/nginx/sites-available/pulse /etc/nginx/sites-enabled/
sudo nginx -t && sudo systemctl reload nginx

# 3. Obtain a cert (this also rewrites the config to use HTTPS).
sudo certbot --nginx -d pulse.example.com
```

The example `nginx.conf` already includes the WebSocket upgrade block and
the proxy headers Pulse needs.

## Verifying

```bash
# 1. HTTP should redirect to HTTPS.
curl -I http://pulse.example.com/api/v1/health    # → 301

# 2. HTTPS health check should return 200.
curl https://pulse.example.com/api/v1/health

# 3. WebSocket upgrade should succeed (101 Switching Protocols).
curl -i -N \
  -H "Connection: Upgrade" -H "Upgrade: websocket" \
  -H "Sec-WebSocket-Version: 13" \
  -H "Sec-WebSocket-Key: dGVzdA==" \
  https://pulse.example.com/api/v1/realtime

# 4. Build version (D-G5) reachable through the proxy.
curl https://pulse.example.com/api/v1/version
```

## Notes

- **Beacon → Pulse**: beacons must use `pulse_server: https://...` once a
  proxy is in place. The `https://` scheme is what enables JWT-over-TLS.
- **HTTP/2**: both nginx and Caddy enable HTTP/2 by default over TLS;
  no Pulse-side change is needed.
- **Self-signed certs in dev**: set `BEACON_INSECURE_SKIP_VERIFY=true` on
  the beacon side (dev only — never in production).
