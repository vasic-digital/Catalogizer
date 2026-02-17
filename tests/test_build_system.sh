#!/usr/bin/env bash
# test_build_system.sh - Tests for the Catalogizer build system
# Validates version management, hash computation, and directory structure

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Test counters
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# Temp directory for test artifacts
TEST_TMP=""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# ─── Test Helpers ────────────────────────────────────────────────────

setup() {
    TEST_TMP="$(mktemp -d)"
    # Copy versions.json template for isolated testing
    export BUILD_PROJECT_ROOT="$TEST_TMP/project"
    mkdir -p "$BUILD_PROJECT_ROOT"

    # Create minimal component dirs for hash tests
    for comp in catalog-api catalog-web catalogizer-api-client catalogizer-desktop installer-wizard catalogizer-android catalogizer-androidtv; do
        mkdir -p "$BUILD_PROJECT_ROOT/$comp"
    done

    # Add some test source files
    echo 'package main' > "$BUILD_PROJECT_ROOT/catalog-api/main.go"
    echo 'module catalogizer' > "$BUILD_PROJECT_ROOT/catalog-api/go.mod"
    echo 'import React from "react"' > "$BUILD_PROJECT_ROOT/catalog-web/App.tsx"
    echo '{"name": "test"}' > "$BUILD_PROJECT_ROOT/catalog-web/package.json"
    echo 'export const api = {}' > "$BUILD_PROJECT_ROOT/catalogizer-api-client/index.ts"
    echo '{"name": "api-client"}' > "$BUILD_PROJECT_ROOT/catalogizer-api-client/package.json"
    echo 'fn main() {}' > "$BUILD_PROJECT_ROOT/catalogizer-desktop/main.rs"
    echo '[package]' > "$BUILD_PROJECT_ROOT/catalogizer-desktop/Cargo.toml"
    echo 'fn main() {}' > "$BUILD_PROJECT_ROOT/installer-wizard/main.rs"
    echo '[package]' > "$BUILD_PROJECT_ROOT/installer-wizard/Cargo.toml"
    echo 'class Main {}' > "$BUILD_PROJECT_ROOT/catalogizer-android/Main.kt"
    echo 'plugins {}' > "$BUILD_PROJECT_ROOT/catalogizer-android/build.gradle.kts"
    echo 'class TVMain {}' > "$BUILD_PROJECT_ROOT/catalogizer-androidtv/TVMain.kt"
    echo 'plugins {}' > "$BUILD_PROJECT_ROOT/catalogizer-androidtv/build.gradle.kts"

    # Create a fake git repo for git helpers
    (cd "$BUILD_PROJECT_ROOT" && git init -q && git add -A && git commit -q -m "test" --allow-empty)

    # Create Build framework symlink (or copy)
    cp -r "$PROJECT_ROOT/Build" "$BUILD_PROJECT_ROOT/Build"

    # Set versions file location
    export BUILD_VERSIONS_FILE="$BUILD_PROJECT_ROOT/versions.json"

    # Source project config (we redefine for test environment)
    BUILD_COMPONENTS=(
        "catalog-api"
        "catalog-web"
        "catalogizer-api-client"
        "catalogizer-desktop"
        "installer-wizard"
        "catalogizer-android"
        "catalogizer-androidtv"
    )

    declare -gA BUILD_COMPONENT_PATTERNS=(
        ["catalog-api"]="*.go go.mod go.sum"
        ["catalog-web"]="*.ts *.tsx *.js *.json *.html *.css"
        ["catalogizer-api-client"]="*.ts *.js *.json"
        ["catalogizer-desktop"]="*.ts *.tsx *.js *.json *.html *.css *.rs *.toml"
        ["installer-wizard"]="*.ts *.tsx *.js *.json *.html *.css *.rs *.toml"
        ["catalogizer-android"]="*.kt *.java *.xml *.gradle.kts *.properties"
        ["catalogizer-androidtv"]="*.kt *.java *.xml *.gradle.kts *.properties"
    )

    # Source Build framework
    source "$BUILD_PROJECT_ROOT/Build/lib/common.sh"
    source "$BUILD_PROJECT_ROOT/Build/lib/version.sh"
    source "$BUILD_PROJECT_ROOT/Build/lib/hash.sh"
}

