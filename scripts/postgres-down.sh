#!/usr/bin/env bash
# Stop and remove the Catalogizer PostgreSQL container.
#
# Usage: ./scripts/postgres-down.sh

set -euo pipefail

CONTAINER_NAME="catalogizer-postgres"

if podman ps --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
    echo "Stopping PostgreSQL container '${CONTAINER_NAME}'..."
    podman stop "${CONTAINER_NAME}" >/dev/null
fi

if podman ps -a --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
    echo "Removing PostgreSQL container '${CONTAINER_NAME}'..."
    podman rm "${CONTAINER_NAME}" >/dev/null
fi

echo "PostgreSQL container stopped and removed."
echo "Note: Data volume 'catalogizer-pgdata' is preserved. To remove it: podman volume rm catalogizer-pgdata"
