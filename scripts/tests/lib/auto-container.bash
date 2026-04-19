#!/usr/bin/env bash
# Tests for scripts/lib/auto-container.sh
# Sourced by scripts/tests/run-all.sh — uses pass/fail/assert_* helpers
# from the harness.
#
# Strategy: install a sandbox PATH with mock `podman` / `docker` /
# `cargo` binaries, then call the functions in a subshell to isolate
# exit codes + captured output.

# shellcheck shell=bash
# shellcheck disable=SC2317  # detect_runtime / is_container overrides are invoked indirectly via the sourced lib
# shellcheck disable=SC2016  # single-quoted strings here are scripts written into mock files — expansion happens at mock-run time, not now

_SUITE_TMP="$(mktemp -d -t auto-container-test.XXXXXX)"
trap 'rm -rf "$_SUITE_TMP"' EXIT

# Minimal stub for logging helpers pulled in via Build/lib/common.sh.
# The real lib prints colours; we just need silence in tests.
_stub_logs() {
    cat <<'EOF' > "$_SUITE_TMP/log_stubs.sh"
log_info()    { :; }
log_step()    { :; }
log_success() { :; }
log_warn()    { :; }
log_error()   { printf 'ERROR: %s\n' "$*" >&2; }
EOF
}
_stub_logs

# ─── mock builders ───────────────────────────────────────────────────

_mk_mock() {
    local name="$1" behaviour="$2"
    local path="$_SUITE_TMP/bin/$name"
    mkdir -p "$_SUITE_TMP/bin"
    cat > "$path" <<EOF
#!/usr/bin/env bash
$behaviour
EOF
    chmod +x "$path"
}

# Sources auto-container.sh inside a subshell with a custom PATH and
# given BUILD_PROJECT_ROOT. Returns exit code + captured stdout/stderr
# via eval-safe encoding.
_source_lib_in_subshell() {
    local cmd="$1"
    (
        export PATH="$_SUITE_TMP/bin:/usr/bin:/bin"
        export BUILD_PROJECT_ROOT="$_SUITE_TMP"
        # shellcheck disable=SC1091
        source "$_SUITE_TMP/log_stubs.sh"
        # Provide detect_runtime since auto-container.sh expects it from
        # Build/lib/common.sh.
        detect_runtime() { echo "podman"; }
        is_container() { return 1; }  # pretend host, not container
        # shellcheck disable=SC1091
        source "$BUILD_PROJECT_ROOT_REAL/scripts/lib/auto-container.sh"
        eval "$cmd"
    )
}

BUILD_PROJECT_ROOT_REAL="$ROOT"
export BUILD_PROJECT_ROOT_REAL

# ─── need_toolchain ──────────────────────────────────────────────────

{
    _mk_mock present_tool "exit 0"

    rc=0
    _source_lib_in_subshell 'need_toolchain present_tool' > /dev/null 2>&1 || rc=$?
    assert_rc "$rc" 0 "need_toolchain: returns 0 when tool is on PATH"

    rc=0
    _source_lib_in_subshell 'need_toolchain definitely_not_on_path_xyz' > /dev/null 2>&1 || rc=$?
    assert_rc "$rc" 1 "need_toolchain: returns 1 when tool is absent"
}

# ─── builder_image_exists ─────────────────────────────────────────────

{
    # podman image exists <tag>  → exit 0
    _mk_mock podman 'if [[ "$1" == "image" && "$2" == "exists" ]]; then exit 0; fi; exit 1'

    rc=0
    _source_lib_in_subshell 'builder_image_exists' > /dev/null 2>&1 || rc=$?
    assert_rc "$rc" 0 "builder_image_exists: 0 when podman reports present"

    # flip to missing
    _mk_mock podman 'if [[ "$1" == "image" && "$2" == "exists" ]]; then exit 1; fi; exit 1'
    rc=0
    _source_lib_in_subshell 'builder_image_exists' > /dev/null 2>&1 || rc=$?
    assert_rc "$rc" 1 "builder_image_exists: 1 when podman reports missing"
}

# ─── ensure_builder_image: short-circuits when image present ─────────

{
    _mk_mock podman 'if [[ "$1" == "image" && "$2" == "exists" ]]; then exit 0; fi; echo "podman build called"; exit 0'
    out=$(_source_lib_in_subshell 'ensure_builder_image' 2>&1) || true
    # Should NOT have invoked podman build
    if [[ "$out" == *"podman build called"* ]]; then
        fail "ensure_builder_image: should short-circuit when image present" "output=$out"
    else
        pass "ensure_builder_image: short-circuits when image present"
    fi
}

# ─── dispatch_if_missing: local path when tool present ───────────────

{
    _mk_mock cargo 'echo "local-cargo $*"; exit 0'
    out=$(_source_lib_in_subshell 'dispatch_if_missing cargo "cargo --version"' 2>&1) || true
    assert_contains "$out" "local-cargo --version" \
        "dispatch_if_missing: runs locally when tool is present"
}

# ─── dispatch_if_missing: disabled auto-dispatch ─────────────────────

{
    # remove cargo from PATH
    rm -f "$_SUITE_TMP/bin/cargo"
    rc=0
    out=$(CATALOGIZER_BUILDER_AUTO=0 _source_lib_in_subshell 'dispatch_if_missing cargo "cargo --version"' 2>&1) || rc=$?
    assert_rc "$rc" 1 "dispatch_if_missing: returns non-zero when tool missing + AUTO=0"
    assert_contains "$out" "CATALOGIZER_BUILDER_AUTO=0 but toolchain missing" \
        "dispatch_if_missing: surfaces clear error when AUTO disabled"
}

# ─── CATALOGIZER_BUILDER_IMAGE override ──────────────────────────────

{
    # confirm env override propagates into builder_image_exists
    _mk_mock podman 'echo "check=$3"; if [[ "$1" == "image" && "$2" == "exists" ]]; then exit 0; fi; exit 1'
    out=$(CATALOGIZER_BUILDER_IMAGE=localhost/custom-builder:test \
        _source_lib_in_subshell 'builder_image_exists' 2>&1) || true
    assert_contains "$out" "check=localhost/custom-builder:test" \
        "CATALOGIZER_BUILDER_IMAGE override: passed through to podman"
}
