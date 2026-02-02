# Universal Rename Detection System

## Overview

The Universal Rename Detection System is a sophisticated file and directory rename tracking solution that works across all supported protocols (Local, SMB, FTP, NFS, WebDAV) in the Catalogizer system. It efficiently detects and handles file/directory renames without triggering unnecessary rescans, maintaining data synchronization in real-time.

## Key Features

### ✅ Multi-Protocol Support
- **Local Filesystem**: Real-time detection using fsnotify
- **SMB**: Intelligent polling with batch operations
- **FTP**: Network-optimized detection with extended timeouts
- **NFS**: Efficient scanning with inode-aware tracking
- **WebDAV**: HTTP-based detection with ETag support

### ✅ Intelligent Detection
- **Move Window Optimization**: Protocol-specific timing windows (2s for Local, 30s for FTP)
- **Hash-Based Tracking**: Uses file content hashes for reliable identification
- **Directory Tree Handling**: Recursive rename detection for directory moves
- **Collision Prevention**: Prevents false positives and duplicate processing

### ✅ Performance Optimization
- **No Unnecessary Rescans**: Only metadata updates for rename operations
- **Batch Processing**: Efficient handling of bulk rename operations
- **Resource Management**: Configurable workers and queue sizes
- **Memory Efficient**: Automatic cleanup of expired tracking data

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                   Universal Rename Tracker                      │
├─────────────────────────────────────────────────────────────────┤
│  Protocol Handlers                                             │
│  ┌─────────┬─────────┬─────────┬─────────┬─────────────────────┐│
│  │ Local   │ SMB     │ FTP     │ NFS     │ WebDAV              ││
│  │ Handler │ Handler │ Handler │ Handler │ Handler             ││
│  └─────────┴─────────┴─────────┴─────────┴─────────────────────┘│
├─────────────────────────────────────────────────────────────────┤
│  Enhanced Change Watcher                                        │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │ • Real-time event processing                               ││
│  │ • Debounced change detection                               ││
│  │ • Protocol-aware file identification                      ││
│  │ • Concurrent worker processing                             ││
│  └─────────────────────────────────────────────────────────────┘│
├─────────────────────────────────────────────────────────────────┤
│  File Repository Extensions                                     │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │ • Efficient path updates                                   ││
│  │ • Batch directory moves                                    ││
│  │ • Hash-based file lookups                                 ││
│  │ • Metadata synchronization                                ││
│  └─────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────┘
```

## Protocol-Specific Behavior

### Local Filesystem
```go
Move Window: 2 seconds
Real-time Events: ✅ Yes (fsnotify)
Atomic Operations: ✅ Yes
Hash Calculation: ✅ Yes (< 100MB files)
```

### SMB Protocol
```go
Move Window: 10 seconds
Real-time Events: ❌ No (polling-based)
Atomic Operations: ❌ No (copy + delete)
Batch Size: 500 files
Network Optimization: ✅ Yes
```

### FTP Protocol
```go
Move Window: 30 seconds
Real-time Events: ❌ No (polling-based)
Atomic Operations: ❌ No (copy + delete)
Batch Size: 100 files
Retry Logic: ✅ Yes
```

### NFS Protocol
```go
Move Window: 5 seconds
Real-time Events: ❌ No (typically polling)
Atomic Operations: ✅ Partial (server dependent)
Batch Size: 800 files
Inode Support: ✅ Yes (when available)
```

### WebDAV Protocol
```go
Move Window: 15 seconds
Real-time Events: ❌ No (HTTP-based polling)
Atomic Operations: ❌ No (copy + delete)
ETag Support: ✅ Yes (when available)
Batch Size: 200 files
```

## Installation and Setup

### 1. Database Schema
The system automatically creates required tables:

```sql
-- Universal rename tracking
CREATE TABLE universal_rename_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    storage_root_id INTEGER NOT NULL,
    protocol TEXT NOT NULL,
    old_path TEXT NOT NULL,
    new_path TEXT NOT NULL,
    is_directory BOOLEAN NOT NULL,
    size INTEGER NOT NULL,
    file_hash TEXT,
    detected_at TIMESTAMP NOT NULL,
    processed_at TIMESTAMP,
    status TEXT NOT NULL DEFAULT 'pending'
);
```

### 2. Service Initialization
```go
// Initialize universal rename tracker
db := // your database connection
logger := zap.NewProduction()

renameTracker := services.NewUniversalRenameTracker(db, logger)
if err := renameTracker.Start(); err != nil {
    log.Fatal("Failed to start rename tracker:", err)
}
defer renameTracker.Stop()

// Initialize enhanced watcher
mediaDB := // your media database
analyzer := // your media analyzer

watcher := realtime.NewEnhancedChangeWatcher(mediaDB, analyzer, renameTracker, logger)
if err := watcher.Start(); err != nil {
    log.Fatal("Failed to start watcher:", err)
}
defer watcher.Stop()
```

## Usage Examples

### Basic File Rename Detection
```go
ctx := context.Background()

