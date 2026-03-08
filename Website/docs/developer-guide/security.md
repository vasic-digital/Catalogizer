# Security Guide

Catalogizer implements defense-in-depth security across authentication, data protection, network hardening, and vulnerability scanning.

---

## JWT Authentication Flow

Authentication uses a two-token system: a short-lived access token and a long-lived refresh token.

```
1. Client sends POST /api/v1/auth/login with username + password
2. Server validates credentials, returns access_token + refresh_token
3. Client includes access_token in Authorization: Bearer <token> header
4. Middleware validates token signature, expiry, and claims on each request
5. When access_token expires, client sends POST /api/v1/auth/refresh with refresh_token
6. Server issues a new access_token (and optionally rotates the refresh_token)
```

Tokens are signed with HMAC-SHA256 using the `JWT_SECRET` environment variable. Access tokens expire after 24 hours. Refresh tokens expire after 7 days.

The auth middleware in `internal/auth/` extracts the token, validates claims, and injects the authenticated user into the Gin context. Unauthenticated requests receive a `401 Unauthorized` response.

---

## Role-Based Access Control

Four roles control access to API endpoints:

| Role | Permissions |
|------|-------------|
| **Admin** | Full access. User management, configuration, challenges, storage root CRUD. |
| **Moderator** | Manage collections, edit metadata, trigger scans. Cannot manage users or system config. |
| **User** | Browse catalog, create personal collections, manage favorites, stream media. |
| **Viewer** | Read-only access. Browse and search only, no modifications. |

Role checks are enforced in middleware. Handlers annotated with role requirements reject requests from users with insufficient privileges, returning `403 Forbidden`.

---

## SQLCipher Database Encryption

SQLite databases can be encrypted at rest using SQLCipher (AES-256-CBC).

- Set the `DB_ENCRYPTION_KEY` environment variable to enable encryption.
- The key is applied via `PRAGMA key` immediately after opening the database connection.
- Encrypted databases are unreadable without the correct key.
- This applies to development/embedded deployments. PostgreSQL production deployments use PostgreSQL's own encryption features.

---

## Security Headers

The security middleware adds the following headers to all responses:

| Header | Value | Purpose |
|--------|-------|---------|
| `X-Frame-Options` | `DENY` | Prevents clickjacking by disallowing iframe embedding |
| `X-Content-Type-Options` | `nosniff` | Prevents MIME type sniffing |
| `Content-Security-Policy` | `default-src 'self'` | Restricts resource loading origins |
| `Strict-Transport-Security` | `max-age=31536000; includeSubDomains` | Enforces HTTPS for one year |
| `Referrer-Policy` | `strict-origin-when-cross-origin` | Limits referrer information leakage |
| `X-XSS-Protection` | `1; mode=block` | Enables browser XSS filter |

---

## CORS Configuration

Cross-Origin Resource Sharing is configured in middleware with these defaults:

- **Allowed Origins**: Configurable via `CORS_ORIGINS` env var. Defaults to `http://localhost:3000` in development.
- **Allowed Methods**: `GET`, `POST`, `PUT`, `DELETE`, `OPTIONS`, `PATCH`
- **Allowed Headers**: `Authorization`, `Content-Type`, `X-Requested-With`
- **Exposed Headers**: `X-Total-Count`, `X-Page`, `X-Per-Page`
- **Credentials**: Enabled (cookies and auth headers allowed cross-origin)
- **Max Age**: 12 hours (preflight cache duration)

---

## Rate Limiting

Rate limiting protects against brute-force and abuse:

| Scope | Limit | Endpoints |
|-------|-------|-----------|
| Per-IP | 100 requests/minute | All endpoints |
| Per-IP | 5 requests/minute | `/auth/login`, `/auth/register` |
| Per-User | 1000 requests/minute | Authenticated endpoints |
| Per-User | 10 requests/minute | `/scans/start`, `/challenges/run-all` |

Exceeded limits return `429 Too Many Requests` with a `Retry-After` header. The rate limiter uses the `RateLimiter/` submodule with sliding window counters backed by Redis (or in-memory fallback).

---

## Input Validation

All user input is validated before processing:

- **SQL Injection**: Parameterized queries exclusively. No string concatenation in SQL. The `database.DB` wrapper enforces placeholder-based queries.
- **XSS**: Output encoding on all user-supplied content. The frontend uses React's built-in escaping. API responses are JSON-only (no HTML rendering of user input).
- **Path Traversal**: File paths are canonicalized and validated against allowed storage roots. Requests containing `..` sequences in path parameters are rejected.
- **Request Size**: Maximum request body size enforced (default 10 MB). File uploads have separate configurable limits.

---

## Secrets Management

- Store secrets in environment variables or `.env` files, never in source code.
- The `.env` file is listed in `.gitignore` and must never be committed.
- `config.json` may hold non-sensitive defaults. Env vars always override `config.json`.
- Required secrets: `JWT_SECRET`, `ADMIN_PASSWORD`. Optional: `DB_ENCRYPTION_KEY`, `TMDB_API_KEY`, `OMDB_API_KEY`.
- In production containers, use Podman secrets or mount env files as volumes.

---

## Security Scanning

Run these tools regularly to detect vulnerabilities:

| Tool | Scope | Command |
|------|-------|---------|
| **govulncheck** | Go dependency vulnerabilities | `govulncheck ./...` |
| **npm audit** | Node.js dependency vulnerabilities | `cd catalog-web && npm audit` |
| **Trivy** | Container image scanning | `trivy image catalogizer-api:latest` |
| **Snyk** | Cross-language dependency analysis | `snyk test` |
| **SonarQube** | Static analysis and code smells | Via `docker-compose.security.yml` |

The project maintains a zero-vulnerability policy for production dependencies. The full security scan suite can be run with:

```bash
./scripts/run-all-tests.sh --security
```
