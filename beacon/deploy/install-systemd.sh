#!/usr/bin/env bash
# Install the Node-Pulse beacon agent as a systemd service on the current host.
#
# Usage:
#   sudo ./beacon/deploy/install-systemd.sh
#
# What it does (idempotent):
#   1. Builds the static Linux AMD64 binary (skips if /usr/local/bin/beacon exists
#      and -b/--skip-build is passed).
#   2. Creates a dedicated `beacon` system user and the directories it owns
#      (/etc/beacon, /var/lib/beacon, /var/log/beacon).
#   3. Installs the binary to /usr/local/bin/beacon and the unit file to
#      /etc/systemd/system/beacon.service.
#   4. Seeds /etc/beacon/beacon.yaml from beacon.yaml.example if none exists
#      (the operator MUST edit it before starting — pulse_server/api_key required).
#   5. Reloads systemd, enables and (if config is non-empty) starts the service.
#
# This closes the D-G4 gap from docs/user-journey.md §23.2: `make install`
# previously only copied the binary, leaving every operator to hand-roll
# process supervision.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
BINARY_DEST="/usr/local/bin/beacon"
UNIT_SRC="${SCRIPT_DIR}/beacon.service"
UNIT_DEST="/etc/systemd/system/beacon.service"
CONFIG_DIR="/etc/beacon"
CONFIG_FILE="${CONFIG_DIR}/beacon.yaml"
STATE_DIR="/var/lib/beacon"
LOG_DIR="/var/log/beacon"

SKIP_BUILD=0
for arg in "$@"; do
  case "$arg" in
    -b|--skip-build) SKIP_BUILD=1 ;;
    -h|--help)
      cat <<EOF
Usage: sudo $0 [-b|--skip-build]
  -b, --skip-build  Do not rebuild; assume ${BINARY_DEST} is up to date.
EOF
      exit 0 ;;
    *) echo "Unknown argument: $arg" >&2; exit 2 ;;
  esac
done

if [[ $EUID -ne 0 ]]; then
  echo "ERROR: must run as root (use sudo)" >&2
  exit 1
fi

echo "==> [1/5] Build beacon (skip with -b)"
if [[ $SKIP_BUILD -eq 1 && -x "${BINARY_DEST}" ]]; then
  echo "    skipped (--skip-build)"
else
  (cd "${REPO_ROOT}/beacon" && make build)
  echo "    install binary to ${BINARY_DEST}"
  install -m 0755 "${REPO_ROOT}/beacon/build/beacon" "${BINARY_DEST}"
fi

echo "==> [2/5] Create beacon user and directories"
if ! id beacon &>/dev/null; then
  useradd --system --no-create-home --shell /usr/sbin/nologin beacon
fi
install -d -o beacon -g beacon -m 0750 "${CONFIG_DIR}" "${STATE_DIR}" "${LOG_DIR}"

echo "==> [3/5] Install systemd unit"
install -m 0644 "${UNIT_SRC}" "${UNIT_DEST}"

echo "==> [4/5] Seed config (will not overwrite)"
if [[ ! -f "${CONFIG_FILE}" ]]; then
  install -o beacon -g beacon -m 0640 \
    "${REPO_ROOT}/beacon/beacon.yaml.example" "${CONFIG_FILE}"
  echo "    seeded ${CONFIG_FILE} from beacon.yaml.example"
  echo "    >>> EDIT ${CONFIG_FILE} (pulse_server, node_id, node_name, api_key) <<<"
else
  echo "    ${CONFIG_FILE} already exists; left untouched"
fi

echo "==> [5/5] Enable service"
systemctl daemon-reload
systemctl enable beacon.service

if grep -qE '^\s*pulse_server:\s*$|^\s*api_key:\s*$' "${CONFIG_FILE}" 2>/dev/null; then
  echo ""
  echo "Config still has empty required fields; NOT starting the service."
  echo "Edit ${CONFIG_FILE}, then:  sudo systemctl start beacon"
else
  echo "Config looks populated; starting service."
  systemctl restart beacon
  systemctl --no-pager --full status beacon || true
fi

echo ""
echo "Done. Useful commands:"
echo "  sudo systemctl status beacon"
echo "  sudo journalctl -u beacon -f"
echo "  sudo systemctl reload beacon   # SIGHUP → hot-reload config"