// Track file deletion (typically called by file system watcher)
renameTracker.TrackDelete(
    ctx,
    fileID,        // int64: Database file ID
    "/old/path.txt", // string: Original file path
    "storage1",    // string: Storage root name
    "smb",         // string: Protocol type
    1024,          // int64: File size
    &fileHash,     // *string: File hash (optional)
    false,         // bool: Is directory
    protocolData,  // map[string]interface{}: Protocol-specific data
)

// Detect file creation (typically called on new file detection)
pendingMove, isMove := renameTracker.DetectCreate(
    ctx,
    "/new/path.txt", // string: New file path
    "storage1",      // string: Storage root name
    "smb",           // string: Protocol type
    1024,            // int64: File size
    &fileHash,       // *string: File hash (optional)
    false,           // bool: Is directory
    protocolData,    // map[string]interface{}: Protocol-specific data
)

if isMove {
    // Process the detected move
    client := // your filesystem client
    err := renameTracker.ProcessMove(ctx, client, pendingMove, "/new/path.txt")
    if err != nil {
        log.Error("Failed to process move:", err)
    }
}
```

### Directory Rename Detection
```go
// Directory renames work similarly but with isDirectory = true
renameTracker.TrackDelete(ctx, dirID, "/old/directory", "storage1", "local", 0, nil, true, nil)

// When new directory is detected
pendingMove, isMove := renameTracker.DetectCreate(ctx, "/new/directory", "storage1", "local", 0, nil, true, nil)

if isMove {
    // Processes all child files and subdirectories automatically
    err := renameTracker.ProcessMove(ctx, client, pendingMove, "/new/directory")
}
```

### Batch Operations
```go
// For bulk rename operations, the system automatically handles batching
for i := 0; i < 100; i++ {
    oldPath := fmt.Sprintf("/batch/file_%d.txt", i)
    newPath := fmt.Sprintf("/renamed/file_%d.txt", i)

    // Track deletion
    renameTracker.TrackDelete(ctx, int64(i), oldPath, "storage1", "smb", 1024, nil, false, nil)

    // Detect and process move
    if pendingMove, isMove := renameTracker.DetectCreate(ctx, newPath, "storage1", "smb", 1024, nil, false, nil); isMove {
        renameTracker.ProcessMove(ctx, client, pendingMove, newPath)
    }
}
```

## Configuration

### Protocol-Specific Settings
```go
// Custom protocol handler registration
handlerFactory := services.NewProtocolHandlerFactory(logger)

// Register custom SMB handler with different settings
customSMBHandler := &CustomSMBHandler{
    moveWindow: 15 * time.Second, // Custom move window
    batchSize:  1000,             // Custom batch size
}
renameTracker.RegisterProtocolHandler("smb", customSMBHandler)
```

### Performance Tuning
```go
// Adjust worker count based on system resources
watcher := realtime.NewEnhancedChangeWatcher(mediaDB, analyzer, renameTracker, logger)
watcher.SetWorkers(8) // Increase for high-throughput systems

// Adjust queue size for high-volume environments
watcher.SetQueueSize(50000)

// Adjust debounce delay for faster detection
watcher.SetDebounceDelay(1 * time.Second)
```

## Monitoring and Statistics

### Real-time Statistics
```go
// Get comprehensive statistics
stats := renameTracker.GetStatistics()

fmt.Printf("Total pending moves: %d\n", stats["total_pending_moves"])
fmt.Printf("Success rate: %.2f%%\n", stats["success_rate"])

// Protocol-specific statistics
if pendingByProtocol, ok := stats["pending_by_protocol"].(map[string]int); ok {
    for protocol, count := range pendingByProtocol {
        fmt.Printf("Protocol %s: %d pending moves\n", protocol, count)
    }
}
```

### Enhanced Watcher Statistics
```go
since := time.Now().Add(-24 * time.Hour)
watcherStats, err := watcher.GetStatistics(since)
if err == nil {
    fmt.Printf("Changes in last 24h: %+v\n", watcherStats["changes_by_type"])
    fmt.Printf("Queue length: %d\n", watcherStats["queue_length"])
    fmt.Printf("Active workers: %d\n", watcherStats["workers"])
}
```

## Testing

### Unit Tests
```bash
# Run all rename detection unit tests
go test ./internal/services/... -v

# Run specific test patterns
go test ./internal/services -run TestRenameTracker -v
go test ./internal/media/realtime -run TestEnhancedWatcher -v
```

### Integration Tests
```bash
# Run protocol-specific integration tests
go test ./tests/integration -run TestProtocolRenameDetection -v

