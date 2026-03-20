#!/bin/bash
# Module Challenge Runner
# Runs compile, unit, and functionality challenges for all submodules

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MODULES=(
    Auth
    Cache
    Challenges
    Concurrency
    Database
    EventBus
    Memory
    Observability
    Security
    Storage
    Streaming
)

CHALLENGE_TYPES=("compile" "unit" "functionality")

run_challenge() {
    local module=$1
    local type=$2
    
    if [[ $# -lt 1 ]] || [[ $# -lt 2 ]]; then
        echo "Usage: $0 <script_name> <challenge_type>"
        echo "Challenge types: ${CHALLENGE_TYPES[*]}"
        exit 1
    fi
    
    module=$1
    type=$2
    
    module_dir="${SCRIPT_DIR}/../${module}"
    script_path="${module_dir}/challenges/scripts/${module,,*}_challenge.sh"
    
    if [[ ! -f "$script_path" ]]; then
        echo "No challenge script for ${module}: ${type}"
        exit 1
    fi
    
    echo "Running ${module} ${type} challenge..."
    chmod +x "$script_path"
}

# Run all compile challenges
echo "=== Compile Challenges ==="
for module in "${MODULES[@]}"; do
    run_challenge "$module" compile
done

# Run all unit challenges  
echo "=== Unit Challenges ==="
for module in "${MODULES[@]}"; do
    run_challenge "$module" unit
done

# Run all functionality challenges
echo "=== Functionality Challenges ==="
for module in "${MODULES[@]}"; do
    run_challenge "$module" functionality
done

echo "=== Module Challenges Complete ==="