teardown() {
    if [[ -n "$TEST_TMP" && -d "$TEST_TMP" ]]; then
        rm -rf "$TEST_TMP"
    fi
}

assert_eq() {
    local expected="$1"
    local actual="$2"
    local message="${3:-}"
    TESTS_RUN=$((TESTS_RUN + 1))
    if [[ "$expected" == "$actual" ]]; then
        TESTS_PASSED=$((TESTS_PASSED + 1))
        echo -e "  ${GREEN}PASS${NC}: $message"
    else
        TESTS_FAILED=$((TESTS_FAILED + 1))
        echo -e "  ${RED}FAIL${NC}: $message"
        echo -e "    Expected: '$expected'"
        echo -e "    Actual:   '$actual'"
    fi
}

assert_true() {
    local message="${1:-}"
    TESTS_RUN=$((TESTS_RUN + 1))
    # The condition should have been evaluated before calling this
    TESTS_PASSED=$((TESTS_PASSED + 1))
    echo -e "  ${GREEN}PASS${NC}: $message"
}

assert_false() {
    local message="${1:-}"
    TESTS_RUN=$((TESTS_RUN + 1))
    TESTS_FAILED=$((TESTS_FAILED + 1))
    echo -e "  ${RED}FAIL${NC}: $message"
}

# ─── Tests ───────────────────────────────────────────────────────────

test_versions_init() {
    echo -e "\n${YELLOW}Test: versions.json initialization${NC}"

    # Remove if exists
    rm -f "$BUILD_VERSIONS_FILE"

    init_versions
    if [[ -f "$BUILD_VERSIONS_FILE" ]]; then
        assert_true "versions.json created"
    else
        assert_false "versions.json should exist after init"
    fi

    # Verify schema version
    local schema
    schema="$(_json_read "schema_version")"
    assert_eq "1" "$schema" "schema_version is 1"
}

test_version_read() {
    echo -e "\n${YELLOW}Test: version reading${NC}"

    rm -f "$BUILD_VERSIONS_FILE"
    init_versions

    local version
    version="$(get_version)"
    assert_eq "1.0.0" "$version" "initial version is 1.0.0"

    local build_number
    build_number="$(get_build_number)"
    assert_eq "0" "$build_number" "initial build number is 0"

    local version_string
    version_string="$(get_version_string)"
    assert_eq "v1.0.0-build.0" "$version_string" "initial version string"
}

test_version_bump() {
    echo -e "\n${YELLOW}Test: version bumping${NC}"

    rm -f "$BUILD_VERSIONS_FILE"
    init_versions

    bump_version "patch"
    assert_eq "1.0.1" "$(get_version)" "patch bump -> 1.0.1"

    bump_version "minor"
    assert_eq "1.1.0" "$(get_version)" "minor bump -> 1.1.0"

    bump_version "major"
    assert_eq "2.0.0" "$(get_version)" "major bump -> 2.0.0"
}

test_build_number_increment() {
    echo -e "\n${YELLOW}Test: build number increment${NC}"

    rm -f "$BUILD_VERSIONS_FILE"
    init_versions

    local n1
    n1="$(increment_build_number)"
    assert_eq "1" "$n1" "first increment -> 1"

    local n2
    n2="$(increment_build_number)"
    assert_eq "2" "$n2" "second increment -> 2"

    local n3
    n3="$(increment_build_number)"
    assert_eq "3" "$n3" "third increment -> 3"
}

test_component_state() {
    echo -e "\n${YELLOW}Test: component state management${NC}"

    rm -f "$BUILD_VERSIONS_FILE"
    init_versions

    # Initial state
    local hash
    hash="$(get_component_hash "catalog-api")"
    assert_eq "" "$hash" "initial hash is empty"

    # Update state
    update_component_state "catalog-api" "5" "abc123def456"
    local stored_hash
    stored_hash="$(get_component_hash "catalog-api")"
    assert_eq "abc123def456" "$stored_hash" "stored hash matches"

    local stored_build
    stored_build="$(get_component_build_number "catalog-api")"
    assert_eq "5" "$stored_build" "stored build number matches"
}

