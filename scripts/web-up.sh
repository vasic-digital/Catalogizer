#!/usr/bin/env bash
# Start the Catalogizer Web frontend in a Podman container.
# Mounts catalog-web and required submodules, then runs the Vite dev server.
#
# The catalog-web package.json references submodules via file: links:
#   @vasic-digital/websocket-client -> file:../WebSocket-Client-TS
#   @vasic-digital/ui-components    -> file:../UI-Components-React
# npm creates symlinks in node_modules that resolve to ../../.. relative paths.
# Inside the container, these resolve to /WebSocket-Client-TS and
# /UI-Components-React, so we mount the submodule directories there.
#
# Usage: ./scripts/web-up.sh

set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
CONTAINER_NAME="catalog-web"
IMAGE_NAME="docker.io/library/node:18-alpine"

# Check if already running
if podman ps --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
    echo "[web] Container '${CONTAINER_NAME}' is already running."
    podman ps --filter "name=${CONTAINER_NAME}" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
    exit 0
fi

# Remove stopped container if exists
if podman ps -a --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
    echo "[web] Removing stopped container '${CONTAINER_NAME}'..."
    podman rm "${CONTAINER_NAME}" >/dev/null
fi

# Ensure node_modules are installed
if [[ ! -d "${PROJECT_ROOT}/catalog-web/node_modules" ]]; then
    echo "[web] Installing dependencies (npm install)..."
    podman run --rm \
        --network host \
        -v "${PROJECT_ROOT}/catalog-web:/app:z" \
        -v "${PROJECT_ROOT}/WebSocket-Client-TS:/WebSocket-Client-TS:ro" \
        -v "${PROJECT_ROOT}/UI-Components-React:/UI-Components-React:ro" \
        -w /app \
        "${IMAGE_NAME}" \
        sh -c "npm install" 2>&1 | tail -5
fi

# Start container with Vite dev server
# Mount submodules at /WebSocket-Client-TS and /UI-Components-React so that
# the node_modules symlinks (../../../WebSocket-Client-TS) resolve correctly.
echo "[web] Starting '${CONTAINER_NAME}' on port 3000..."
podman run -d \
    --name "${CONTAINER_NAME}" \
    --network host \
    -v "${PROJECT_ROOT}/catalog-web:/app:Z" \
    -v "${PROJECT_ROOT}/WebSocket-Client-TS:/WebSocket-Client-TS:ro" \
    -v "${PROJECT_ROOT}/UI-Components-React:/UI-Components-React:ro" \
    -w /app \
    "${IMAGE_NAME}" \
    sh -c "npx vite --host 0.0.0.0 --port 3000"

# Wait for the web server to be ready
echo -n "[web] Waiting for web app"
for i in $(seq 1 30); do
    if curl -sf http://localhost:3000/ >/dev/null 2>&1; then
        echo " OK"
        echo "[web] Catalogizer Web is running at http://localhost:3000"
        exit 0
    fi
    echo -n "."
    sleep 1
done

echo " TIMEOUT"
echo "[web] Warning: Web app did not respond in time. Checking logs..."
podman logs --tail 10 "${CONTAINER_NAME}" 2>&1
exit 1
