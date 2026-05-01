# Catalogizer → HelixCode QA Integration

This example demonstrates how **Catalogizer** (or any external Go application) can orchestrate QA sessions through a **HelixCode** server's REST API.

## What it demonstrates

1. **Start Session** — `POST /api/v1/qa/session` with platforms, banks, and coverage targets
2. **Poll Status** — `GET /api/v1/qa/session/{id}/status` for progress updates
3. **Cancel Session** — `DELETE /api/v1/qa/session/{id}` for cleanup
4. **Authentication** — Bearer token JWT auth against HelixCode

## Prerequisites

- HelixCode server running with QA enabled (`qa.enabled = true`)
- Valid JWT token from HelixCode authentication

## Usage

```bash
export HELIXCODE_URL=http://localhost:8080
export HELIXCODE_TOKEN=<your-jwt-token>

go run ./examples/helixcode-qa-integration
```

## Architecture

```
┌─────────────┐      HTTP/REST      ┌─────────────┐
│ Catalogizer │  ─────────────────► │  HelixCode  │
│  (client)   │   Bearer JWT Auth   │  (server)   │
│             │ ◄─────────────────  │  QA Engine  │
└─────────────┘   Session State     └─────────────┘
```

## Production considerations

- Use SSE (`Accept: text/event-stream`) instead of polling for live progress
- Store session IDs in your application's database for tracking
- Retrieve reports asynchronously once sessions complete
- Handle token refresh for long-running sessions