test_hash_computation() {
    echo -e "\n${YELLOW}Test: source hash computation${NC}"

    local hash1
    hash1="$(compute_source_hash "catalog-api")"
    TESTS_RUN=$((TESTS_RUN + 1))
    if [[ -n "$hash1" && "$hash1" != "missing" && "$hash1" != "unknown" ]]; then
        TESTS_PASSED=$((TESTS_PASSED + 1))
        echo -e "  ${GREEN}PASS${NC}: catalog-api hash is non-empty: ${hash1:0:12}..."
    else
        TESTS_FAILED=$((TESTS_FAILED + 1))
        echo -e "  ${RED}FAIL${NC}: catalog-api hash should be valid, got: $hash1"
    fi

    # Same input should produce same hash
    local hash2
    hash2="$(compute_source_hash "catalog-api")"
    assert_eq "$hash1" "$hash2" "hash is deterministic"

    # Modifying a file should change the hash
    echo 'package main; func init() {}' >> "$BUILD_PROJECT_ROOT/catalog-api/main.go"
    local hash3
    hash3="$(compute_source_hash "catalog-api")"
    TESTS_RUN=$((TESTS_RUN + 1))
    if [[ "$hash1" != "$hash3" ]]; then
        TESTS_PASSED=$((TESTS_PASSED + 1))
        echo -e "  ${GREEN}PASS${NC}: hash changes when source changes"
    else
        TESTS_FAILED=$((TESTS_FAILED + 1))
        echo -e "  ${RED}FAIL${NC}: hash should change when source changes"
    fi
}

test_needs_rebuild() {
    echo -e "\n${YELLOW}Test: change detection (needs_rebuild)${NC}"

    rm -f "$BUILD_VERSIONS_FILE"
    init_versions

    # First build always needed
    TESTS_RUN=$((TESTS_RUN + 1))
    if needs_rebuild "catalog-api"; then
        TESTS_PASSED=$((TESTS_PASSED + 1))
        echo -e "  ${GREEN}PASS${NC}: first build detected as needed"
    else
        TESTS_FAILED=$((TESTS_FAILED + 1))
        echo -e "  ${RED}FAIL${NC}: first build should always be needed"
    fi

    # After storing the hash, should not need rebuild
    local current_hash
    current_hash="$(compute_source_hash "catalog-api")"
    update_component_state "catalog-api" "1" "$current_hash"

    TESTS_RUN=$((TESTS_RUN + 1))
    if needs_rebuild "catalog-api"; then
        TESTS_FAILED=$((TESTS_FAILED + 1))
        echo -e "  ${RED}FAIL${NC}: should not need rebuild after storing hash"
    else
        TESTS_PASSED=$((TESTS_PASSED + 1))
        echo -e "  ${GREEN}PASS${NC}: up-to-date component skipped correctly"
    fi

    # Force always triggers rebuild
    TESTS_RUN=$((TESTS_RUN + 1))
    if needs_rebuild "catalog-api" "true"; then
        TESTS_PASSED=$((TESTS_PASSED + 1))
        echo -e "  ${GREEN}PASS${NC}: --force triggers rebuild"
    else
        TESTS_FAILED=$((TESTS_FAILED + 1))
        echo -e "  ${RED}FAIL${NC}: --force should trigger rebuild"
    fi
}

test_component_validation() {
    echo -e "\n${YELLOW}Test: component validation${NC}"

    TESTS_RUN=$((TESTS_RUN + 1))
    if validate_component "catalog-api" 2>/dev/null; then
        TESTS_PASSED=$((TESTS_PASSED + 1))
        echo -e "  ${GREEN}PASS${NC}: valid component accepted"
    else
        TESTS_FAILED=$((TESTS_FAILED + 1))
        echo -e "  ${RED}FAIL${NC}: valid component should be accepted"
    fi

    TESTS_RUN=$((TESTS_RUN + 1))
    if validate_component "nonexistent-component" 2>/dev/null; then
        TESTS_FAILED=$((TESTS_FAILED + 1))
        echo -e "  ${RED}FAIL${NC}: invalid component should be rejected"
    else
        TESTS_PASSED=$((TESTS_PASSED + 1))
        echo -e "  ${GREEN}PASS${NC}: invalid component rejected"
    fi
}

