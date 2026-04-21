# Verification Command Reference

> Copy the relevant chain into the `VERIFICATION COMMANDS` section of
> `templates/AI_TASK_ASSIGNMENT.md`. Every command must exit 0 before
> the task can be claimed done.
>
> Source: Master Plan v2 §8.1.

## Backend (catalog-api)

Full chain:

```bash
cd catalog-api
go fmt ./...
go vet ./...
golangci-lint run --config .golangci.yml ./...   # if present; skip if not installed
GOMAXPROCS=3 go test -short -count=1 -p 2 -parallel 2 ./...
GOMAXPROCS=3 go test -tags=integration -count=1 -p 2 -parallel 2 ./...
GOMAXPROCS=3 go test -race -count=1 -p 2 -parallel 2 ./...
```

Single test:

```bash
go test -v -run TestName ./path/to/pkg/
go test -v -run TestSuite/TestSubtest ./pkg/
```

## Frontend (catalog-web)

```bash
cd catalog-web
npm run lint           # ESLint --max-warnings 0
npm run type-check     # tsc --noEmit
npm test               # vitest single run
npm run test:e2e       # Playwright
npm run build          # tsc + vite
```

Single file:

```bash
npx vitest run path/to/file.test.ts
npx vitest run -t "test name pattern"
```

## Android Mobile (catalogizer-android)

```bash
cd catalogizer-android
./gradlew ktlintCheck
./gradlew lintDebug
./gradlew testDebugUnitTest
./gradlew connectedDebugAndroidTest   # requires device
./gradlew assembleRelease
```

Single test:

```bash
./gradlew :app:testDebugUnitTest --tests ClassName
./gradlew :app:testDebugUnitTest --tests ClassName.methodName
```

## Android TV (catalogizer-androidtv)

```bash
cd catalogizer-androidtv
./gradlew lintDebug
./gradlew testDebugUnitTest
./gradlew assembleDebug
# Release requires signing config at ../docker/signing/signing.properties
./gradlew assembleRelease
```

## Desktop (catalogizer-desktop, installer-wizard)

```bash
cd catalogizer-desktop   # or installer-wizard
npm run tauri:dev        # interactive smoke
npm run tauri:build
# Verify no bare unwrap() in Rust (see RULE-DESK-001)
! grep -rn "unwrap()" src-tauri/src/ --include="*.rs" | grep -v "// SAFE"
```

## API Client Library (catalogizer-api-client)

```bash
cd catalogizer-api-client
npm run build
npm test
```

## Submodule Checks (any digital.vasic.* / @vasic-digital/*)

```bash
cd Submodule
go build ./...                       # Go submodules
go test -count=1 -race ./...
go vet ./...

cd Submodule                         # TS submodules
npm run build
npm test
```

## Pre-Commit Chain

```bash
cd catalog-api && go fmt ./... && go vet ./...
cd catalog-web && npm run lint && npm run type-check
cd "$REPO_ROOT" && pre-commit run --all-files
```

Verify in a browser that there are zero console warnings / errors
(RULE-WEB-003).

Verify `.gitignore` covers `.env` in project root and every submodule:

```bash
git -C "$REPO_ROOT" ls-files --cached -- '*.env' ':!*.env.example'
# must output nothing
```

## Landmine Pre-Flight

```bash
scripts/detect-landmines.sh
```

If absent, use the inline checks from `docs/LANDMINES.md#appendix-quick-detection-script`.

## HelixQA Full Campaign

```bash
./scripts/helixqa-orchestrator.sh androidtv
./scripts/helixqa-orchestrator.sh android
./scripts/helixqa-orchestrator.sh web
./scripts/helixqa-orchestrator.sh desktop
./scripts/helixqa-orchestrator.sh all
```

Each run writes to `qa-results/session-<timestamp>/`. The orchestrator
exits non-zero on any FATAL BLOCKER / Session-failed / stagnation /
foreground-drift-unrecovered event.

## Full-QA Master Cycle (Article VII)

```bash
./scripts/services-down.sh            # clean slate
./scripts/services-up.sh              # all services
./scripts/run-all-tests.sh            # unit + integration + security
cd catalog-api && ./catalog-api &     # then RunAll via binary
curl -fsS -X POST -H "Authorization: Bearer $ADMIN_TOKEN" http://localhost:8080/api/v1/challenges/run-all
./scripts/helixqa-orchestrator.sh all
# review qa-results/session-*/videos + screenshots
# file tickets in docs/reports/qa-sessions/<date>/tickets/
# fix with 4 artefacts (unit test + fixes-validation + HelixQA bank + challenge)
# rebuild + repeat until clean pass
```

## Exit-Code Protocol

- **exit 0** → all checks pass, task may be claimed complete
- **exit non-zero** → task must remain open, iterate or escalate
- Never ignore a warning — RULE-CONST-005, RULE-WEB-003
- Never silence a failing test with `t.Skip()` — RULE-GO-006
