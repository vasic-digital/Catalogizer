# Diagram Sources

Mermaid (`.mmd`) source files for Catalogizer architecture diagrams. These are the canonical source of truth — ASCII diagrams in `docs/diagrams/*.md` are derived renders, not source.

## Files

| File | Topic | Render type |
|---|---|---|
| `architecture.mmd` | C4 container view — clients, api, db, monitoring, external storage/metadata | C4Context |
| `media-aggregation.mmd` | Post-scan media entity pipeline (title parsing → hierarchy → entity API) | flowchart |
| `auth-flow.mmd` | JWT login + refresh + protected request + logout | sequenceDiagram |
| `helixqa-pipeline.mmd` | Autonomous QA pipeline (Learn → Plan → Execute → Curiosity → Analyze) | flowchart |
| `database-dialect-rewriting.mmd` | database.DB wrapper SQL rewriting for PostgreSQL ↔ SQLite parity | flowchart |
| `tv-channels-flow.mmd` | Android TV home screen channels + Watch Next + deep linking | flowchart |

## Rendering

Rendering is done via `mermaid-cli` (mmdc) in a container to avoid host-level Node installs.

```bash
# One-shot render to SVG (rootless Podman)
podman run --rm -v "$PWD/docs/diagrams/sources:/data" \
  docker.io/minlag/mermaid-cli:latest \
  -i /data/architecture.mmd -o /data/architecture.svg -t dark

# Batch render all .mmd files
for src in docs/diagrams/sources/*.mmd; do
  name=$(basename "${src%.mmd}")
  podman run --rm -v "$PWD/docs/diagrams/sources:/data" \
    docker.io/minlag/mermaid-cli:latest \
    -i "/data/${name}.mmd" -o "/data/${name}.svg" -t dark
done
```

Generated `.svg` files are gitignored under `docs/diagrams/sources/*.svg` — rebuild them as part of the release pipeline, don't commit.

## Editing conventions

- Keep diagrams focused on one subsystem each — if a diagram has more than ~20 nodes, split it.
- Use `classDef` to color-code layers (primary concern, database, UI, external).
- Always include a header comment with the diagram's purpose and rendering notes.
- Reference the source files from `docs/architecture/*.md` narrative docs using relative paths.
- Update diagrams whenever the corresponding code changes — a stale diagram is worse than no diagram.
