# Build Container Auto-Dispatch

**Status:** active (2026-04-19)
**Source:** `scripts/lib/auto-container.sh`
**Tests:** `scripts/tests/lib/auto-container.bash` (9/9 passing)

## Problem

The host machine often does not have every toolchain the build needs:

- `catalogizer-desktop` + `installer-wizard` require **Rust** (`cargo`, `rustc`)
- `catalogizer-android[tv]` requires **JDK 21 + Android SDK**
- Some components require **Node 20**, **Go 1.25**, or C headers for `libvlc`

Per the Constitution (Article VII §7.1), no build may `sudo`, become
`root`, or install system packages on the host. Forcing every operator
to bootstrap the toolchain themselves is both disallowed and wasteful.

## Solution

When a build step needs a toolchain that is not on the host `PATH`, the
build transparently re-executes itself inside the rootless
`catalogizer-builder` container.

```
 host (operator, no Rust) ──┐
                            │
 scripts/release-build.sh   │
  └─ scripts/lib/           │
      build-desktop.sh      │
       └─ command -v cargo? ├── yes ─▶ run locally
                            └── no  ─▶ podman run --rm --entrypoint=""
                                              --network host
                                              --userns=keep-id
                                              -v $PWD:/workspace:z
                                              localhost/catalogizer-builder:latest
                                              bash -c "build_desktop …"
```

The container image is **built locally once** on first use
(~8–15 min) and cached. Every subsequent dispatch is instant.

## Public API

### `need_toolchain <name>`

Returns 0 if `name` is in `PATH`, 1 otherwise.

```bash
if need_toolchain cargo; then
    cargo build
else
    run_in_builder "cargo build"
fi
```

### `builder_image_exists`

Returns 0 if `$CATALOGIZER_BUILDER_IMAGE` (default
`localhost/catalogizer-builder:latest`) is already in the local image
store.

### `ensure_builder_image`

Short-circuits if the image exists. Otherwise runs
`podman build -f docker/Dockerfile.builder`. Respects
`$CATALOGIZER_BUILDER_IMAGE` for custom tags.

### `run_in_builder "<shell command>"`

Runs the given command inside the builder container with the project
mounted at `/workspace`. Automatically calls `ensure_builder_image`
first.

```bash
run_in_builder "cd /workspace/catalogizer-desktop && npm run tauri:build -- --bundles deb,rpm"
```

Flags used:

| Flag | Purpose |
|---|---|
| `--rm` | Remove container on exit. Prevents name collisions + disk leak. |
| `--network host` | SSL fetches from crates.io, dl.google.com, etc. |
| `--userns=keep-id` | Preserve the host user's UID so bind-mount writes land with the right ownership. |
| `--entrypoint=""` | Override `ENTRYPOINT` baked into the builder image so we can run an ad-hoc `bash -c` command. |
| `-v $BUILD_PROJECT_ROOT:/workspace:z` | Project bind-mount with SELinux relabel. |
| `-e APPIMAGE_EXTRACT_AND_RUN=1` | Tauri + AppImage bundling works without FUSE in the container. |
| `-e GOTOOLCHAIN=local` | Prevent Go from auto-downloading a newer toolchain. |

### `dispatch_if_missing <tool> "<shell command>"`

One-line convenience: runs `command` locally if `tool` is present, else
dispatches into the builder.

## Configuration

Environment variables:

| Variable | Default | Purpose |
|---|---|---|
| `CATALOGIZER_BUILDER_IMAGE` | `localhost/catalogizer-builder:latest` | Image tag to run. |
| `CATALOGIZER_BUILDER_AUTO` | `1` | Set to `0` to refuse auto-dispatch (fails loudly if the toolchain is missing — useful for CI lanes that want explicit handling). |

## How the builder image is wired

`docker/Dockerfile.builder` layers:

1. Base `ubuntu:24.04`
2. System packages (build-essential, curl, git, pkg-config, libvlc-dev, …)
3. Go 1.25 to `/usr/local/go`
4. Node 20 + npm to `/usr/local`
5. **Rust stable to `/opt/rustup` + `/opt/cargo`** (world-readable so
   `--userns=keep-id` dispatches work — see rebuild note below)
6. JDK 21 + Android SDK to `/opt/android-sdk`

`ENTRYPOINT` is `/project/scripts/build-test-release.sh` (the CI entry
point). Ad-hoc `run_in_builder` calls override it with
`--entrypoint=""`.

### Why `/opt/cargo` and not `/root/.cargo`

`rustup` defaults to `$HOME/.cargo` with `0700` permissions. Under
rootless podman with `--userns=keep-id`, the host user's UID is
preserved inside the container, so `/root/.cargo` (owned by container
root) is **unreadable** to the dispatched user. Moving `CARGO_HOME` to
`/opt/cargo` with `chmod -R a+rX` makes the toolchain usable from any
UID mapping.

If you are bumping the Rust toolchain, remember to also:

- Rebuild the builder image: `podman build --network host -f docker/Dockerfile.builder -t localhost/catalogizer-builder:latest .`
- Nuke the `cargo-cache` volume so the new toolchain's registry is
  rebuilt fresh: `podman volume rm catalogizer_cargo-cache`

## Tests

Pure-bash harness under `scripts/tests/`:

```
scripts/tests/run-all.sh             # run every test file
scripts/tests/lib/auto-container.bash  # 9 tests covering need_toolchain,
                                       # builder_image_exists, ensure_builder_image
                                       # short-circuit, dispatch_if_missing
                                       # (local + AUTO=0), image override
```

Run:

```bash
./scripts/tests/run-all.sh
```

Expected:

```
──  passed: 9   failed: 0
ALL TESTS PASSED
```

No `bats` dependency — tests use only `bash` + `podman` mocks.

## End-to-end verification

To prove the dispatch works end-to-end without running a full Tauri
build (~30 min), run a trivial command:

```bash
bash -c '
set -uo pipefail
export BUILD_PROJECT_ROOT="$PWD"
source Build/lib/common.sh
source scripts/lib/auto-container.sh
run_in_builder "cargo --version && rustc --version && node --version && npm --version"
'
```

Expected output:

```
cargo 1.xx.0 (…)
rustc 1.xx.0 (…)
v20.xx.0
10.xx.x
```

If that succeeds, a subsequent `./scripts/release-build.sh --local` of
`catalogizer-desktop` / `installer-wizard` will transparently route
through the container because those components call `dispatch_if_missing
cargo …` internally.

## Caveats / known limitations

1. **First-use cost** — building the image takes 8–15 min and requires
   ~4 GB of network bandwidth (crates.io, Android SDK mirror, JDK).
2. **Host-only AppImage** — `.AppImage` bundling requires FUSE and a
   working `xdg-open`. Containers produce `.deb` + `.rpm` only; AppImage
   must be built on the host (either via direct Tauri or by bundling
   from inside the AppImage later).
3. **No GPU access** — `run_in_builder` intentionally does not pass
   `--gpus=all`. GPU-accelerated components (OCU CUDA sidecar) use a
   different runtime path.
4. **Podman or Docker** — `detect_runtime` prefers `podman`, falls back
   to `docker`. Both are supported.
