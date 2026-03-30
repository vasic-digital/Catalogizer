#!/bin/bash
# Deploy llama.cpp with RPC support on all vision hosts.
# Usage: ./scripts/deploy-vision-hosts.sh
#
# Reads HELIX_VISION_HOSTS and HELIX_VISION_USER from
# environment or HelixQA/.env. Each host gets llama.cpp
# cloned, built with -DGGML_RPC=ON (and -DGGML_CUDA=ON
# if available).
#
# Prerequisites: SSH key-based auth to all hosts.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Source .env if available.
if [ -f "$PROJECT_ROOT/HelixQA/.env" ]; then
    set -a
    # shellcheck disable=SC1091
    source "$PROJECT_ROOT/HelixQA/.env"
    set +a
fi

HOSTS="${HELIX_VISION_HOSTS:-thinker.local,amber.local}"
USER="${HELIX_VISION_MULTI_USER:-${HELIX_VISION_USER:-milosvasic}}"

echo "=== Deploying llama.cpp with RPC on vision hosts ==="
echo "Hosts: $HOSTS"
echo "User:  $USER"
echo ""

SUCCESS=0
FAIL=0

for host in ${HOSTS//,/ }; do
    echo "--- Deploying on $host ---"

    if ! ssh -o BatchMode=yes -o ConnectTimeout=10 \
         -o StrictHostKeyChecking=no \
         "$USER@$host" true 2>/dev/null; then
        echo "ERROR: Cannot reach $host via SSH, skipping."
        FAIL=$((FAIL + 1))
        continue
    fi

    ssh -o BatchMode=yes -o StrictHostKeyChecking=no \
        "$USER@$host" bash -s <<'REMOTE'
set -euo pipefail

echo "Host: $(hostname)"

# Clone or update llama.cpp.
if [ ! -d ~/llama.cpp ]; then
    echo "Cloning llama.cpp..."
    git clone --depth 1 \
        https://github.com/ggerganov/llama.cpp.git \
        ~/llama.cpp
else
    echo "Updating llama.cpp..."
    cd ~/llama.cpp && git pull --ff-only || true
fi

cd ~/llama.cpp

# Build with RPC + CUDA, fallback to RPC-only.
echo "Building llama.cpp with RPC support..."
if cmake -B build \
    -DGGML_RPC=ON \
    -DGGML_CUDA=ON \
    -DCMAKE_BUILD_TYPE=Release 2>/dev/null; then
    echo "CUDA detected, building with GPU support."
else
    echo "No CUDA, building CPU-only with RPC."
    cmake -B build \
        -DGGML_RPC=ON \
        -DCMAKE_BUILD_TYPE=Release
fi

cmake --build build --config Release -j"$(nproc)"

# Verify binaries.
echo ""
if [ -x build/bin/llama-server ]; then
    echo "llama-server: OK"
else
    echo "llama-server: MISSING" >&2
    exit 1
fi
if [ -x build/bin/rpc-server ]; then
    echo "rpc-server:   OK"
else
    echo "rpc-server:   MISSING (RPC build may have failed)" >&2
    exit 1
fi

echo "llama.cpp deployed successfully on $(hostname)"
REMOTE

    if [ $? -eq 0 ]; then
        echo "--- $host: SUCCESS ---"
        SUCCESS=$((SUCCESS + 1))
    else
        echo "--- $host: FAILED ---"
        FAIL=$((FAIL + 1))
    fi
    echo ""
done

echo "=== Deployment complete ==="
echo "Success: $SUCCESS, Failed: $FAIL"

if [ "$FAIL" -gt 0 ]; then
    exit 1
fi
