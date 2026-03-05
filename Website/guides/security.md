---
title: Security Guide
description: Security features and configuration for Catalogizer - authentication, encryption, TLS, and access control
---

# Security Guide

Catalogizer provides multiple layers of security for protecting your media catalog and user data. This guide covers authentication, access control, transport security, and operational security practices.

---

## JWT Authentication

Catalogizer uses JSON Web Token (JWT) authentication for API access.

### How It Works

1. The user sends credentials to `/api/v1/auth/login`
2. The server validates credentials and returns an access token and a refresh token
3. The client includes the access token in the `Authorization: Bearer <token>` header on every request
4. The auth middleware validates the token on each request
5. When the access token expires, the client uses the refresh token at `/api/v1/auth/refresh` to get a new one

### Token Configuration

| Setting | Default | Description |
|---------|---------|-------------|
| `JWT_SECRET` | (required) | Secret key for signing tokens |
| `JWT_EXPIRY` | `24h` | Access token lifetime |
| `JWT_REFRESH_EXPIRY` | `168h` | Refresh token lifetime (7 days) |

Set a strong, unique `JWT_SECRET` in production. If the secret changes, all existing tokens are invalidated and users must log in again.

---

## Role-Based Access Control

Catalogizer implements role-based access control (RBAC) to restrict actions by user role.

### Built-in Roles

| Role | Permissions |
|------|-------------|
| **admin** | Full access: user management, configuration, all media operations |
| **user** | Browse, search, play media, manage own favorites and collections |
| **viewer** | Read-only access to the catalog |

### Permission Enforcement

- Route-level middleware checks the user's role before processing the request
- Admin-only endpoints (user management, configuration, system scans) reject non-admin users with `403 Forbidden`
- Collection visibility (public, private, shared) is enforced at the query level
- Users can only modify their own favorites, playlists, and private collections

---

## Two-Factor Authentication

Administrators can enable two-factor authentication (2FA) for user accounts.

1. Enable 2FA from the user's profile or admin user management
2. Scan the QR code with an authenticator app (Google Authenticator, Authy, etc.)
3. Enter the verification code to confirm setup
4. Subsequent logins require both password and TOTP code

---

## Transport Security

### HTTP/3 (QUIC) with TLS

Catalogizer serves all traffic over HTTP/3 (QUIC) with TLS encryption. The server generates self-signed TLS certificates at startup for development. In production, use proper certificates from a certificate authority.

- HTTP/3 provides encrypted transport by default (TLS 1.3 is mandatory for QUIC)
- Fallback to HTTP/2 with TLS when HTTP/3 is unavailable
- HTTP/1.1 should not be used in production

### Brotli Compression

All HTTP responses are compressed with Brotli for reduced bandwidth usage. The compression middleware applies Brotli encoding when the client sends `Accept-Encoding: br`. Gzip is used as a fallback for clients that do not support Brotli.

### Reverse Proxy TLS

For production, terminate TLS at a reverse proxy (Nginx, Caddy, or Traefik). The project includes Nginx configuration files in `config/nginx.conf` with TLS settings.

---

## Database Encryption

SQLCipher provides AES-256 encryption for the SQLite database at rest.

- Set `DB_ENCRYPTION_KEY` to exactly 32 characters
- Without this key, the database file is unreadable
- Store the encryption key separately from database backups
- If using PostgreSQL, rely on PostgreSQL's native encryption and access control features

---

## CORS Configuration

Cross-Origin Resource Sharing is configured in the API middleware to control which origins can access the API.

- Development: Allows `localhost` origins by default
- Production: Configure allowed origins to match your frontend domain
- The CORS middleware sets appropriate `Access-Control-Allow-Origin`, `Access-Control-Allow-Methods`, and `Access-Control-Allow-Headers` headers

---

## Rate Limiting

The API includes rate limiting to prevent abuse and brute-force attacks.

- Authentication endpoints have stricter limits to prevent credential stuffing
- General API endpoints have configurable request-per-second limits
- Rate limit headers (`X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset`) are included in responses
- Exceeding the limit returns `429 Too Many Requests`

---

## Input Validation

All API inputs are validated by the input validation middleware before reaching handlers.

- Request body validation for JSON payloads
- Path parameter sanitization
- Query parameter type checking
- Protection against injection attacks and oversized payloads

---

## Security Testing

Catalogizer includes tools for security auditing.

```bash
# Go vulnerability check
govulncheck ./...

# Dependency vulnerability scanning
npm audit --production

# Security-focused test suite
./scripts/security-test.sh

# Snyk scanning
./scripts/snyk-scan.sh

# SonarQube analysis
./scripts/sonarqube-scan.sh
```

The `docker-compose.security.yml` file provides a containerized security scanning environment with pre-configured tools.

---

## Security Best Practices

- Set a strong `JWT_SECRET` (32+ characters, randomly generated)
- Change the default `ADMIN_PASSWORD` immediately after first login
- Enable 2FA for all admin accounts
- Use TLS in production (terminate at reverse proxy or use proper certificates)
- Set `DB_ENCRYPTION_KEY` for SQLite deployments
- Keep dependencies updated and run `govulncheck` and `npm audit` regularly
- Restrict network access to the API server using firewall rules
- Back up the encryption key separately from the database
