# Module 7: Protocol Implementations - Script

**Duration**: 60 minutes
**Module**: 7 - Protocol Implementations

---

## Scene 1: SMB/CIFS Client (0:00 - 15:00)

**[Visual: Architecture diagram showing FileSystemClient interface with SMB, FTP, NFS, WebDAV, Local implementations]**

**Narrator**: Welcome to Module 7. Catalogizer supports five filesystem protocols, all accessed through a single unified interface. This module explores each protocol implementation, the resilience patterns for unreliable networks, and the factory pattern that ties everything together.

**[Visual: Open `catalog-api/filesystem/smb_client.go`]**

**Narrator**: The SMB client handles Windows file sharing, the most common protocol for NAS devices. It wraps the Go SMB library to implement the `FileSystemClient` interface, handling connections, authentication, share mounting, and file operations.

**[Visual: Show SMB configuration]**

**Narrator**: SMB connections require host, port (default 445), share name, username, password, and domain. The domain defaults to "WORKGROUP" for simple home networks. Each connection negotiates the SMB dialect automatically.

```go
// catalog-api/filesystem/factory.go
case "smb":
    smbConfig := &SmbConfig{
        Host:     getStringSetting(config.Settings, "host", ""),
        Port:     getIntSetting(config.Settings, "port", 445),
        Share:    getStringSetting(config.Settings, "share", ""),
        Username: getStringSetting(config.Settings, "username", ""),
        Password: getStringSetting(config.Settings, "password", ""),
        Domain:   getStringSetting(config.Settings, "domain", "WORKGROUP"),
    }
    return NewSmbClient(smbConfig), nil
```

**[Visual: Open `catalog-api/internal/smb/resilience.go`]**

**Narrator**: SMB connections over home networks are inherently unreliable. NAS devices go to sleep, WiFi drops, and connections time out. The `ResilientSMBManager` handles all of this with circuit breakers, offline caching, and exponential backoff retry.

```go
// catalog-api/internal/smb/resilience.go
type ResilientSMBManager struct {
    sources       map[string]*SMBSource
    logger        *zap.Logger
    offlineCache  *OfflineCache
    healthChecker *HealthChecker
    eventChannel  chan SMBEvent
    stopChannel   chan struct{}
    wg            sync.WaitGroup
}
```

**[Visual: Show `ConnectionState` enum]**

**Narrator**: Each SMB source tracks its connection state: Connected, Disconnected, Reconnecting, or Offline. State transitions emit events that propagate to the WebSocket handler, so the frontend can display connection health in real time.

```go
// catalog-api/internal/smb/resilience.go
type ConnectionState int

const (
    StateConnected ConnectionState = iota
    StateDisconnected
    StateReconnecting
    StateOffline
)
```

**[Visual: Show SMBSource struct with retry configuration]**

**Narrator**: Each `SMBSource` carries its own retry configuration: maximum retry attempts, retry delay with exponential backoff, connection timeout, and health check interval. When a connection fails, the manager retries with increasing delays. After exhausting retries, the source is marked Offline and serves from its offline cache.

```go
// catalog-api/internal/smb/resilience.go
type SMBSource struct {
    ID                  string
    Name                string
    Path                string
    State               ConnectionState
    RetryAttempts       int
    MaxRetryAttempts    int
    RetryDelay          time.Duration
    ConnectionTimeout   time.Duration
    HealthCheckInterval time.Duration
    IsEnabled           bool
}
```

**[Visual: Show health metrics integration]**

**Narrator**: Connection state maps to Prometheus metrics. Connected is healthy, Reconnecting is degraded, and Disconnected/Offline is marked as offline. The metrics endpoint exposes these values for monitoring and alerting.

---

## Scene 2: WebDAV Client (15:00 - 30:00)

**[Visual: Open `catalog-api/filesystem/webdav_client.go`]**

**Narrator**: WebDAV extends HTTP with file management verbs. The WebDAV client translates `FileSystemClient` interface calls into HTTP methods: PROPFIND for listing, GET for reading, PUT for writing, DELETE for removing, and MKCOL for creating directories.

**[Visual: Show WebDAV configuration]**

**Narrator**: WebDAV requires a URL (the server root), optional username and password for HTTP Basic authentication, and a base path within the server.

```go
// catalog-api/filesystem/factory.go
case "webdav":
    webdavConfig := &WebDAVConfig{
        URL:      getStringSetting(config.Settings, "url", ""),
        Username: getStringSetting(config.Settings, "username", ""),
        Password: getStringSetting(config.Settings, "password", ""),
        Path:     getStringSetting(config.Settings, "path", ""),
    }
    return NewWebDAVClient(webdavConfig), nil
```

**[Visual: Show PROPFIND XML parsing]**

**Narrator**: PROPFIND responses are XML. The client parses multistatus responses to extract file names, sizes, modification times, content types, and resource types. This XML parsing is the most complex part of the WebDAV implementation.

