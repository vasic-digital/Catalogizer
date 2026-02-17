#!/usr/bin/env bash
# hash.sh - Source hash computation and change detection for the Build framework
# Computes SHA256 hashes of source files per component to detect changes
#
# Usage: source this file after common.sh and version.sh
#   source Build/lib/common.sh
#   source Build/lib/version.sh
#   source Build/lib/hash.sh
#
# Projects must define:
#   BUILD_COMPONENT_PATTERNS - associative array mapping component -> file patterns
#   BUILD_COMPONENTS - array of component names
#
# Example:
#   declare -A BUILD_COMPONENT_PATTERNS=(
#       ["my-api"]="*.go go.mod go.sum"
#       ["my-web"]="*.ts *.tsx *.js *.json *.html *.css"
#   )

set -euo pipefail

# Source common/version if not already loaded
if ! declare -f log_info &>/dev/null; then
    source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/common.sh"
fi
if ! declare -f get_component_hash &>/dev/null; then
    source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/version.sh"
fi

# Default exclude patterns (can be extended by projects via BUILD_HASH_EXCLUDE_DIRS)
BUILD_HASH_EXCLUDE_DIRS_DEFAULT=(
    "node_modules"
    "dist"
    "build"
    "target"
    ".git"
    "coverage"
    ".gradle"
    ".idea"
    ".vscode"
    "__pycache__"
)

# Compute source hash for a component
# Hash = SHA256 of (sorted file list -> individual sha256sums -> combined sha256sum)
compute_source_hash() {
    local component="$1"
    local comp_dir
    comp_dir="$(component_dir "$component")"

    if [[ ! -d "$comp_dir" ]]; then
        log_warn "Component directory not found: $comp_dir"
        echo "missing"
        return 0
    fi

    # Get patterns from project-defined associative array
    if [[ -z "${BUILD_COMPONENT_PATTERNS[$component]:-}" ]]; then
        log_error "No file patterns defined for component: $component"
        log_error "Define BUILD_COMPONENT_PATTERNS[$component] in your project config."
        echo "unknown"
        return 1
    fi
    local patterns="${BUILD_COMPONENT_PATTERNS[$component]}"

    # Merge default and project-specific exclude dirs
    local exclude_dirs=("${BUILD_HASH_EXCLUDE_DIRS_DEFAULT[@]}")
    if [[ -n "${BUILD_HASH_EXCLUDE_DIRS[*]:-}" ]]; then
        exclude_dirs+=("${BUILD_HASH_EXCLUDE_DIRS[@]}")
    fi

    # Build find command with include patterns and exclude dirs
    local find_args=()
    find_args+=("$comp_dir")

    # Add exclude directories
    local first_exclude=true
    for dir in "${exclude_dirs[@]}"; do
        if $first_exclude; then
            find_args+=("(" "-name" "$dir" "-prune" ")")
            first_exclude=false
        else
            find_args+=("-o" "(" "-name" "$dir" "-prune" ")")
        fi
    done

    # Add file pattern matches
    find_args+=("-o" "(")
    local first_pattern=true
    for pattern in $patterns; do
        if $first_pattern; then
            find_args+=("-name" "$pattern")
            first_pattern=false
        else
            find_args+=("-o" "-name" "$pattern")
        fi
    done
    find_args+=(")" "-type" "f" "-print")

    # Find files, sort, compute individual hashes, then combine
    local hash
    hash=$(find "${find_args[@]}" 2>/dev/null | \
        sort | \
        xargs -r sha256sum 2>/dev/null | \
        sha256sum | \
        awk '{print $1}')

    echo "$hash"
}

# Check if a component needs rebuilding
# Returns 0 (true) if rebuild needed, 1 (false) if up to date
needs_rebuild() {
    local component="$1"
    local force="${2:-false}"

    if [[ "$force" == "true" ]]; then
        return 0
    fi

    init_versions

    local stored_hash current_hash
    stored_hash="$(get_component_hash "$component")"
    current_hash="$(compute_source_hash "$component")"

    # No previous hash means first build
    if [[ -z "$stored_hash" || "$stored_hash" == "null" ]]; then
        return 0
    fi

    # Compare hashes
    if [[ "$stored_hash" != "$current_hash" ]]; then
        return 0
    fi

    # Up to date
    return 1
}

# Show hash status for all components
show_hash_status() {
    local force="${1:-false}"
    init_versions

    log_header "Component Change Detection Status"
    printf "%-25s %-12s %-14s\n" "COMPONENT" "STATUS" "HASH (first 12)"
    printf "%-25s %-12s %-14s\n" "─────────────────────────" "────────────" "──────────────"

    for component in "${BUILD_COMPONENTS[@]}"; do
        local current_hash stored_hash status
        current_hash="$(compute_source_hash "$component")"
        stored_hash="$(get_component_hash "$component")"
        local short_hash="${current_hash:0:12}"

        if [[ "$force" == "true" ]]; then
            status="FORCE"
        elif [[ -z "$stored_hash" || "$stored_hash" == "null" ]]; then
            status="NEW"
        elif [[ "$stored_hash" != "$current_hash" ]]; then
            status="CHANGED"
        else
            status="UP-TO-DATE"
        fi

        case "$status" in
            NEW|CHANGED|FORCE) printf "%-25s ${BUILD_YELLOW}%-12s${BUILD_NC} %s\n" "$component" "$status" "$short_hash" ;;
            UP-TO-DATE)        printf "%-25s ${BUILD_GREEN}%-12s${BUILD_NC} %s\n" "$component" "$status" "$short_hash" ;;
        esac
    done
    echo ""
}