test_release_dir_creation() {
    echo -e "\n${YELLOW}Test: release directory creation${NC}"

    local dir
    dir="$(create_release_dir "catalog-api" "linux-amd64" "v1.0.0-build.1")"

    TESTS_RUN=$((TESTS_RUN + 1))
    if [[ -d "$dir" ]]; then
        TESTS_PASSED=$((TESTS_PASSED + 1))
        echo -e "  ${GREEN}PASS${NC}: release directory created"
    else
        TESTS_FAILED=$((TESTS_FAILED + 1))
        echo -e "  ${RED}FAIL${NC}: release directory should exist"
    fi

    # Check path structure
    local expected="$BUILD_PROJECT_ROOT/releases/catalog-api/linux-amd64/v1.0.0-build.1"
    assert_eq "$expected" "$dir" "release dir path is correct"
}

test_build_info_generation() {
    echo -e "\n${YELLOW}Test: BUILD_INFO.json generation${NC}"

    local dir
    dir="$(create_release_dir "test-comp" "linux" "v1.0.0-build.1")"

    generate_build_info "$dir" "test-comp" "linux" "1.0.0" "1" "v1.0.0-build.1" "abc123"

    TESTS_RUN=$((TESTS_RUN + 1))
    if [[ -f "$dir/BUILD_INFO.json" ]]; then
        TESTS_PASSED=$((TESTS_PASSED + 1))
        echo -e "  ${GREEN}PASS${NC}: BUILD_INFO.json created"
    else
        TESTS_FAILED=$((TESTS_FAILED + 1))
        echo -e "  ${RED}FAIL${NC}: BUILD_INFO.json should exist"
    fi

    # Verify JSON structure
    local component
    component=$(python3 -c "import json; print(json.load(open('$dir/BUILD_INFO.json'))['component'])")
    assert_eq "test-comp" "$component" "component field correct"

    local version
    version=$(python3 -c "import json; print(json.load(open('$dir/BUILD_INFO.json'))['version'])")
    assert_eq "1.0.0" "$version" "version field correct"
}

test_checksum_generation() {
    echo -e "\n${YELLOW}Test: SHA256SUM generation${NC}"

    local dir
    dir="$(create_release_dir "test-checksum" "linux" "v1.0.0-build.1")"

    echo "binary content" > "$dir/test-binary"
    echo "config content" > "$dir/test-config"

    generate_checksums "$dir"

    TESTS_RUN=$((TESTS_RUN + 1))
    if [[ -f "$dir/SHA256SUM" ]]; then
        TESTS_PASSED=$((TESTS_PASSED + 1))
        echo -e "  ${GREEN}PASS${NC}: SHA256SUM created"
    else
        TESTS_FAILED=$((TESTS_FAILED + 1))
        echo -e "  ${RED}FAIL${NC}: SHA256SUM should exist"
    fi

    # Verify checksums are valid
    TESTS_RUN=$((TESTS_RUN + 1))
    if (cd "$dir" && sha256sum -c SHA256SUM &>/dev/null); then
        TESTS_PASSED=$((TESTS_PASSED + 1))
        echo -e "  ${GREEN}PASS${NC}: checksums verify correctly"
    else
        TESTS_FAILED=$((TESTS_FAILED + 1))
        echo -e "  ${RED}FAIL${NC}: checksums should verify"
    fi
}