**[Visual: Show batch operations]**

**Narrator**: WebDAV supports batch operations through COPY and MOVE verbs at the HTTP level. The client exposes these as single-call operations, unlike SMB where multi-file moves require iterative per-file calls.

**[Visual: Open `catalog-api/services/webdav_client.go`]**

**Narrator**: There is also a domain-level WebDAV client in the `services/` package. This wraps the protocol-level client with domain logic: path resolution relative to storage roots, metadata extraction during listing, and error translation to domain errors.

---

## Scene 3: FTP Client (30:00 - 40:00)

**[Visual: Open `catalog-api/filesystem/ftp_client.go`]**

**Narrator**: FTP is the oldest protocol Catalogizer supports. The FTP client handles both active and passive mode connections, binary and ASCII transfer types, and standard directory listing.

**[Visual: Show FTP configuration]**

**Narrator**: FTP requires host, port (default 21), username, password, and a base path.

```go
// catalog-api/filesystem/factory.go
case "ftp":
    ftpConfig := &FTPConfig{
        Host:     getStringSetting(config.Settings, "host", ""),
        Port:     getIntSetting(config.Settings, "port", 21),
        Username: getStringSetting(config.Settings, "username", ""),
        Password: getStringSetting(config.Settings, "password", ""),
        Path:     getStringSetting(config.Settings, "path", ""),
    }
    return NewFTPClient(ftpConfig), nil
```

**[Visual: Show passive mode handling]**

**Narrator**: Passive mode is the default for FTP clients behind NAT. In passive mode, the server opens a data port and the client connects to it, avoiding firewall issues. The FTP client uses passive mode by default and falls back to active mode only when configured.

**[Visual: Show binary vs ASCII transfer]**

**Narrator**: Binary transfer mode preserves file contents exactly. ASCII mode translates line endings between platforms. Catalogizer always uses binary mode for media files to prevent corruption. The transfer type is set before each file operation.

**[Visual: Show directory listing parsing]**

**Narrator**: FTP directory listings use the LIST command, which returns platform-dependent text output. The client parses both Unix-style (`-rwxr-xr-x`) and Windows-style listings to extract filenames, sizes, and dates.

---

## Scene 4: NFS and Local Filesystems (40:00 - 50:00)

**[Visual: Open `catalog-api/filesystem/nfs_client.go`]**

**Narrator**: NFS (Network File System) uses a mount-point approach. The NFS client works with pre-mounted NFS shares on the host system. It translates the `FileSystemClient` interface into standard filesystem operations on the mount point.

**[Visual: Show platform-specific NFS implementations]**

**Narrator**: NFS has platform-specific implementations: `nfs_client.go` for Linux, `nfs_client_darwin.go` for macOS, and `nfs_client_windows.go` for Windows. Each handles mount point detection and permission semantics appropriate for its platform.

```go
// catalog-api/filesystem/factory.go
case "nfs":
    nfsConfig := &NFSConfig{
        Host:       getStringSetting(config.Settings, "host", ""),
        Path:       getStringSetting(config.Settings, "path", ""),
        MountPoint: getStringSetting(config.Settings, "mount_point", ""),
        Options:    getStringSetting(config.Settings, "options", "vers=3"),
    }
    client, err := NewNFSClient(*nfsConfig)
    // ...
```

**[Visual: Open `catalog-api/filesystem/local_client.go`]**

**Narrator**: The local filesystem client is the simplest implementation. It delegates directly to Go's `os` package for file operations. Despite its simplicity, it still implements the full `FileSystemClient` interface, making it interchangeable with network protocols.

**[Visual: Show symbolic link resolution]**

**Narrator**: Both NFS and local clients handle symbolic links carefully. Symlinks are resolved to their targets, but the client tracks the original path to prevent infinite loops in circular symlink chains.

---

## Scene 5: Protocol Factory Pattern (50:00 - 60:00)

**[Visual: Open `catalog-api/filesystem/interface.go`]**

**Narrator**: The entire protocol abstraction rests on the `FileSystemClient` interface, which is a type alias for `digital.vasic.filesystem/pkg/client.Client` from the Filesystem submodule.

```go
// catalog-api/filesystem/interface.go
type FileSystemClient = client.Client
type StorageConfig = client.StorageConfig
type ClientFactory = client.Factory
type FileInfo = client.FileInfo
```

**[Visual: Show the interface methods]**

**Narrator**: The interface defines operations that every filesystem must support: listing directories, reading files, writing files, deleting files, creating directories, getting file info, and checking existence. New protocols are added by implementing this interface -- nothing else needs to change.

**[Visual: Open `catalog-api/filesystem/factory.go`]**

**Narrator**: The `DefaultClientFactory` implements the factory pattern. Given a `StorageConfig` with a protocol field, it creates the appropriate client. This is the single entry point for all filesystem access.

