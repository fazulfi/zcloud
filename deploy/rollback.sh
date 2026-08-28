#!/usr/bin/env bash
# Roll back zrouter to a previously published immutable image tag.
# Usage: rollback.sh [--dry-run] <previous_tag>
# Exit codes: 0 success, 2 usage error, 1 operational error.
set -euo pipefail

COMPOSE_FILE="${COMPOSE_FILE:-docker-compose.zcloud.yml}"
ENV_FILE="${ENV_FILE:-.env.zcloud}"
PROJECT_DIR="${PROJECT_DIR:-/opt/zcloud/deploy}"
DRY_RUN=false
tag=""
for arg in "$@"; do
  case "$arg" in
    --dry-run) DRY_RUN=true ;;
    -h|--help) sed -n '2,4p' "$0"; exit 0 ;;
    -*) printf 'Unknown option: %s\n' "$arg" >&2; exit 2 ;;
    *) [[ -z "$tag" ]] || { printf 'Only one image tag is accepted\n' >&2; exit 2; }; tag=$arg ;;
  esac
done
[[ -n "$tag" ]] || { printf 'Usage: %s [--dry-run] <previous_tag>\n' "$0" >&2; exit 2; }
[[ "$tag" != *:* && "$tag" != */* ]] || { printf 'Use a version tag, not an image reference\n' >&2; exit 2; }
if "$DRY_RUN"; then printf 'DRY RUN: set ZCLOUD_VERSION=%s and recreate zrouter\n' "$tag"; exit 0; fi
read -r -p "Rollback zrouter to $tag? Type ROLLBACK to continue: " confirmation
[[ "$confirmation" == ROLLBACK ]] || { printf 'Rollback cancelled\n'; exit 1; }
cd "$PROJECT_DIR"
compose=(docker compose -f "$COMPOSE_FILE" --env-file "$ENV_FILE")
current_version=$(grep -E '^ZCLOUD_VERSION=' "$ENV_FILE" | cut -d= -f2- || true)
printf 'Rolling back from %s to %s\n' "${current_version:-unknown}" "$tag"
"${compose[@]}" stop zrouter
ZCLOUD_VERSION="$tag" "${compose[@]}" up -d --no-build zrouter
for _ in {1..30}; do
  if curl -fsS --max-time 5 http://127.0.0.1:18080/health/ready >/dev/null; then
    printf 'Rollback health check passed for %s\n' "$tag"
    exit 0
  fi
  sleep 2
done
printf 'Rollback health check failed for %s\n' "$tag" >&2
"${compose[@]}" logs --tail=80 zrouter >&2 || true
exit 1
