#!/usr/bin/env bash
# Bring up the catalogizer stack on thinker.local (rootless Podman).
#
# Reads deployment/thinker.local.env and launches postgres, redis, and
# catalog-api with the dev overlay image. Idempotent: recreates containers
# if they already exist.
#
# Run locally OR via:
#   ssh thinker.local bash < deployment/thinker-up.sh
#
# No sudo, no root. All ports bound to 127.0.0.1 for safety.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENV_FILE="${SCRIPT_DIR}/thinker.local.env"

if [[ -f "${ENV_FILE}" ]]; then
    set -a
    # shellcheck disable=SC1090
    source "${ENV_FILE}"
    set +a
fi

: "${CZ_THINKER_NETWORK:=catalogizer-thinker}"
: "${CZ_THINKER_POSTGRES_PORT:=5445}"
: "${CZ_THINKER_REDIS_PORT:=6391}"
: "${CZ_THINKER_API_PORT:=8092}"
: "${CZ_THINKER_POSTGRES_NAME:=cz-postgres-thinker}"
: "${CZ_THINKER_REDIS_NAME:=cz-redis-thinker}"
: "${CZ_THINKER_API_NAME:=cz-api-thinker}"
: "${CZ_THINKER_POSTGRES_CPUS:=1}"
: "${CZ_THINKER_POSTGRES_MEMORY:=2g}"
: "${CZ_THINKER_REDIS_CPUS:=1}"
: "${CZ_THINKER_REDIS_MEMORY:=1g}"
: "${CZ_THINKER_API_CPUS:=2}"
: "${CZ_THINKER_API_MEMORY:=4g}"
: "${POSTGRES_USER:=catalogizer}"
: "${POSTGRES_PASSWORD:=catalogizer}"
: "${POSTGRES_DB:=catalogizer}"
: "${JWT_SECRET:=dev-secret-for-operational-run-needs-32-plus-chars-to-be-valid}"
: "${ADMIN_PASSWORD:=admin123}"
: "${GIN_MODE:=release}"
: "${DB_TYPE:=postgres}"

API_IMAGE="${API_IMAGE:-localhost/catalogizer-api:latest}"
API_CONFIG="${API_CONFIG:-/tmp/catalogizer-run-config.json}"

echo "==> Target host: $(hostname)"
echo "==> API image: ${API_IMAGE}"
echo "==> Network: ${CZ_THINKER_NETWORK}"

podman network inspect "${CZ_THINKER_NETWORK}" >/dev/null 2>&1 \
    || podman network create "${CZ_THINKER_NETWORK}"

for name in "${CZ_THINKER_API_NAME}" "${CZ_THINKER_REDIS_NAME}" "${CZ_THINKER_POSTGRES_NAME}"; do
    if podman ps -a --format '{{.Names}}' | grep -qx "${name}"; then
        echo "==> removing stale ${name}"
        podman rm -f "${name}" >/dev/null
    fi
done

echo "==> starting postgres"
podman run -d \
    --name "${CZ_THINKER_POSTGRES_NAME}" \
    --network "${CZ_THINKER_NETWORK}" \
    -e POSTGRES_USER="${POSTGRES_USER}" \
    -e POSTGRES_PASSWORD="${POSTGRES_PASSWORD}" \
    -e POSTGRES_DB="${POSTGRES_DB}" \
    -p "127.0.0.1:${CZ_THINKER_POSTGRES_PORT}:5432" \
    --cpus="${CZ_THINKER_POSTGRES_CPUS}" \
    --memory="${CZ_THINKER_POSTGRES_MEMORY}" \
    docker.io/library/postgres:16-alpine >/dev/null

echo "==> starting redis"
podman run -d \
    --name "${CZ_THINKER_REDIS_NAME}" \
    --network "${CZ_THINKER_NETWORK}" \
    -p "127.0.0.1:${CZ_THINKER_REDIS_PORT}:6379" \
    --cpus="${CZ_THINKER_REDIS_CPUS}" \
    --memory="${CZ_THINKER_REDIS_MEMORY}" \
    docker.io/library/redis:7-alpine >/dev/null

MOUNT_ARGS=()
if [[ -f "${API_CONFIG}" ]]; then
    MOUNT_ARGS+=(-v "${API_CONFIG}:/app/config.json:ro")
fi

echo "==> starting catalog-api"
podman run -d \
    --name "${CZ_THINKER_API_NAME}" \
    --network "${CZ_THINKER_NETWORK}" \
    -p "127.0.0.1:${CZ_THINKER_API_PORT}:8080" \
    --cpus="${CZ_THINKER_API_CPUS}" \
    --memory="${CZ_THINKER_API_MEMORY}" \
    "${MOUNT_ARGS[@]}" \
    -e DATABASE_HOST="${CZ_THINKER_POSTGRES_NAME}" \
    -e DATABASE_PORT=5432 \
    -e DATABASE_NAME="${POSTGRES_DB}" \
    -e DATABASE_USER="${POSTGRES_USER}" \
    -e DATABASE_PASSWORD="${POSTGRES_PASSWORD}" \
    -e REDIS_HOST="${CZ_THINKER_REDIS_NAME}" \
    -e REDIS_PORT=6379 \
    -e JWT_SECRET="${JWT_SECRET}" \
    -e ADMIN_PASSWORD="${ADMIN_PASSWORD}" \
    -e PORT=8080 \
    -e GIN_MODE="${GIN_MODE}" \
    -e DB_TYPE="${DB_TYPE}" \
    "${API_IMAGE}" >/dev/null

echo "==> waiting for /health"
for _ in 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15; do
    if curl -sf "http://127.0.0.1:${CZ_THINKER_API_PORT}/health" >/dev/null 2>&1; then
        echo "==> catalog-api healthy on http://127.0.0.1:${CZ_THINKER_API_PORT}"
        exit 0
    fi
    sleep 1
done
echo "!! catalog-api did not become healthy within 15s"
podman logs --tail 40 "${CZ_THINKER_API_NAME}" || true
exit 1
