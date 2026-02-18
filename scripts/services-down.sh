#!/usr/bin/env bash
# Stop all Catalogizer services (Web + API).
# Services are stopped in reverse dependency order.
#
# Usage: ./scripts/services-down.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

echo "=== Stopping Catalogizer Services ==="
echo ""

# Stop Web first
"${SCRIPT_DIR}/web-down.sh"

# Stop API
"${SCRIPT_DIR}/api-down.sh"

# Stop PostgreSQL
"${SCRIPT_DIR}/postgres-down.sh"

echo ""
echo "=== All Services Stopped ==="
