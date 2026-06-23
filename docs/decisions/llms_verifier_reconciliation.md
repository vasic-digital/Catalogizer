# LLMsVerifier — Fork ↔ Upstream Reconciliation Decision

**Revision:** 1
**Last modified:** 2026-06-23T00:00:00Z
**Status:** OPEN — operator decision required (merge direction)
**Scope:** `submodules/llms_verifier` (was root `LLMsVerifier`) vs upstream `vasic-digital/LLMsVerifier`

---

## Summary

The catalogizer-embedded `LLMsVerifier` (now relocated to `submodules/llms_verifier`,
snake_case per §11.4.29) is a **catalogizer-specific fork** that has diverged from its
GitHub/GitLab upstream `vasic-digital/LLMsVerifier` in a **mutually incompatible** way.
A true git-submodule conversion to the current upstream is therefore **CONVERT-UNSAFE**
and was **not** performed (§11.4.122 no silent content removal, §11.4.6 no-guessing).
The safe relocate (content-preserving `git mv`, both modules build green) stands as the
current state.

## Evidence (captured 2026-06-23)

| Axis | Embedded (`submodules/llms_verifier`) | Upstream (`vasic-digital/LLMsVerifier`) |
|---|---|---|
| Go module path | `digital.vasic.llmsverifier` | `llmsverifier` (**different**) |
| `pkg/helixqa` (imported by helix_qa) | **present** | **MISSING** |
| `pkg/` tree | catalogizer-specific (42 tracked files) | divergent superset (~300+ files) |
| Embedded-unique | `.env.example`, `ARCHITECTURE.md`, `docs/ARCHITECTURE.md`, `docs/PROVIDER_USAGE_CONFIGURATION.md`, the whole `pkg/` | — |

`helix_qa/go.mod` replaces `digital.vasic.llmsverifier => ../llms_verifier` and imports
`digital.vasic.llmsverifier/pkg/helixqa`. The upstream provides **neither** that module
name **nor** `pkg/helixqa`, so pinning a submodule at the upstream HEAD would:
1. **Break `helix_qa`'s build** (missing package + wrong module path), and
2. **Delete catalogizer-specific code** (the embedded `pkg/` incl. `pkg/helixqa`).

## Options (operator decision — merge direction)

1. **Keep the fork as-is (current state).** `submodules/llms_verifier` stays a tracked
   in-repo directory at the correct snake_case path. helix_qa builds green. No upstream
   change. *(Lowest risk; the fork is not shared back.)*
2. **Promote the fork upstream.** Push the catalogizer fork (module
   `digital.vasic.llmsverifier` + `pkg/helixqa`) to upstream — e.g. to a clearly-named
   branch first (non-destructive), then reconcile with `main` deliberately. Then convert
   `submodules/llms_verifier` to a true submodule pinned at the reconciled commit.
   *(Affects the shared `vasic-digital/LLMsVerifier` repo — requires explicit intent so
   other consumers' work on the diverged upstream is not lost; §11.4.113 no force.)*
3. **Adopt upstream + port helix_qa.** Convert to the upstream submodule and port
   `helix_qa` (and any other consumer) off `pkg/helixqa` / the `digital.vasic.llmsverifier`
   module name onto upstream's API. *(Largest code change; only if the fork is obsolete.)*

## Recommendation

Default to **Option 1** until the operator decides the fork's fate. Options 2/3 mutate a
shared repository / break a consumer and must not be done silently (§11.4.66 / §11.4.122).

## Cross-references

§11.4.6 · §11.4.28 · §11.4.29 · §11.4.66 · §11.4.113 · §11.4.122