```go
// catalog-api/filesystem/factory.go
type DefaultClientFactory struct{}

func (f *DefaultClientFactory) CreateClient(config *StorageConfig) (FileSystemClient, error) {
    switch config.Protocol {
    case "smb":
        return NewSmbClient(smbConfig), nil
    case "ftp":
        return NewFTPClient(ftpConfig), nil
    case "nfs":
        return NewNFSClient(*nfsConfig)
    case "webdav":
        return NewWebDAVClient(webdavConfig), nil
    case "local":
        return NewLocalClient(localConfig), nil
    default:
        return nil, fmt.Errorf("unsupported protocol: %s", config.Protocol)
    }
}
```

**[Visual: Show protocol auto-detection concept]**

**Narrator**: The factory reads the `Protocol` field from the storage configuration. When users add a storage root through the UI, the configuration wizard helps them select the protocol and provides appropriate input fields for each one.

**[Visual: Show how the UniversalScanner uses the factory]**

**Narrator**: The `UniversalScanner` injects the factory and calls `CreateClient()` at the start of each scan job. This means the scanner never knows or cares which protocol it is using. A scan of an SMB NAS and a scan of a local directory follow the exact same code path.

**[Visual: Show the `DirectoryTreeInfo` extension]**

**Narrator**: Catalogizer extends the base module with `DirectoryTreeInfo`, a Catalogizer-specific struct for representing recursive directory trees with file counts, sizes, and depth. This is used by the browsing UI to render storage structure.

```go
// catalog-api/filesystem/interface.go
type DirectoryTreeInfo struct {
    Path       string
    TotalFiles int
    TotalDirs  int
    TotalSize  int64
    MaxDepth   int
    Files      []*FileInfo
    Subdirs    []*DirectoryTreeInfo
}
```

**[Visual: Course title card]**

**Narrator**: Protocol abstraction is what makes Catalogizer's multi-storage vision possible. A clean interface, five concrete implementations, a factory for creation, and resilience patterns for reliability. In Module 8, we optimize performance with HTTP/3 and Brotli compression.

---

## Key Code Examples

### Adding a New Protocol
```go
// 1. Implement FileSystemClient interface
type MyProtocolClient struct {
    config *MyProtocolConfig
}

func (c *MyProtocolClient) ListDirectory(ctx context.Context, path string) ([]*FileInfo, error) {
    // Protocol-specific implementation
}

// 2. Add case to factory
case "myprotocol":
    return NewMyProtocolClient(config), nil

// 3. That's it - scanner, aggregation, and UI all work automatically
```

### SMB Resilience Configuration
```go
source := &SMBSource{
    MaxRetryAttempts:    5,
    RetryDelay:          5 * time.Second,  // exponential backoff
    ConnectionTimeout:   30 * time.Second,
    HealthCheckInterval: 60 * time.Second,
}
```

### Comprehensive Test Suite
```bash
# Test all protocol implementations
cd catalog-api
go test ./filesystem/... -v

# Key test files:
# filesystem/smb_client_test.go
# filesystem/ftp_client_test.go
# filesystem/nfs_client_test.go
# filesystem/webdav_client_test.go
# filesystem/local_client_test.go
# filesystem/factory_test.go
# filesystem/comprehensive_test.go
```

---

## Quiz Questions

1. How does the `ResilientSMBManager` handle a NAS device going to sleep?
   **Answer**: When an SMB connection fails, the source transitions from Connected to Disconnected. The manager begins exponential backoff retries (configurable max attempts and delay). During reconnection, the source transitions to Reconnecting state. If retries are exhausted, it transitions to Offline and serves cached data from the `OfflineCache`. All state transitions emit events and update Prometheus metrics.

2. What makes the protocol factory pattern powerful for extensibility?
   **Answer**: Adding a new protocol requires only two steps: implement the `FileSystemClient` interface and add a case to the factory's switch statement. All existing code -- the scanner, aggregation service, browsing handlers, and UI -- works automatically with the new protocol because they depend on the interface, not concrete types.

3. How does the WebDAV client differ from SMB in its approach to file operations?
   **Answer**: WebDAV maps filesystem operations to HTTP methods (PROPFIND for listing, GET for reading, PUT for writing, DELETE for removing, MKCOL for directories). It parses XML multistatus responses. SMB uses a binary protocol with direct share mounting. WebDAV natively supports COPY and MOVE as single HTTP calls, while SMB requires iterative per-file operations.

4. Why does the NFS client have platform-specific implementation files?
   **Answer**: NFS mount semantics differ across operating systems. Linux, macOS, and Windows have different mount point detection, permission models, and NFS version support. Separate files (`nfs_client.go`, `nfs_client_darwin.go`, `nfs_client_windows.go`) use Go build tags to compile the correct implementation for each platform.
