# Operations Runbook

Centralized SRE runbook for Node-Pulse. Closes the O-G7 gap from
`docs/user-journey.md` §23.2 — operational knowledge was previously
scattered across 8+ docs/files.

> See also: [deployment-tls.md](deployment-tls.md) (TLS proxy),
> [upgrade.md](upgrade.md) (version migrations),
> [observability.md](../observability.md) (metrics/tracing reference),
> [authentication.md](../authentication.md) (RBAC, sessions).

## 1. Health & daily checks

### 1.1 Read the health endpoint

```bash
curl -s https://pulse.example.com/api/v1/health | jq
```

Three-state model (`health.go`):

| Status | HTTP | Meaning |
|--------|------|---------|
| `healthy` | 200 | All subsystems nominal |
| `degraded` | 200 | Non-critical subsystem degraded (e.g. webhook delivery rate, auth-cleanup LastError, scheduler stale) |
| `unhealthy` | 503 | Critical failure: DB down, alert engine channel full, scheduler metrics-cleanup stale >3x interval |

The `checks` map names which subsystem tripped the status. The `scheduler`
and `alert_system` blocks have per-task detail (`last_run`, `last_error`,
`run_count`).

### 1.2 Frontend dashboards

- `/integrations/health` — overall + per-subsystem (15s poll)
- `/performance` — P95/P99 trends + anomalies (60s poll)
- `/settings/system-config` — read-only config + revalidate (admin)

### 1.3 External scrape

Prometheus scrapes both:
- Pulse `/metrics` (port 6532)
- Beacon `/metrics` (port 2112)

> A reference `prometheus.yml` + dashboard JSON is a future gap (O-G8);
> `docs/observability.md` has example scrape configs in the meantime.

## 2. Common incidents

### 2.1 `/health` returns 503 unhealthy

| `checks` value | Likely cause | Action |
|----------------|--------------|--------|
| `database: error: ...` | DB unreachable | Check `docker compose logs postgres`; verify `POSTGRES_PASSWORD`; restore DB if corrupt |
| `alert_engine: full` | Alert channel saturated (worker pool backed up) | Inspect `pulse` logs for `alert engine channel full`; consider raising worker pool or checking a stuck webhook |
| `scheduler: unhealthy: metrics-cleanup is stale` | metrics-cleanup task hasn't run for >3x interval | Check scheduler goroutine; `docker compose restart pulse` if wedged |

### 2.2 Login failures / account lockouts

- 5 failed logins → 10-min lock (`auth_handler.go: MaxFailedLoginAttempts`).
- Admins can release immediately: UI → Users → "Unlock" button (D-G2,
  O-G2), or `POST /api/v1/admin/users/:id/unlock`.
- 5 logins/min/IP → 429 rate limit. Different symptom (network-level, not
  per-account); raise the limit or wait it out.

### 2.3 Beacon offline

1. Check the beacon host: `sudo systemctl status beacon` (systemd) or the
   process directly.
2. `journalctl -u beacon -f` — look for "Pulse unreachable", JWT refresh
   failures, or TLS errors.
3. Pulse marks a node offline after 5 min of no heartbeat
   (`NodeStatusSweeper`); a brief blip self-heals, sustained failure
   usually means network or credentials.
4. Beacons buffer up to ~`resume.cache_size` heartbeats locally and
   replay on reconnect — no data loss for normal outages.

### 2.4 Disk filling up

The most likely culprit used to be unbounded `auth_audit_logs` /
`refresh_tokens` / `sessions` / `api_keys` growth; v3.1 wired `auth-cleanup`
(O-G1) to prune them on `cleanup.interval_seconds` (default daily). If disk
still fills:

```sql
-- Spot-check table sizes
SELECT relname, pg_size_pretty(pg_total_relation_size(relid))
FROM pg_catalog.pg_statio_user_tables
ORDER BY pg_total_relation_size(relid) DESC LIMIT 10;
```

`metrics` (per-second samples) is the next largest — adjust
`cleanup.retention_days` in `pulse.yaml` if you need to keep less.

