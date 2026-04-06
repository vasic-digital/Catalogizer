# Structured Logging Guide

**Date:** 2026-04-06  
**Version:** 2.2.0  
**Package:** `catalog-api/internal/logging`

---

## Overview

The Catalogizer API uses a centralized structured logging system built on top of Uber's `zap` library. This provides high-performance, structured logging with support for both development (human-readable) and production (JSON) formats.

## Quick Start

### Initialization

```go
import "catalogizer/internal/logging"

// Initialize at application startup
func main() {
    // Development mode (colorized console output)
    err := logging.Init("development")
    if err != nil {
        panic(err)
    }
    defer logging.Sync()
    
    // Or production mode (JSON output)
    err = logging.Init("production")
}
```

### Basic Usage

```go
// Simple messages
logging.Info("Server started")
logging.Debugf("Processing file %s", filename)

// With fields
logging.With(
    logging.String("user_id", userID),
    logging.Int("file_count", count),
).Info("Files processed")

// Error logging
logging.With(
    logging.ErrorField(err),
    logging.String("operation", "database_query"),
).Error("Operation failed")
```

## Log Levels

| Level | Method | Use Case |
|-------|--------|----------|
| DEBUG | `Debug()`, `Debugf()` | Development troubleshooting |
| INFO | `Info()`, `Infof()` | General operational information |
| WARN | `Warn()`, `Warnf()` | Warnings, recoverable issues |
| ERROR | `Error()`, `Errorf()` | Errors requiring attention |
| FATAL | `Fatal()`, `Fatalf()` | Critical errors (exits process) |

## Field Types

```go
// String fields
logging.String("key", "value")

// Numeric fields
logging.Int("count", 42)
logging.Int64("size", fileSize)
logging.Float64("ratio", 0.95)

// Boolean fields
logging.Bool("enabled", true)

// Error fields
logging.ErrorField(err)

// Any type
logging.Any("data", complexStruct)

// Duration
logging.Duration("elapsed", time.Since(start))
```

## Environment Modes

### Development Mode
- Colorized console output
- Human-readable format
- Caller information included
- Stack traces on errors

### Production Mode
- JSON format
- Structured for log aggregation
- Timestamps in ISO8601
- Efficient encoding

## Nil-Safe Design

All logging functions are nil-safe, meaning they won't panic if the logger is not initialized:

```go
// Safe to call even if Init() wasn't called
logging.Info("This won't panic")  // No-op if logger is nil

// Safe in unit tests
func TestSomething(t *testing.T) {
    // No need to initialize logger in tests
    logging.Debug("Test debug message")  // Won't panic
}
```

## Best Practices

### 1. Use Structured Fields

**Good:**
```go
logging.With(
    logging.String("user_id", user.ID),
    logging.String("action", "login"),
    logging.String("ip", clientIP),
).Info("User authenticated")
```

**Avoid:**
```go
logging.Infof("User %s logged in from %s", user.ID, clientIP)  // Less searchable
```

### 2. Log at Appropriate Levels

```go
// DEBUG: Detailed troubleshooting info
logging.Debugf("Processing chunk %d of %d", chunk, total)

// INFO: Normal operations
logging.Info("File scan completed")

// WARN: Recoverable issues
logging.Warn("Retrying failed request")

// ERROR: Actual errors
logging.With(logging.ErrorField(err)).Error("Database connection failed")
```

### 3. Include Context

```go
logger := logging.With(
    logging.String("request_id", requestID),
    logging.String("user_id", userID),
)

logger.Info("Request started")
// ... processing ...
logger.Info("Request completed")
```

### 4. Error Handling

```go
if err != nil {
    logging.With(
        logging.ErrorField(err),
        logging.String("operation", "file_read"),
        logging.String("path", filePath),
    ).Error("Failed to read file")
    return err
}
```

## Configuration

### Environment Variables

```bash
# Set logging mode
export APP_ENV=production  # or development

# Log level (when supported)
export LOG_LEVEL=info
```

### Programmatic Configuration

```go
config := zap.NewProductionConfig()
config.EncoderConfig.TimeKey = "timestamp"
config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

err := logging.InitWithConfig(config)
```

## Integration with Services

All new services should use the structured logging package:

```go
type MyService struct {
    db *database.DB
}

func (s *MyService) DoSomething(ctx context.Context) error {
    logging.With(logging.String("service", "MyService")).Debug("Starting operation")
    
    // ... do work ...
    
    if err != nil {
        logging.With(logging.ErrorField(err)).Error("Operation failed")
        return err
    }
    
    logging.Info("Operation completed successfully")
    return nil
}
```

## Migration from Print Statements

Old code using `fmt.Printf` or `log.Printf` should be migrated:

```go
// Before
log.Printf("Processing user %d with %d files", userID, fileCount)

// After
logging.With(
    logging.Int("user_id", userID),
    logging.Int("file_count", fileCount),
).Info("Processing user files")
```

## Testing

The logging package includes comprehensive tests:

```bash
cd catalog-api
go test ./internal/logging/... -v
```

## Performance

- Zero-allocation logging for simple messages
- Efficient JSON encoding in production
- Async log writing to prevent blocking
- Benchmarks available in `logger_test.go`

## Troubleshooting

### Logs not appearing
- Verify `logging.Init()` was called
- Check log level (DEBUG logs won't show in production)

### Panics in tests
- Logging is nil-safe; panics likely from other causes
- Ensure `logging.Logger != nil` check if needed

### Performance issues
- Avoid logging in hot loops
- Use `Debug()` level for verbose output
- Consider sampling for high-volume logs

## References

- [Zap Documentation](https://pkg.go.dev/go.uber.org/zap)
- [Structured Logging Best Practices](https://github.com/uber-go/zap/blob/master/FAQ.md)