# Run with specific protocol (requires test environment)
SMB_TEST_SERVER=test.local go test ./tests/integration -run TestSMB -v
```

### Load Testing
```bash
# Run concurrent operation tests
go test ./tests/integration -run TestConcurrentOperations -v -count=5

# Performance benchmarks
go test ./internal/services -bench=BenchmarkRenameDetection -v
```

## Troubleshooting

### Common Issues

#### 1. Move Detection Not Working
```go
// Check protocol capabilities
capabilities, err := services.GetProtocolCapabilities("smb", logger)
if err != nil {
    log.Error("Failed to get capabilities:", err)
}
fmt.Printf("Move window: %v\n", capabilities.MoveWindow)
fmt.Printf("Real-time support: %v\n", capabilities.SupportsRealTimeNotification)
```

#### 2. False Positive Detection
```go
// Verify file identification
handler := services.NewSMBProtocolHandler(logger)
identifier, err := handler.GetFileIdentifier(ctx, path, size, isDir)
if err != nil {
    log.Error("File identification failed:", err)
}
```

#### 3. Performance Issues
```go
// Monitor queue lengths and worker utilization
stats := watcher.GetStatistics(time.Now().Add(-time.Hour))
queueLength := stats["queue_length"].(int)
if queueLength > 10000 {
    log.Warn("High queue length detected:", queueLength)
}
```

### Debug Logging
```go
// Enable debug logging for detailed tracking
logger := zap.NewDevelopment()
renameTracker := services.NewUniversalRenameTracker(db, logger)

// This will log all tracking operations, detections, and processing
```

### Health Checks
```go
// Implement health check endpoint
func renameTrackerHealthCheck() error {
    stats := renameTracker.GetStatistics()

    // Check for excessive pending moves
    if totalPending := stats["total_pending_moves"].(int); totalPending > 1000 {
        return fmt.Errorf("too many pending moves: %d", totalPending)
    }

    // Check success rate
    if successRate := stats["success_rate"].(float64); successRate < 90.0 {
        return fmt.Errorf("low success rate: %.2f%%", successRate)
    }

    return nil
}
```

## Best Practices

### 1. Protocol Selection
- **Use Local** for mounted network drives when possible (better performance)
- **Use SMB** for Windows networks with batch operations
- **Use NFS** for Unix environments requiring atomic operations
- **Use FTP** only when other protocols are unavailable
- **Use WebDAV** for cloud storage integration

### 2. Performance Optimization
- Configure appropriate move windows based on network latency
- Use batch operations for bulk rename scenarios
- Monitor queue lengths and adjust worker counts
- Implement circuit breakers for failing protocols

### 3. Error Handling
- Always handle move processing failures gracefully
- Implement retry logic for transient network errors
- Log all failed operations for debugging
- Use health checks to monitor system status

### 4. Security Considerations
- Validate all file paths to prevent directory traversal
- Implement proper authentication for network protocols
- Use encrypted connections (FTPS, HTTPS) when available
- Log all rename operations for audit trails

## API Reference

### UniversalRenameTracker
```go
type UniversalRenameTracker interface {
    Start() error
    Stop()
    TrackDelete(ctx context.Context, fileID int64, path, storageRoot, protocol string, size int64, fileHash *string, isDirectory bool, protocolData map[string]interface{})
    DetectCreate(ctx context.Context, newPath, storageRoot, protocol string, size int64, fileHash *string, isDirectory bool, protocolData map[string]interface{}) (*UniversalPendingMove, bool)
    ProcessMove(ctx context.Context, client filesystem.FileSystemClient, oldMove *UniversalPendingMove, newPath string) error
    GetStatistics() map[string]interface{}
}
```

### ProtocolHandler
```go
type ProtocolHandler interface {
    GetFileIdentifier(ctx context.Context, path string, size int64, isDir bool) (string, error)
    PerformMove(ctx context.Context, client filesystem.FileSystemClient, oldPath, newPath string, isDir bool) error
    ValidateMove(ctx context.Context, client filesystem.FileSystemClient, oldPath, newPath string) error
    GetMoveWindow() time.Duration
    SupportsRealTimeNotification() bool
}
```

### EnhancedChangeWatcher
```go
type EnhancedChangeWatcher interface {
    Start() error
    Stop()
    WatchPath(smbRoot, localMountPath string) error
    UnwatchPath(smbRoot string)
    GetStatistics(since time.Time) (map[string]interface{}, error)
}
```

## License

This rename detection system is part of the Catalogizer project and follows the same licensing terms.

## Contributing

When contributing to the rename detection system:

1. **Add tests** for any new protocol handlers
2. **Update documentation** for new features
3. **Follow naming conventions** for consistency
4. **Test with real protocols** when possible
5. **Measure performance impact** of changes

## Support

For issues related to rename detection:

1. Check the troubleshooting section above
2. Enable debug logging to identify the problem
3. Review protocol-specific capabilities
4. Test with simple rename operations first
5. Report issues with full debug logs and system information