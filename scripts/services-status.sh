#!/usr/bin/env bash
# Show status of all Catalogizer services.
#
# Usage: ./scripts/services-status.sh

set -euo pipefail

SERVICES=("catalogizer-api" "catalog-web")
URLS=("http://localhost:8080/health" "http://localhost:3000/")

echo "=== Catalogizer Service Status ==="
echo ""

for i in "${!SERVICES[@]}"; do
    name="${SERVICES[$i]}"
    url="${URLS[$i]}"

    if podman ps --format '{{.Names}}' | grep -q "^${name}$"; then
        status=$(podman ps --filter "name=${name}" --format "{{.Status}}")

        # Check if the service actually responds
        if curl -sf "${url}" >/dev/null 2>&1; then
            echo "  [UP]   ${name} - ${status} - ${url}"
        else
            echo "  [WARN] ${name} - ${status} - not responding at ${url}"
        fi
    else
        echo "  [DOWN] ${name}"
    fi
done

echo ""
