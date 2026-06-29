# §11.4.76 / Boot-provisioning — provision-when-DOWN PROVEN

**Revision:** 1 · **Last modified:** 2026-06-30T00:30:00Z
**Verdict: PROVEN** — the catalog-api binary provisions a down infra stack on the remote host via
the digital.vasic.containers submodule (remote compose-up over SSH), SAFELY proven against a
throwaway test stack (different ports/names) so the live infra was never touched.

## Real proof (§11.4.6, no bluff)
1. A test stack (postgres/redis/minio on ports 35432/36379/39000, `*-test` names) was ensured DOWN.
2. `infra.Provision()` (INFRA_PROVISION_ENABLED=true, test ports, INFRA_STAGE_COMPOSE=true) ran:
   - pre-flight: "infra endpoints needing provisioning: [redis minio postgres]" (down-case),
   - `mkdir -p catalogizer-infra-test` on thinker over SSH,
   - staged (scp) deploy/infra-compose-test.yml → thinker:catalogizer-infra-test/infra-compose-test.yml,
   - `podman-compose -f catalogizer-infra-test/infra-compose-test.yml up -d` over SSH,
   - Boot summary: **3 remote, 0 failed** → `PROVISION_RESULT: OK`.
3. VERIFIED on thinker.local: `catalogizer-postgres-test`, `catalogizer-redis-test`,
   `catalogizer-minio-test` all "Up". Then torn down; the LIVE *-dev stack stayed up + the API
   kept serving 27750 entities throughout (untouched).

## Fix found en route (§11.4.102)
First attempt's staging failed because the harness ran with cwd=catalog-api/, so a relative
INFRA_COMPOSE_FILE missed; using an absolute path resolved it. (Operational note: set
INFRA_COMPOSE_FILE to an absolute path, or run from the repo root.)

## Honest boundary
Proven against a test stack (safe). The live-infra teardown→reprovision (same data volumes
reattach) is the same code path; it was deliberately NOT run against the live serving stack to
avoid disruption. The provisioner's down-case is now end-to-end proven.
