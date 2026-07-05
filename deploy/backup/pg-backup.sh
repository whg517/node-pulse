#!/usr/bin/env bash
# Node-Pulse PostgreSQL backup script.
#
# Takes a pg_dump of the pulse database, compresses it, and retains the last
# N days. Designed to run from cron / systemd timer alongside the prod
# docker-compose stack (deploy/docker/docker-compose.prod.yml).
#
# Usage (cron — daily at 02:00):
#   0 2 * * *  /opt/node-pulse/deploy/backup/pg-backup.sh >> /var/log/node-pulse-backup.log 2>&1
#
# Or via the bundled systemd timer (deploy/backup/node-pulse-backup.{service,timer}).
#
# Required env (read from the prod .env if SCRIPT_DIR/../.env exists, else the
# environment):
#   POSTGRES_PASSWORD   - password for the pulse DB user
#   POSTGRES_DB         - DB name (default: nodepulse)
#   POSTGRES_USER       - DB user (default: nodepulse)
#   BACKUP_DIR          - where to write dumps (default: /var/backups/node-pulse)
#   BACKUP_RETENTION_DAYS - how many days to keep (default: 14)
#
# Closes the D-G2 gap from docs/user-journey.md §23.1.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

# Load .env if present (prod compose uses the same file).
if [[ -f "${REPO_ROOT}/.env" ]]; then
  # shellcheck disable=SC1090
  set -a; . "${REPO_ROOT}/.env"; set +a
fi

: "${POSTGRES_DB:=nodepulse}"
: "${POSTGRES_USER:=nodepulse}"
: "${POSTGRES_PASSWORD:?POSTGRES_PASSWORD must be set (in .env or environment)}"
: "${BACKUP_DIR:=/var/backups/node-pulse}"
: "${BACKUP_RETENTION_DAYS:=14}"
# Allow overriding the docker compose invocation; default targets the prod stack.
: "${DOCKER_COMPOSE:=docker compose -f ${REPO_ROOT}/deploy/docker/docker-compose.prod.yml}"

TIMESTAMP="$(date -u +"%Y%m%dT%H%M%SZ")"
DUMP_FILE="${BACKUP_DIR}/${POSTGRES_DB}-${TIMESTAMP}.sql.gz"

mkdir -p "${BACKUP_DIR}"

echo "[$(date -u +%FT%TZ)] starting backup → ${DUMP_FILE}"

# Dump via the running postgres container (single source of truth for the
# schema + data). We pipe through gzip to keep the file small.
${DOCKER_COMPOSE} exec -T postgres pg_dump \
  -U "${POSTGRES_USER}" \
  -d "${POSTGRES_DB}" \
  --no-owner --no-privileges \
  --clean --if-exists \
  | gzip -9 > "${DUMP_FILE}"

if [[ ! -s "${DUMP_FILE}" ]]; then
  echo "[$(date -u +%FT%TZ)] ERROR: dump file is empty" >&2
  exit 1
fi

SIZE="$(du -h "${DUMP_FILE}" | cut -f1)"
echo "[$(date -u +%FT%TZ)] backup complete (${SIZE})"

# Prune old backups.
find "${BACKUP_DIR}" -name "${POSTGRES_DB}-*.sql.gz" -mtime "+${BACKUP_RETENTION_DAYS}" -print -delete \
  | sed 's/^/[prune] deleted /'

echo "[$(date -u +%FT%TZ)] done"
