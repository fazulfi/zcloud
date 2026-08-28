#!/usr/bin/env bash
# Backup PostgreSQL (custom format) and Redis for the zcloud compose stack.
# Usage: backup.sh [--dry-run]
# Exit codes: 0 success, 2 usage error, 1 operational error.
set -euo pipefail

COMPOSE_FILE="${COMPOSE_FILE:-docker-compose.zcloud.yml}"
ENV_FILE="${ENV_FILE:-.env.zcloud}"
BACKUP_DIR="${BACKUP_DIR:-/opt/zcloud/backups}"
PROJECT_DIR="${PROJECT_DIR:-/opt/zcloud/deploy}"
DRY_RUN=false
for arg in "$@"; do
  case "$arg" in
    --dry-run) DRY_RUN=true ;;
    -h|--help) sed -n '2,4p' "$0"; exit 0 ;;
    *) printf 'Unknown argument: %s\n' "$arg" >&2; exit 2 ;;
  esac
done

compose=(docker compose -f "$COMPOSE_FILE" --env-file "$ENV_FILE")
timestamp=$(date -u +%Y%m%dT%H%M%SZ)
pg_file="$BACKUP_DIR/postgres-$timestamp.dump"
redis_file="$BACKUP_DIR/redis-$timestamp.rdb"
mkdir -p "$BACKUP_DIR"
chmod 700 "$BACKUP_DIR"

if "$DRY_RUN"; then
  printf 'DRY RUN: PostgreSQL -> %s; Redis -> %s\n' "$pg_file" "$redis_file"
  exit 0
fi

cd "$PROJECT_DIR"
set -a
# shellcheck disable=SC1091
source "$ENV_FILE"
set +a
"${compose[@]}" exec -T postgres pg_dump -U "${POSTGRES_USER:-zcloud}" -d "${POSTGRES_DB:-zcloud}" -Fc > "$pg_file"
# Use explicit -a auth: --rdb (SYNC replication) ignores REDISCLI_AUTH, so
# export alone leaves the transfer unauthenticated when a password is set.
redis_auth=()
if [[ -n "${REDIS_PASSWORD:-}" ]]; then
  redis_auth=(-a "$REDIS_PASSWORD" --no-auth-warning)
fi
"${compose[@]}" exec -T redis redis-cli "${redis_auth[@]}" BGSAVE >/dev/null
for _ in {1..60}; do
  if "${compose[@]}" exec -T redis redis-cli "${redis_auth[@]}" LASTSAVE >/dev/null 2>&1; then break; fi
  sleep 1
done
"${compose[@]}" exec -T redis redis-cli "${redis_auth[@]}" --rdb - > "$redis_file"
chmod 600 "$pg_file" "$redis_file"

# Optional envelope encryption: set BACKUP_GPG_RECIPIENT to enable when gpg exists.
if [[ -n "${BACKUP_GPG_RECIPIENT:-}" && $(command -v gpg || true) ]]; then
  gpg --batch --yes --trust-model always --recipient "$BACKUP_GPG_RECIPIENT" --encrypt "$pg_file"
  gpg --batch --yes --trust-model always --recipient "$BACKUP_GPG_RECIPIENT" --encrypt "$redis_file"
  rm -f "$pg_file" "$redis_file"
fi

for kind in postgres redis; do
  extension=dump
  [[ "$kind" == redis ]] && extension=rdb
  declare -A weekly_seen=()
  declare -A monthly_seen=()
  while IFS= read -r file; do
    mtime=$(stat -c %Y "$file")
    age_days=$(( ( $(date +%s) - mtime ) / 86400 ))
    if (( age_days > 180 )); then
      rm -f "$file"
    elif (( age_days > 56 )); then
      period=$(date -u -d "@$mtime" +%Y-%m)
      [[ -n "${monthly_seen[$period]:-}" ]] && rm -f "$file" || monthly_seen[$period]=1
    elif (( age_days > 14 )); then
      period=$(date -u -d "@$mtime" +%G-%V)
      [[ -n "${weekly_seen[$period]:-}" ]] && rm -f "$file" || weekly_seen[$period]=1
    fi
  done < <(ls -1t "$BACKUP_DIR"/$kind-*.$extension 2>/dev/null || true)
done
printf 'Backup complete: %s\n' "$timestamp"
