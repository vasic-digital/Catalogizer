# §11.4.173 Containerized + Distributed Build — PROOF (first increment)

**Revision:** 1
**Last modified:** 2026-06-30T00:20:00Z
**Mandate:** §11.4.173 — every build runs inside a container distributed to the remote build host
(thinker.local), artifact brought back; never on the bare host.

## Verdict: PROVEN — catalog-api built in a container ON thinker.local, artifact verified back here.

`scripts/build_in_container.sh` synced catalog-api + its 15 replaced submodules to thinker.local,
built catalog-api INSIDE a golang container there (CGO on for go-sqlite3 + the docx CGO dep), and
copied the binary BACK to deploy/artifacts/. Rock-solid proof (§11.4.6, no bluff):

```
=== §11.4.173 BUILD PROOF ===
  build_host:   thinker.local (in container docker.io/library/golang:1.25-bookworm)
  remote_md5:   4c6a267de143ea71fd1edb6cc35d108b  (size 124934288)
  local_md5:    4c6a267de143ea71fd1edb6cc35d108b
  md5_match:    YES
  artifact:     ELF 64-bit LSB executable, x86-64, GNU/Linux, dynamically linked, not stripped
  [build] OK — built on thinker.local in a container, artifact verified on this host.
```

The brought-back `deploy/artifacts/catalog-api-container` is a valid Linux x86-64 ELF Go binary;
its md5 EXACTLY matches the artifact produced inside the remote container (identity preserved on
copy-back).

## Root cause fixed en route (§11.4.102)
First run FAILED on the alpine image: `undefined reference to __snprintf_chk` — a CGO dependency
(docx.c) references glibc `_FORTIFY_SOURCE` symbols that musl libc (alpine) lacks. Fix: build in
`golang:1.25-bookworm` (debian/glibc, gcc preinstalled). The alpine-vs-glibc CGO lesson is now
encoded in the script + this doc.

## Honest scope (§11.4.6) — first increment only
PROVEN: the catalog-api (Go) build is containerized + distributed + artifact-returned. REMAINING:
- Android (Gradle) builds — the APK builds still run on the host (the §11.4.173 operational gap).
- Desktop + web builds — containerize similarly.
- A first-class build primitive in the digital.vasic.containers submodule (§11.4.74 extend) instead
  of the ad-hoc podman wrapper this script uses.
