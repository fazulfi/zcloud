#!/usr/bin/env bash
# Restore PostgreSQL custom dump or Redis RDB produced by backup.sh.
# Usage: restore.sh [--dry-run] <backupfile>
# Exit codes: 0 success, 2 usage error, 1 operational error.
set -euo pipefail

COMPOSE_FILE="${COMPOSE_FILE:-docker-compose.zcloud.yml}"
ENV_FILE="${ENV_FILE:-.env.zcloud}"
PROJECT_DIR="${PROJECT_DIR:-/opt/zcloud/deploy}"
DRY_RUN=false
backup_file=""
for arg in "$@"; do
  case "$arg" in
    --dry-run) DRY_RUN=true ;;
    -h|--help) sed -n '2,4p' "$0"; exit 0 ;;
    -*) printf 'Unknown option: %s\n' "$arg" >&2; exit 2 ;;
    *) [[ -z "$backup_file" ]] || { printf 'Only one backup file is accepted\n' >&2; exit 2; }; backup_file=$arg ;;
  esac
done
[[ -n "$backup_file" ]] || { printf 'Usage: %s [--dry-run] <backupfile>\n' "$0" >&2; exit 2; }
[[ -f "$backup_file" ]] || { printf 'Backup file not found: %s\n' "$backup_file" >&2; exit 1; }

case "$backup_file" in
  *.gpg) command -v gpg >/dev/null || { printf 'gpg is required for encrypted backups\n' >&2; exit 1; };;
  *.dump|*.rdb) ;;
  *) printf 'Unsupported backup extension: %s\n' "$backup_file" >&2; exit 2 ;;
esac

if "$DRY_RUN"; then printf 'DRY RUN: restore %s\n' "$backup_file"; exit 0; fi
read -r -p "Destructive restore of $backup_file. Type RESTORE to continue: " confirmation
[[ "$confirmation" == RESTORE ]] || { printf 'Restore cancelled\n'; exit 1; }
cd "$PROJECT_DIR"
set -a
# shellcheck disable=SC1091
source "$ENV_FILE"
set +a
compose=(docker compose -f "$COMPOSE_FILE" --env-file "$ENV_FILE")
restore_file="$backup_file"
tmp_file=""
cleanup() { [[ -n "$tmp_file" ]] && rm -f "$tmp_file"; }
trap cleanup EXIT
if [[ "$restore_file" == *.gpg ]]; then
  tmp_file=$(mktemp)
  gpg --batch --decrypt --output "$tmp_file" "$restore_file"
  restore_file="$tmp_file"
fi

case "$backup_file" in
  *postgres-*.dump|*postgres-*.dump.gpg) "${compose[@]}" exec -T postgres pg_restore --clean --if-exists --no-owner -U "${POSTGRES_USER:-zcloud}" -d "${POSTGRES_DB:-zcloud}" < "$restore_file" ;;
  *redis-*.rdb|*redis-*.rdb.gpg) "${compose[@]}" exec -T redis sh -c 'cat > /data/restore.rdb' < "$restore_file"; printf 'Redis RDB copied to /data/restore.rdb; restart Redis during the approved maintenance window to load it.\n' ;;
  *) printf 'Decrypted file has unsupported type\n' >&2; exit 2 ;;
esac
redis_auth=()
if [[ -n "${REDIS_PASSWORD:-}" ]]; then
  redis_auth=(-a "$REDIS_PASSWORD" --no-auth-warning)
fi
"${compose[@]}" exec -T postgres pg_isready -U "${POSTGRES_USER:-zcloud}" -d "${POSTGRES_DB:-zcloud}"
"${compose[@]}" exec -T redis redis-cli "${redis_auth[@]}" ping
row_count=$("${compose[@]}" exec -T postgres psql -U "${POSTGRES_USER:-zcloud}" -d "${POSTGRES_DB:-zcloud}" -Atc "SELECT COALESCE(SUM(n_live_tup), 0)::bigint FROM pg_stat_user_tables" | tr -d '\r')
printf 'Restore health checks passed; estimated PostgreSQL live rows: %s\n' "$row_count"