### 2.5 Webhook delivery degraded

`/integrations/health` shows `webhook_delivery.success_rate`. If degraded:

1. UI → Webhooks → check the per-webhook delivery logs (v2.1 G4).
2. Common causes: target endpoint returning 5xx (transient — auto-retried),
   invalid URL after upstream change, or SSRF/https validation rejecting a
   new endpoint.
3. `pulse` logs have the full delivery error per webhook ID.

## 3. Backup & restore

### 3.1 Backup (D-G2)

`deploy/backup/pg-backup.sh` takes a `pg_dump` of the running Postgres
container, gzip-compresses it, and retains `BACKUP_RETENTION_DAYS`
(default 14). Two ways to schedule:

**systemd timer (recommended):**
```bash
sudo install -m 0644 deploy/backup/node-pulse-backup.{service,timer} /etc/systemd/system/
sudo systemctl enable --now node-pulse-backup.timer
systemctl list-timers node-pulse-backup    # confirm next run
```

**cron:**
```cron
0 2 * * *  /opt/node-pulse/deploy/backup/pg-backup.sh >> /var/log/node-pulse-backup.log 2>&1
```

The script reads `POSTGRES_PASSWORD` etc. from the prod `.env` automatically.

### 3.2 Restore

> Test a restore into a non-production DB at least once before you need it
> for real. An untested backup is not a backup.

```bash
# 1. Stop Pulse so nothing writes while you restore.
docker compose -f deploy/docker/docker-compose.prod.yml stop pulse

# 2. Restore the dump into the running Postgres container.
DUMP=/var/backups/node-pulse/nodepulse-20260706T020000Z.sql.gz
gunzip -c "$DUMP" | docker compose -f deploy/docker/docker-compose.prod.yml \
  exec -T postgres psql -U nodepulse -d nodepulse

# 3. Restart Pulse, verify version + health.
docker compose -f deploy/docker/docker-compose.prod.yml up -d pulse
curl https://pulse.example.com/api/v1/version
curl https://pulse.example.com/api/v1/health
```

The dump was taken with `--clean --if-exists --no-owner`, so it drops and
recreates objects cleanly. For a **full** restore (including a dropped DB),
target a fresh Postgres volume instead of overwriting the live one.

### 3.3 Restore into a fresh DB (disaster recovery)

```bash
# Stop the whole stack, nuke the volume, restart postgres only.
docker compose -f deploy/docker/docker-compose.prod.yml down
docker volume rm node-pulse_postgres_data     # ⚠️ destroys current data

docker compose -f deploy/docker/docker-compose.prod.yml up -d postgres
sleep 5
gunzip -c "$DUMP" | docker compose -f deploy/docker/docker-compose.prod.yml \
  exec -T postgres psql -U nodepulse -d nodepulse

docker compose -f deploy/docker/docker-compose.prod.yml up -d
```

## 4. Configuration changes

- **Pulse**: edit `pulse.yaml` (or `PULSE_*` env), then **restart** —
  hot-reload is the O-G4 gap (not yet implemented, unlike Beacon's SIGHUP).
- **Beacon**: edit `beacon.yaml`, then `sudo systemctl reload beacon`
  (SIGHUP) — no restart needed.

## 5. Useful one-liners

```bash
# Tail Pulse logs (docker)
docker compose -f deploy/docker/docker-compose.prod.yml logs -f pulse

# Tail Beacon logs (systemd)
sudo journalctl -u beacon -f

# Force-flush a stuck node state
# (after 5 min of no heartbeat Pulse auto-marks offline; manual override is rare)

# Confirm the auth cleanup task ran recently
curl -s https://pulse.example.com/api/v1/health | jq '.scheduler.tasks'

# List active sessions count
docker compose -f deploy/docker/docker-compose.prod.yml exec postgres \
  psql -U nodepulse -d nodepulse -c 'SELECT count(*) FROM sessions WHERE expires_at > NOW();'
```
