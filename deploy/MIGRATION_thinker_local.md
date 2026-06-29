# Infra migration to thinker.local (2026-06-29)

**Operator mandate (2026-06-29):** move all infra (postgres/redis/minio) off the dev host
(nezha) onto `thinker.local` to free nezha resources; teardown nezha's local infra. Follow-up
mandate: the distribution MUST be executed by the main binary at boot via the
`digital.vasic.containers` submodule (§11.4.76) — see task #28 (the manual steps below are the
stopgap until that lands).

## What was done (manual stopgap, §9-safe)
1. **Backup (§9.2):** `pg_dump catalogizer_dev` → 7.2 MB gz (27 750 items + 3 217 covers).
2. **Stand up infra on thinker.local** (rootless podman 4.9.3, §11.4.161): `deploy/infra-compose.yml`
   → postgres 25432, redis 26379, minio 29000/29001. All healthy.
3. **Restore + verify:** dump loaded → thinker.local reports **27 750 items + 3 217 covers** (exact
   match, zero loss).
4. **Repoint API:** `DATABASE_HOST=thinker.local`, `REDIS_ADDR=thinker.local:26379`,
   `STORAGE_ENDPOINT=thinker.local:29000` (in the gitignored `run_catalog_api_durable.sh` launcher).
   API verified serving **27 750 entities** from thinker.local.
5. **Teardown nezha infra:** `podman-compose -f docker-compose.dev.yml down` — ZERO catalogizer
   containers remain on nezha; host memory freed (avail ~40 G).

## Why this also fixes the recurring postgres flapping
nezha was under resource pressure — its catalogizer-postgres container was repeatedly being
removed (PENDING_FORENSICS all session). Moving infra to thinker.local (26 G free RAM, 673 G NVMe)
removes the contention.

## Endpoints (thinker.local, reachable from nezha — verified)
| service | host:port | creds |
|---|---|---|
| postgres | thinker.local:25432 | catalogizer / dev_password_change_me / catalogizer_dev |
| redis | thinker.local:26379 | — |
| minio | thinker.local:29000 (console 29001) | minioadmin / minioadmin123 |

## NEXT (task #28 — the real requirement)
The main binary (catalog-api) MUST provision this distributed infra at boot via
`digital.vasic.containers` pkg/boot + WithDistributor + pkg/compose + pkg/health (§11.4.76
on-demand-infra invariant), replacing the manual steps above. The distribution target
(thinker.local) MUST be config-injected (§11.4.28), never hardcoded in the submodule.
