# AGENTS.md - Assets Module

## Multi-Agent Coordination

When multiple agents work on this module:

1. **Package boundaries are hard contracts** — each package has a well-defined interface. Changes to interfaces require coordination.
2. **Tests are the source of truth** — run `go test ./... -race` before committing.
3. **No circular dependencies** — dependency flow is: asset <- store, resolver, event, defaults <- manager.

## Key Files

| File | Owner | Notes |
|------|-------|-------|
| `pkg/asset/asset.go` | Core | ID, Type, Status — do not change without updating all consumers |
| `pkg/store/store.go` | Storage | Store interface — implementations must be thread-safe |
| `pkg/resolver/resolver.go` | Resolution | Resolver interface — all implementations must handle context cancellation |
| `pkg/event/event.go` | Events | Event types — adding new types is safe, removing is breaking |
| `pkg/manager/manager.go` | Orchestration | Coordinates all packages — test thoroughly after changes |
