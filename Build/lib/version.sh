#!/usr/bin/env bash
# version.sh - Generic version management for the Build framework
# Reads/writes a versions.json file with semantic versioning and build numbers
#
# Usage: source this file after common.sh
#   source Build/lib/common.sh
#   source Build/lib/version.sh
#
# Requires: BUILD_PROJECT_ROOT to be set

set -euo pipefail

# Source common if not already loaded
if ! declare -f log_info &>/dev/null; then
    source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/common.sh"
fi

# Default versions file location (can be overridden)
BUILD_VERSIONS_FILE="${BUILD_VERSIONS_FILE:-$BUILD_PROJECT_ROOT/versions.json}"

# Initialize versions.json with given components
# Usage: init_versions "comp1" "comp2" "comp3" ...
# If no args, uses BUILD_COMPONENTS array
init_versions() {
    if [[ -f "$BUILD_VERSIONS_FILE" ]]; then
        return 0
    fi

    local components=("$@")
    if [[ ${#components[@]} -eq 0 ]]; then
        if [[ -z "${BUILD_COMPONENTS[*]:-}" ]]; then
            log_error "No components specified. Pass as args or set BUILD_COMPONENTS."
            return 1
        fi
        components=("${BUILD_COMPONENTS[@]}")
    fi

    log_info "Initializing versions.json with ${#components[@]} components"

    # Build components JSON
    local comp_json=""
    for comp in "${components[@]}"; do
        if [[ -n "$comp_json" ]]; then
            comp_json+=","
        fi
        comp_json+="
    \"$comp\": {
      \"last_build_number\": 0,
      \"last_build_date\": null,
      \"last_source_hash\": null,
      \"last_git_commit\": null
    }"
    done

    cat > "$BUILD_VERSIONS_FILE" <<EOF
{
  "schema_version": 1,
  "global": {
    "major": 1,
    "minor": 0,
    "patch": 0,
    "build_number": 0
  },
  "components": {$comp_json
  }
}
EOF
}

# Read a value from versions.json using python3
_json_read() {
    local query="$1"
    python3 -c "
import json, sys
with open('$BUILD_VERSIONS_FILE') as f:
    d = json.load(f)
keys = '$query'.split('.')
v = d
for k in keys:
    v = v[k]
print(v if v is not None else '')
" 2>/dev/null
}

# Write a value to versions.json
_json_write() {
    local query="$1"
    local value="$2"
    python3 -c "
import json
with open('$BUILD_VERSIONS_FILE') as f:
    d = json.load(f)
keys = '$query'.split('.')
obj = d
for k in keys[:-1]:
    obj = obj[k]
val = '$value'
if val == 'null' or val == '':
    obj[keys[-1]] = None
elif val.isdigit():
    obj[keys[-1]] = int(val)
else:
    obj[keys[-1]] = val
with open('$BUILD_VERSIONS_FILE', 'w') as f:
    json.dump(d, f, indent=2)
    f.write('\n')
"
}

# Get the global version string (e.g., "1.0.0")
get_version() {
    local major minor patch
    major="$(_json_read "global.major")"
    minor="$(_json_read "global.minor")"
    patch="$(_json_read "global.patch")"
    echo "${major}.${minor}.${patch}"
}

# Get the global build number
get_build_number() {
    _json_read "global.build_number"
}

# Get full version string (e.g., "v1.0.0-build.3")
get_version_string() {
    local version build_number
    version="$(get_version)"
    build_number="$(get_build_number)"
    echo "v${version}-build.${build_number}"
}

# Increment version (major, minor, or patch)
bump_version() {
    local part="$1"
    local major minor patch
    major="$(_json_read "global.major")"
    minor="$(_json_read "global.minor")"
    patch="$(_json_read "global.patch")"

    case "$part" in
        major)
            major=$((major + 1))
            minor=0
            patch=0
            ;;
        minor)
            minor=$((minor + 1))
            patch=0
            ;;
        patch)
            patch=$((patch + 1))
            ;;
        *)
            log_error "Invalid bump type: $part (use major, minor, or patch)"
            return 1
            ;;
    esac

    _json_write "global.major" "$major"
    _json_write "global.minor" "$minor"
    _json_write "global.patch" "$patch"
    log_info "Version bumped to ${major}.${minor}.${patch}"
}

# Increment build number and return the new value
increment_build_number() {
    local current
    current="$(get_build_number)"
    local next=$((current + 1))
    _json_write "global.build_number" "$next"
    echo "$next"
}

# Get component's last source hash
get_component_hash() {
    local component="$1"
    _json_read "components.${component}.last_source_hash"
}

# Get component's last build number
get_component_build_number() {
    local component="$1"
    _json_read "components.${component}.last_build_number"
}

# Update component build state after successful build
update_component_state() {
    local component="$1"
    local build_number="$2"
    local source_hash="$3"

    _json_write "components.${component}.last_build_number" "$build_number"
    _json_write "components.${component}.last_build_date" "$(build_timestamp)"
    _json_write "components.${component}.last_source_hash" "$source_hash"
    _json_write "components.${component}.last_git_commit" "$(git_short_commit)"
}
