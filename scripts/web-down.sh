#!/usr/bin/env bash
# Stop the Catalogizer Web frontend service.
#
# Usage: ./scripts/web-down.sh

set -euo pipefail

CONTAINER_NAME="catalog-web"

if podman ps --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
    echo "[web] Stopping '${CONTAINER_NAME}'..."
    podman stop "${CONTAINER_NAME}" >/dev/null
    podman rm "${CONTAINER_NAME}" >/dev/null
    echo "[web] Stopped."
else
    echo "[web] Container '${CONTAINER_NAME}' is not running."
    # Clean up stopped container if exists
    if podman ps -a --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
        podman rm "${CONTAINER_NAME}" >/dev/null
        echo "[web] Removed stopped container."
    fi
fi