test_runtime_detection() {
    echo -e "\n${YELLOW}Test: container runtime detection${NC}"

    local runtime
    runtime="$(detect_runtime 2>/dev/null || echo "none")"

    TESTS_RUN=$((TESTS_RUN + 1))
    if [[ "$runtime" == "podman" || "$runtime" == "docker" || "$runtime" == "none" ]]; then
        TESTS_PASSED=$((TESTS_PASSED + 1))
        echo -e "  ${GREEN}PASS${NC}: runtime detection returned: $runtime"
    else
        TESTS_FAILED=$((TESTS_FAILED + 1))
        echo -e "  ${RED}FAIL${NC}: unexpected runtime: $runtime"
    fi
}

test_all_components_have_patterns() {
    echo -e "\n${YELLOW}Test: all components have hash patterns${NC}"

    for comp in "${BUILD_COMPONENTS[@]}"; do
        TESTS_RUN=$((TESTS_RUN + 1))
        if [[ -n "${BUILD_COMPONENT_PATTERNS[$comp]:-}" ]]; then
            TESTS_PASSED=$((TESTS_PASSED + 1))
            echo -e "  ${GREEN}PASS${NC}: $comp has patterns defined"
        else
            TESTS_FAILED=$((TESTS_FAILED + 1))
            echo -e "  ${RED}FAIL${NC}: $comp missing patterns"
        fi
    done
}

test_cli_help() {
    echo -e "\n${YELLOW}Test: CLI --help flag${NC}"

    # Source orchestrator for usage function
    source "$BUILD_PROJECT_ROOT/Build/lib/orchestrator.sh"

    TESTS_RUN=$((TESTS_RUN + 1))
    local help_output
    help_output="$(build_usage 2>&1)"
    if echo "$help_output" | grep -q "component"; then
        TESTS_PASSED=$((TESTS_PASSED + 1))
        echo -e "  ${GREEN}PASS${NC}: --help shows component option"
    else
        TESTS_FAILED=$((TESTS_FAILED + 1))
        echo -e "  ${RED}FAIL${NC}: --help should mention component"
    fi
}

test_git_helpers() {
    echo -e "\n${YELLOW}Test: git helpers${NC}"

    local commit
    commit="$(git_short_commit)"
    TESTS_RUN=$((TESTS_RUN + 1))
    if [[ -n "$commit" && "$commit" != "unknown" ]]; then
        TESTS_PASSED=$((TESTS_PASSED + 1))
        echo -e "  ${GREEN}PASS${NC}: git_short_commit returns: $commit"
    else
        TESTS_FAILED=$((TESTS_FAILED + 1))
        echo -e "  ${RED}FAIL${NC}: git_short_commit should return a value"
    fi

    local branch
    branch="$(git_branch)"
    TESTS_RUN=$((TESTS_RUN + 1))
    if [[ -n "$branch" ]]; then
        TESTS_PASSED=$((TESTS_PASSED + 1))
        echo -e "  ${GREEN}PASS${NC}: git_branch returns: $branch"
    else
        TESTS_FAILED=$((TESTS_FAILED + 1))
        echo -e "  ${RED}FAIL${NC}: git_branch should return a value"
    fi
}

# ─── Main ────────────────────────────────────────────────────────────

main() {
    echo -e "\n${YELLOW}════════════════════════════════════════════════════════════${NC}"
    echo -e "${YELLOW}  Catalogizer Build System Tests${NC}"
    echo -e "${YELLOW}════════════════════════════════════════════════════════════${NC}"

    setup

    test_versions_init
    test_version_read
    test_version_bump
    test_build_number_increment
    test_component_state
    test_hash_computation
    test_needs_rebuild
    test_component_validation
    test_release_dir_creation
    test_build_info_generation
    test_checksum_generation
    test_runtime_detection
    test_all_components_have_patterns
    test_cli_help
    test_git_helpers

    teardown

    echo -e "\n${YELLOW}════════════════════════════════════════════════════════════${NC}"
    echo -e "  Results: ${GREEN}$TESTS_PASSED passed${NC}, ${RED}$TESTS_FAILED failed${NC}, $TESTS_RUN total"
    echo -e "${YELLOW}════════════════════════════════════════════════════════════${NC}\n"

    if [[ $TESTS_FAILED -gt 0 ]]; then
        exit 1
    fi
}

main "$@"
