# Upgrade & Rollback

Node-Pulse ships two Go services (Pulse + Beacon) plus the embedded React SPA.
This runbook covers the Docker-Compose production deployment (see
`deploy/docker/docker-compose.prod.yml`); the same principles apply to a
binary install with minor command substitutions.

It closes the D-G3 gap from `docs/user-journey.md` §23.2 — previously
upgrades were "blind operations" with no compatibility or rollback guidance.

## 1. Before you upgrade

1. **Check the version you're running.**
   ```bash
   curl https://pulse.example.com/api/v1/version    # D-G5 endpoint
   ```
   Note `version` + `commit` so you can roll back if needed.
2. **Take a fresh backup** (see `deploy/backup/pg-backup.sh` / §3 below).
   ```bash
   sudo /opt/node-pulse/deploy/backup/pg-backup.sh
   ```
3. **Review the changelog / release notes** for breaking schema changes.
   Pulse uses golang-migrate with forward-only versioned files
   (`pulse/internal/db/migrations/0001..000N`); any new `NNNN_*.up.sql`
   runs automatically on the new Pulse's first boot. Incompatible
   migrations are called out in release notes — read them.
4. **Schedule a maintenance window.** Pulse itself is hot-restartable
   (heartbeat writes are batched and the grace period is ~10s), but a
   few seconds of `/api/v1/health` degraded are visible to consumers.

## 2. Performing the upgrade (Docker Compose)

```bash
cd /opt/node-pulse

# 1. Pull the latest code (or tag).
git fetch --tags
git checkout v1.2.3            # pin to a release tag, not main

# 2. Rebuild + restart. Migrations run automatically on startup.
docker compose -f deploy/docker/docker-compose.prod.yml up -d --build

# 3. Verify.
curl https://pulse.example.com/api/v1/health      # → 200 healthy
curl https://pulse.example.com/api/v1/version     # → new version
docker compose -f deploy/docker/docker-compose.prod.yml logs --tail=200 pulse
```

Beacons are independent and can be upgraded node-by-node at any time;
they retry heartbeats with exponential backoff during the Pulse restart.

## 3. Rollback

Rollback risk concentrates in **migrations**: a forward migration that
added a column is safe to roll back from, but one that dropped data is not.

### 3.1 If the new code is bad but migrations are backward-compatible

```bash
cd /opt/node-pulse
git checkout <previous-version-tag>
docker compose -f deploy/docker/docker-compose.prod.yml up -d --build
curl https://pulse.example.com/api/v1/version    # confirm old version
```

The old Pulse will run against the new schema as long as the migration
was purely additive (new tables/columns, no renamed/dropped fields).

### 3.2 If a migration is incompatible with the old code

Run the explicit `migrate-down` against the DB container, then downgrade
the binary:

```bash
# From the repo (needs the migrate CLI; see pulse/Makefile migrate-down)
cd pulse
make migrate-down                  # rolls back the latest migration
# Repeat `make migrate-down` for each migration introduced by the failed release.
# Verify with:  make migrate-version

cd /opt/node-pulse
git checkout <previous-version-tag>
docker compose -f deploy/docker/docker-compose.prod.yml up -d --build
```

> **Warning:** `migrate-down` for a migration that dropped a column will
> restore the column but **not** the data that was in it. Read each
> `NNNN_*.down.sql` before running it; if in doubt, restore from backup
> (§4) into a fresh DB instead of down-migrating.

### 3.3 Last resort: restore from backup

If the upgrade corrupted data or a down-migration is unsafe, restore the
backup taken before the upgrade (see [operations.md](operations.md) §Backup & Restore).

## 4. Compatibility matrix

Pulse migrations are versioned `0001`–`000N`. Each release documents which
migrations it introduces:

| Pulse release | New migrations | Backward-compatible rollback? |
|---------------|----------------|-------------------------------|
| v3.0 → v3.1 | none | ✅ (code-only; just `git checkout`) |
| (future) | (documented per release) | (documented per release) |

When cutting a release, append a row here so operators know what to expect.

## 5. Beacon upgrade notes

- Beacons are stateless wrt Pulse version (only the heartbeat schema matters,
  which has been stable since v1.0).
- Rolling a beacon: stop → `git pull && make build && make install` → start.
  systemd users: `sudo systemctl stop beacon && make build && sudo make install-systemd && sudo systemctl start beacon`.
- Beacons buffer up to ~`resume.cache_size` heartbeats locally during a Pulse
  outage; nothing is lost across a normal upgrade window.
