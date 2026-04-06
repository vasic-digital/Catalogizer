# Quick Reference - Catalogizer Development

## Essential Commands

### Backend (Go)
```bash
cd catalog-api

# Build
go build -o catalog-api .

# Run with hot reload (development)
go run main.go

# Run tests (resource-limited)
GOMAXPROCS=3 go test ./... -p 2 -parallel 2

# Run specific test
go test -v -run TestName ./path/to/pkg/

# Format & lint
go fmt ./... && go vet ./...

# Security scan
gosec ./...
```

### Frontend (React/TypeScript)
```bash
cd catalog-web

# Development server
npm run dev

# Build
npm run build

# Tests
npm run test
npm run test:watch
npm run test:coverage

# Lint & type check
npm run lint
npm run type-check
npm run lint:fix
```

### Android
```bash
cd catalogizer-android

# Build debug APK
./build-fixed.sh

# Or manually
./gradlew :app:assembleDebug

# Run tests
./gradlew test
```

### Security Scanning
```bash
# Full security scan
./scripts/security-scan-full.sh

# Individual tools
cd catalog-api
gosec ./...
trivy filesystem .
go list -json -deps ./... | nancy sleuth
semgrep --config=auto .
```

### Monitoring
```bash
# Start all services
podman-compose -f docker-compose.yml up

# Start monitoring only
podman-compose -f docker-compose.monitoring.yml up

# View logs
podman-compose logs -f catalog-api
```

## Structured Logging

### Basic Usage
```go
import "catalogizer/internal/logging"

// Initialize
logging.Init("development") // or "production"
defer logging.Sync()

// Simple logging
logging.Info("Server started")
logging.Debugf("Processing %s", filename)

// With fields
logging.With(
    logging.String("user_id", userID),
    logging.Int("count", count),
).Info("Operation completed")

// Error logging
logging.With(
    logging.ErrorField(err),
).Error("Operation failed")
```

### Log Levels
- `Debug()` - Development troubleshooting
- `Info()` - Normal operations  
- `Warn()` - Warnings, recoverable issues
- `Error()` - Errors requiring attention
- `Fatal()` - Critical errors (exits)

## Database Migrations

### Migration Files
- Located in: `catalog-api/database/`
- Naming: `migrations_v{N}_{description}.go`
- Register in: `migrations.go`

### Running Migrations
Migrations run automatically on startup:
```bash
./catalog-api
```

### Check Migration Status
```sql
SELECT * FROM migrations ORDER BY version;
```

## Git Workflow

### Pre-Commit Checklist
```bash
./scripts/pre-flight-check.sh
```

### Commit Message Format
```
type(scope): description

[optional body]

[optional footer]
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`

### SSH-Only Git (MANDATORY)
```bash
# Clone with SSH
git clone git@github.com:vasic-digital/Catalogizer.git

# Never use HTTPS
git clone https://github.com/...  # ❌ FORBIDDEN
```

## Project Structure

```
Catalogizer/
├── catalog-api/          # Go backend
│   ├── internal/         # Internal packages
│   ├── handlers/         # HTTP handlers
│   ├── services/         # Business logic
│   ├── repository/       # Data access
│   └── database/         # Migrations
├── catalog-web/          # React frontend
│   ├── src/
│   │   ├── components/   # UI components
│   │   ├── pages/        # Page components
│   │   ├── lib/          # Utilities
│   │   └── hooks/        # React hooks
├── catalogizer-android/  # Android app
├── catalogizer-androidtv/# Android TV app
├── docs/                 # Documentation
├── monitoring/           # Monitoring configs
└── scripts/              # Build & utility scripts
```

## Common Issues

### Backend

**Issue:** Port already in use  
**Fix:** Kill process on port 8080 or use `findAvailablePort`

**Issue:** Database locked  
**Fix:** Only one process can access SQLite at a time

**Issue:** Migration fails  
**Fix:** Check `migrations` table, may need manual rollback

### Frontend

**Issue:** Node modules out of date  
**Fix:** `rm -rf node_modules && npm install`

**Issue:** Type errors after API changes  
**Fix:** `npm run type-check` and update types

### Android

**Issue:** JDK 21 compatibility  
**Fix:** Use `./build-fixed.sh` instead of direct gradle

**Issue:** Build fails with jlink error  
**Fix:** Workarounds applied in `gradle.properties`

## Environment Variables

### Backend (.env)
```bash
PORT=8080
HOST=0.0.0.0
JWT_SECRET=your-secret
DATABASE_TYPE=sqlite
DB_PATH=./data/catalogizer.db
ENABLE_AUTH=true
```

### Frontend (.env.local)
```bash
VITE_API_URL=http://localhost:8080
VITE_WS_URL=ws://localhost:8080/ws
```

## Testing

### Run All Tests
```bash
# Backend
cd catalog-api
GOMAXPROCS=3 go test ./... -p 2 -parallel 2

# Frontend
cd catalog-web
npm run test

# E2E
cd catalog-web
npm run test:e2e
```

### Coverage Reports
```bash
# Go
go test -cover ./...

# Frontend
npm run test:coverage
```

## Resources

### Documentation
- `docs/guides/STRUCTURED_LOGGING.md`
- `SECURITY_AUDIT_REPORT.md`
- `PERFORMANCE_OPTIMIZATION_REPORT.md`
- `docs/DEVELOPER_GUIDE.md`

### Configuration
- `CLAUDE.md` - AI assistant guide
- `AGENTS.md` - Development guide
- `.pre-commit-config.yaml` - Git hooks

### Monitoring
- Prometheus: http://localhost:9090
- Grafana: http://localhost:3001
- AlertManager: http://localhost:9093

## Support

### Debugging
- Backend logs: `journalctl -u catalogizer -f`
- Frontend: Browser DevTools
- Android: `adb logcat | grep catalogizer`

### Emergency Contacts
- See `docs/SECURITY_INCIDENT_RESPONSE.md`
- See `docs/DISASTER_RECOVERY.md`

---

*Quick Reference - v2.2.0*
