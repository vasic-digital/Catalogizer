#!/usr/bin/env bash
set -euo pipefail

HOST="${1:?Usage: $0 <user@host>}"
shift

REMOTE_SCRIPT=$(cat <<'REMOTE_EOF'
set -e

echo "=== Checking build host prerequisites ==="

errors=0

check_cmd() {
    if command -v "$1" &>/dev/null; then
        echo "  [OK] $1 found: $(command -v "$1")"
    else
        echo "  [MISSING] $1 not found"
        errors=$((errors + 1))
    fi
}

check_runtime() {
    if command -v podman &>/dev/null; then
        echo "  [OK] podman found: $(command -v podman)"
        RUNTIME="podman"
    elif command -v docker &>/dev/null; then
        echo "  [OK] docker found: $(command -v docker)"
        echo "  [WARN] podman not found, using docker as container runtime"
        RUNTIME="docker"
        # Don't count as error since docker is acceptable
    else
        echo "  [MISSING] Neither podman nor docker found"
        errors=$((errors + 1))
    fi
}

check_runtime
check_cmd go
check_cmd node
check_cmd npm
check_cmd java
check_cmd rsync

echo ""
echo "=== Checking storage ==="
df -h /tmp | tail -1 | awk '{print "  /tmp: " $4 " available"}'

echo ""
echo "=== Checking memory ==="
free -h | grep Mem | awk '{print "  RAM: " $2 " total, " $4 " available"}'

echo ""
echo "=== Checking CPU ==="
nproc | awk '{print "  Cores: " $1}'

if [ "$errors" -gt 0 ]; then
    echo ""
    echo "ERROR: $errors prerequisite(s) missing"
    exit 1
fi

echo ""
echo "All prerequisites met."
REMOTE_EOF
)

echo "Checking host: $HOST"
ssh "$HOST" "$REMOTE_SCRIPT"
