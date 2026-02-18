# CLAUDE.md - Assets Module

## Overview

`digital.vasic.assets` is a generic, reusable Go module for lazy asset loading. It provides strategy-based resolution, filesystem storage, event-driven lifecycle notifications, and embedded default placeholders.

**Module**: `digital.vasic.assets` (Go 1.24+)
**Dependencies**: `google/uuid`, `stretchr/testify` (tests only)

## Build & Test

```bash
go build ./...
go test ./... -count=1 -race
go test ./... -short
```

## Code Style

- Standard Go conventions, `gofmt` formatting
- Imports grouped: stdlib, third-party, internal (blank line separated)
- Line length <= 100 chars
- Naming: `camelCase` private, `PascalCase` exported
- Errors: always check, wrap with `fmt.Errorf("...: %w", err)`
- Tests: table-driven where appropriate, `testify` assertions

## Package Structure

| Package | Purpose |
|---------|---------|
| `pkg/asset` | Core types: Asset, ID, Type, Status |
| `pkg/store` | Store interface + FileStore (filesystem) + MemoryStore (tests) |
| `pkg/resolver` | Resolver interface + ChainResolver + HTTPResolver + LocalFileResolver |
| `pkg/event` | EventBus interface + InMemoryBus |
| `pkg/defaults` | Provider interface + EmbeddedProvider with `//go:embed` placeholders |
| `pkg/manager` | Manager orchestrator + worker pool + functional options |

## Key Interfaces

- `store.Store` — Get/Put/Delete/Exists for asset content bytes
- `resolver.Resolver` — Name/CanResolve/Resolve/Priority for strategy-based resolution
- `event.EventBus` — Publish/Subscribe for asset lifecycle events
- `defaults.Provider` — GetDefault/Register for fallback content

## Design Patterns

- **Strategy**: Resolver implementations (HTTP, local, chain)
- **Chain of Responsibility**: ChainResolver tries resolvers in priority order
- **Observer**: EventBus for asset lifecycle notifications
- **Functional Options**: Manager configuration (WithStore, WithResolver, etc.)
- **Worker Pool**: Background resolution goroutines

## Commit Style

Conventional Commits: `feat(resolver): add HTTP resolver with timeout support`
