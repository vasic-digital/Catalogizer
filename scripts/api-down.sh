#!/usr/bin/env bash
# Stop the Catalogizer API service.
#
# Usage: ./scripts/api-down.sh

set -euo pipefail

CONTAINER_NAME="catalogizer-api"

if podman ps --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
    echo "[api] Stopping '${CONTAINER_NAME}'..."
    podman stop "${CONTAINER_NAME}" >/dev/null
    podman rm "${CONTAINER_NAME}" >/dev/null
    echo "[api] Stopped."
else
    echo "[api] Container '${CONTAINER_NAME}' is not running."
    # Clean up stopped container if exists
    if podman ps -a --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
        podman rm "${CONTAINER_NAME}" >/dev/null
        echo "[api] Removed stopped container."
    fi
fi
