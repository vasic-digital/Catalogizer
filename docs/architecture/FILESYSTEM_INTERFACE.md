# Filesystem Interface Architecture

## Overview

The Catalogizer filesystem interface provides a unified abstraction layer for accessing files across multiple protocols: Local, SMB, FTP, NFS, and WebDAV. This design allows the application to treat all remote filesystems uniformly, regardless of the underlying protocol.

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Core Interfaces](#core-interfaces)
3. [Protocol Implementations](#protocol-implementations)
4. [Factory Pattern](#factory-pattern)
5. [Adding New Protocols](#adding-new-protocols)
6. [Error Handling](#error-handling)
7. [Testing Strategies](#testing-strategies)
8. [Performance Considerations](#performance-considerations)

## Architecture Overview

### Design Principles

1. **Protocol Abstraction**: Hide protocol-specific details behind a common interface
2. **Dependency Injection**: Allow easy testing and protocol switching
3. **Fail-Safe**: Graceful error handling with detailed error messages
4. **Extensibility**: Easy to add new protocols without modifying existing code
5. **Type Safety**: Compile-time checking of protocol implementations

### Component Diagram

```
┌─────────────────────────────────────────────────────────┐
│                   Application Layer                      │
│          (Handlers, Services, Repositories)              │
└────────────────────┬────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────┐
│               UnifiedClient Interface                    │
│  - Connect()      - ListDirectory()                      │
│  - Disconnect()   - CreateDirectory()                    │
│  - ReadFile()     - DeleteDirectory()                    │
│  - WriteFile()    - DeleteFile()                         │
│  - GetFileInfo()  - CopyFile()                           │
│  - FileExists()   - GetProtocol()                        │
└────────────────────┬────────────────────────────────────┘
                     │
        ┌────────────┴────────────┐
        │   Factory (NewClient)   │
        └────────────┬────────────┘
                     │
     ┌───────────────┼───────────────┬───────────┬─────────┐
     ▼               ▼               ▼           ▼         ▼
┌─────────┐   ┌──────────┐   ┌──────────┐  ┌─────────┐ ┌────────┐
│  Local  │   │   SMB    │   │   FTP    │  │   NFS   │ │ WebDAV │
│ Client  │   │  Client  │   │  Client  │  │ Client  │ │ Client │
└─────────┘   └──────────┘   └──────────┘  └─────────┘ └────────┘
     │             │               │            │           │
     ▼             ▼               ▼            ▼           ▼
┌─────────┐   ┌──────────┐   ┌──────────┐  ┌─────────┐ ┌────────┐
│  os.*   │   │ go-smb2  │   │  jlaffaye│  │ syscall │ │ net/http│
│         │   │  library │   │/ftp lib  │  │  mount  │ │  DAV   │
└─────────┘   └──────────┘   └──────────┘  └─────────┘ └────────┘
```

## Core Interfaces

### UnifiedClient Interface

**Location:** `catalog-api/filesystem/interface.go`

```go
type UnifiedClient interface {
    // Connection Management
    Connect(ctx context.Context) error
    Disconnect(ctx context.Context) error
    IsConnected() bool
    TestConnection(ctx context.Context) error

    // File Operations
    ReadFile(ctx context.Context, path string) (io.ReadCloser, error)
    WriteFile(ctx context.Context, path string, data io.Reader) error
    GetFileInfo(ctx context.Context, path string) (*FileInfo, error)
    FileExists(ctx context.Context, path string) (bool, error)
    DeleteFile(ctx context.Context, path string) error
    CopyFile(ctx context.Context, srcPath, dstPath string) error

    // Directory Operations
    ListDirectory(ctx context.Context, path string) ([]*FileInfo, error)
    CreateDirectory(ctx context.Context, path string) error
    DeleteDirectory(ctx context.Context, path string) error

    // Metadata
    GetProtocol() string
    GetConfig() interface{}
}
```

### FileInfo Structure

Standardized file metadata across all protocols:

```go
type FileInfo struct {
    Name    string      // File or directory name
    Size    int64       // File size in bytes
    ModTime time.Time   // Last modification time
    IsDir   bool        // True if directory
    Mode    os.FileMode // File permissions
    Path    string      // Full path
}
```

## Protocol Implementations

### Local Filesystem

**Implementation:** `filesystem/local_client.go`

Uses standard `os` package. Direct filesystem access with no network overhead.

**Configuration:**

```go
type LocalConfig struct {
    BasePath string // Root directory for all operations
}
```

**Example:**

```go
config := &LocalConfig{
    BasePath: "/media/storage",
}
client := NewLocalClient(config)
```

**Characteristics:**
- Fastest performance (no network)
- Direct file system access
- Uses OS permissions
- Platform-specific behavior

### SMB/CIFS Protocol

**Implementation:** `filesystem/smb_client.go`

Uses `github.com/hirochachacha/go-smb2` library. Supports Windows shares and Samba servers.

**Configuration:**

```go
type SmbConfig struct {
    Host     string // Server hostname or IP
    Port     int    // Usually 445
    Share    string // Share name (e.g., "public", "C$")
    Username string // Domain user
    Password string // Password
    Domain   string // Windows domain (e.g., "WORKGROUP")
}
```

**Example:**

```go
config := &SmbConfig{
    Host:     "nas.local",
    Port:     445,
    Share:    "media",
    Username: "admin",
    Password: "password",
    Domain:   "WORKGROUP",
}
client := NewSmbClient(config)
```

**Characteristics:**
- Native Windows protocol
- NTLM authentication
- Session-based connection
- Requires explicit mount/unmount

**Advanced Features:**
- Circuit breaker pattern (in `internal/smb/circuit_breaker.go`)
- Offline cache
- Exponential backoff retry
- Connection pooling

### FTP Protocol

**Implementation:** `filesystem/ftp_client.go`

Uses `github.com/jlaffaye/ftp` library. Supports standard FTP servers.

**Configuration:**

```go
type FTPConfig struct {
    Host     string // Server hostname or IP
    Port     int    // Usually 21
    Username string // FTP username
    Password string // FTP password
    Path     string // Base path on server
}
```

**Example:**

```go
config := &FTPConfig{
    Host:     "ftp.example.com",
    Port:     21,
    Username: "ftpuser",
    Password: "ftppass",
    Path:     "/uploads",
}
client := NewFTPClient(config)
```

**Characteristics:**
- Stateful connection
- 30-second timeout
- Path resolution with base path
- No native copy operation (download + upload)

### NFS Protocol

**Implementation:** `filesystem/nfs_client.go` (Linux), `nfs_client_darwin.go` (macOS), `nfs_client_windows.go` (Windows)

Uses system `mount` syscalls. Platform-specific implementations.

**Configuration:**

```go
type NFSConfig struct {
    Host       string // NFS server hostname
    Path       string // Export path on server (e.g., "/export/data")
    MountPoint string // Local mount directory (e.g., "/mnt/nfs")
    Options    string // Mount options (e.g., "vers=4,rw")
}
```

**Example:**

```go
config := NFSConfig{
    Host:       "nfs-server.local",
    Path:       "/export/media",
    MountPoint: "/mnt/nfs-media",
    Options:    "vers=4,rw,soft",
}
client, err := NewNFSClient(config)
```

**Characteristics:**
- Requires root/admin privileges for mounting
- Platform-specific implementation
- Once mounted, acts like local filesystem
- Directory traversal prevention

**Platform Differences:**
- **Linux**: Uses `syscall.Mount()` and reads `/proc/mounts`
- **macOS**: Uses `mount_nfs` command
- **Windows**: Uses `net use` command

### WebDAV Protocol

**Implementation:** `filesystem/webdav_client.go`

Uses `net/http` package with WebDAV-specific methods (PROPFIND, MKCOL, COPY, etc.).

**Configuration:**

```go
type WebDAVConfig struct {
    URL      string // Server URL (e.g., "https://dav.example.com")
    Username string // Basic auth username
    Password string // Basic auth password
    Path     string // Base path on server
}
```

**Example:**

```go
config := &WebDAVConfig{
    URL:      "https://nextcloud.example.com/remote.php/dav/files/user/",
    Username: "user",
    Password: "apptoken",
    Path:     "/media",
}
client := NewWebDAVClient(config)
```

**Characteristics:**
- HTTP-based protocol
- Basic authentication
- SSL/TLS support
- XML-based responses (PROPFIND)
- 30-second timeout
- Native COPY operation

**Supported Cloud Services:**
- Nextcloud/ownCloud
- Box.com
- SharePoint
- Most WebDAV-compliant services

## Factory Pattern

### NewClient Factory Function

**Location:** `filesystem/factory.go`

The factory creates the appropriate client based on configuration:

```go
func NewClient(config interface{}) (UnifiedClient, error) {
    switch cfg := config.(type) {
    case *LocalConfig:
        return NewLocalClient(cfg), nil

    case *SmbConfig:
        return NewSmbClient(cfg), nil

    case *FTPConfig:
        return NewFTPClient(cfg), nil

    case NFSConfig:
        return NewNFSClient(cfg)

    case *WebDAVConfig:
        return NewWebDAVClient(cfg), nil

    default:
        return nil, fmt.Errorf("unknown config type: %T", config)
    }
}
```

### Usage Example

```go
// Create client from database configuration
root := getStorageRootFromDB(rootID)

var client UnifiedClient
var err error

switch root.Protocol {
case "local":
    config := &LocalConfig{BasePath: root.Path}
    client, err = NewClient(config)

case "smb":
    config := &SmbConfig{
        Host:     root.SmbHost,
        Port:     root.SmbPort,
        Share:    root.SmbShare,
        Username: root.SmbUsername,
        Password: root.SmbPassword,
    }
    client, err = NewClient(config)

case "ftp":
    config := &FTPConfig{
        Host:     root.FtpHost,
        Port:     root.FtpPort,
        Username: root.FtpUsername,
        Password: root.FtpPassword,
        Path:     root.Path,
    }
    client, err = NewClient(config)

// ... and so on
}

if err != nil {
    return err
}

// Now use client uniformly
if err := client.Connect(ctx); err != nil {
    return err
}
defer client.Disconnect(ctx)

files, err := client.ListDirectory(ctx, "/")
```

## Adding New Protocols

To add a new protocol (e.g., SFTP), follow these steps:

### 1. Define Configuration Structure

```go
// filesystem/sftp_client.go
type SFTPConfig struct {
    Host       string
    Port       int
    Username   string
    Password   string
    PrivateKey string // Optional: path to private key
}
```

### 2. Implement UnifiedClient Interface

```go
type SFTPClient struct {
    config *SFTPConfig
    client *sftp.Client
    conn   *ssh.Client
}

func NewSFTPClient(config *SFTPConfig) *SFTPClient {
    return &SFTPClient{
        config: config,
    }
}

// Implement all UnifiedClient methods
func (c *SFTPClient) Connect(ctx context.Context) error {
    // SSH connection
    sshConfig := &ssh.ClientConfig{
        User: c.config.Username,
        Auth: []ssh.AuthMethod{
            ssh.Password(c.config.Password),
        },
        HostKeyCallback: ssh.InsecureIgnoreHostKey(),
    }

    addr := fmt.Sprintf("%s:%d", c.config.Host, c.config.Port)
    conn, err := ssh.Dial("tcp", addr, sshConfig)
    if err != nil {
        return fmt.Errorf("failed to connect via SSH: %w", err)
    }

    client, err := sftp.NewClient(conn)
    if err != nil {
        conn.Close()
        return fmt.Errorf("failed to create SFTP client: %w", err)
    }

    c.conn = conn
    c.client = client
    return nil
}

func (c *SFTPClient) ReadFile(ctx context.Context, path string) (io.ReadCloser, error) {
    if c.client == nil {
        return nil, fmt.Errorf("not connected")
    }
    return c.client.Open(path)
}

// ... implement all other methods
```

### 3. Add to Factory

```go
// filesystem/factory.go
func NewClient(config interface{}) (UnifiedClient, error) {
    switch cfg := config.(type) {
    // ... existing cases ...

    case *SFTPConfig:
        return NewSFTPClient(cfg), nil

    default:
        return nil, fmt.Errorf("unknown config type: %T", config)
    }
}
```

### 4. Update Database Schema

```sql
-- Add protocol option
ALTER TABLE storage_roots MODIFY COLUMN protocol
    ENUM('local', 'smb', 'ftp', 'nfs', 'webdav', 'sftp');

-- Add SFTP-specific columns
ALTER TABLE storage_roots ADD COLUMN sftp_host VARCHAR(255);
ALTER TABLE storage_roots ADD COLUMN sftp_port INT DEFAULT 22;
ALTER TABLE storage_roots ADD COLUMN sftp_username VARCHAR(255);
ALTER TABLE storage_roots ADD COLUMN sftp_password VARCHAR(255);
ALTER TABLE storage_roots ADD COLUMN sftp_private_key TEXT;
```

### 5. Create Tests

```go
// filesystem/sftp_client_test.go
func TestSFTPClient_Connect(t *testing.T) {
    config := &SFTPConfig{
        Host:     "localhost",
        Port:     22,
        Username: "testuser",
        Password: "testpass",
    }

    client := NewSFTPClient(config)
    ctx := context.Background()

    err := client.Connect(ctx)
    // Test connection handling
}

// ... more tests
```

### 6. Document Protocol

Add documentation to `docs/guides/PROTOCOL_SFTP_GUIDE.md`.

## Error Handling

### Error Wrapping

All implementations wrap errors with context:

```go
func (c *FTPClient) ReadFile(ctx context.Context, path string) (io.ReadCloser, error) {
    fullPath := c.resolvePath(path)
    resp, err := c.client.Retr(fullPath)
    if err != nil {
        return nil, fmt.Errorf("failed to retrieve FTP file %s: %w", fullPath, err)
    }
    return resp, nil
}
```

### Common Error Patterns

```go
// Not connected
if !c.IsConnected() {
    return fmt.Errorf("not connected")
}

// File not found
if err != nil {
    if isNotExistError(err) {
        return nil, os.ErrNotExist
    }
    return fmt.Errorf("failed to stat file: %w", err)
}

// Connection failed
if err != nil {
    return fmt.Errorf("failed to connect to %s server: %w", protocol, err)
}
```

### Error Type Checking

```go
// Check for specific error types
if errors.Is(err, os.ErrNotExist) {
    // Handle file not found
}

if errors.Is(err, context.Canceled) {
    // Handle context cancellation
}
```

## Testing Strategies

### Unit Testing

Test each client implementation in isolation:

```go
func TestLocalClient_ReadFile(t *testing.T) {
    // Create temp file
    tmpDir := t.TempDir()
    testFile := filepath.Join(tmpDir, "test.txt")
    content := []byte("test content")
    os.WriteFile(testFile, content, 0644)

    // Test client
    config := &LocalConfig{BasePath: tmpDir}
    client := NewLocalClient(config)

    reader, err := client.ReadFile(context.Background(), "test.txt")
    assert.NoError(t, err)
    defer reader.Close()

    data, _ := io.ReadAll(reader)
    assert.Equal(t, content, data)
}
```

### Integration Testing

Test with real protocol servers (optional, skipped by default):

```go
func TestSMBClient_IntegrationConnect(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }

    config := &SmbConfig{
        Host:     os.Getenv("SMB_TEST_HOST"),
        Port:     445,
        Share:    os.Getenv("SMB_TEST_SHARE"),
        Username: os.Getenv("SMB_TEST_USER"),
        Password: os.Getenv("SMB_TEST_PASS"),
    }

    client := NewSmbClient(config)
    err := client.Connect(context.Background())
    require.NoError(t, err)
    defer client.Disconnect(context.Background())

    assert.True(t, client.IsConnected())
}
```

### Mock Testing

Create mock implementations for testing higher layers:

```go
type MockClient struct {
    files map[string][]byte
}

func (m *MockClient) ReadFile(ctx context.Context, path string) (io.ReadCloser, error) {
    data, exists := m.files[path]
    if !exists {
        return nil, os.ErrNotExist
    }
    return io.NopCloser(bytes.NewReader(data)), nil
}

// ... implement other methods
```

## Performance Considerations

### Connection Pooling

For protocols that support it, use connection pools:

```go
type ConnectionPool struct {
    mu      sync.Mutex
    clients []*SmbClient
    config  *SmbConfig
    maxSize int
}

func (p *ConnectionPool) Get(ctx context.Context) (*SmbClient, error) {
    p.mu.Lock()
    defer p.mu.Unlock()

    if len(p.clients) > 0 {
        client := p.clients[len(p.clients)-1]
        p.clients = p.clients[:len(p.clients)-1]
        return client, nil
    }

    // Create new client if pool empty
    client := NewSmbClient(p.config)
    if err := client.Connect(ctx); err != nil {
        return nil, err
    }
    return client, nil
}

func (p *ConnectionPool) Put(client *SmbClient) {
    p.mu.Lock()
    defer p.mu.Unlock()

    if len(p.clients) < p.maxSize {
        p.clients = append(p.clients, client)
    } else {
        client.Disconnect(context.Background())
    }
}
```

### Caching

Cache file listings and metadata:

```go
type CachedClient struct {
    client    UnifiedClient
    dirCache  *cache.Cache
    fileCache *cache.Cache
}

func (c *CachedClient) ListDirectory(ctx context.Context, path string) ([]*FileInfo, error) {
    // Check cache first
    if cached, found := c.dirCache.Get(path); found {
        return cached.([]*FileInfo), nil
    }

    // Fetch from client
    files, err := c.client.ListDirectory(ctx, path)
    if err != nil {
        return nil, err
    }

    // Cache for 5 minutes
    c.dirCache.Set(path, files, 5*time.Minute)
    return files, nil
}
```

### Batch Operations

Group operations to reduce round trips:

```go
func (s *Service) DownloadFiles(ctx context.Context, client UnifiedClient, paths []string) error {
    // Download in parallel
    sem := make(chan struct{}, 10) // Max 10 concurrent downloads
    var wg sync.WaitGroup
    errors := make(chan error, len(paths))

    for _, path := range paths {
        wg.Add(1)
        go func(p string) {
            defer wg.Done()
            sem <- struct{}{}
            defer func() { <-sem }()

            if err := s.downloadFile(ctx, client, p); err != nil {
                errors <- err
            }
        }(path)
    }

    wg.Wait()
    close(errors)

    // Check for errors
    for err := range errors {
        return err
    }
    return nil
}
```

## Real-World Usage

### Example: Media Scanner

```go
func (s *Scanner) ScanStorageRoot(ctx context.Context, rootID int64) error {
    // Get root configuration from database
    root, err := s.repo.GetStorageRoot(ctx, rootID)
    if err != nil {
        return err
    }

    // Create appropriate client
    client, err := CreateClientFromRoot(root)
    if err != nil {
        return err
    }

    // Connect
    if err := client.Connect(ctx); err != nil {
        return fmt.Errorf("failed to connect: %w", err)
    }
    defer client.Disconnect(ctx)

    // Scan recursively
    return s.scanDirectory(ctx, client, "/", rootID)
}

func (s *Scanner) scanDirectory(ctx context.Context, client UnifiedClient, path string, rootID int64) error {
    files, err := client.ListDirectory(ctx, path)
    if err != nil {
        return err
    }

    for _, file := range files {
        if file.IsDir {
            // Recurse into subdirectory
            subPath := filepath.Join(path, file.Name)
            if err := s.scanDirectory(ctx, client, subPath, rootID); err != nil {
                s.logger.Error("Failed to scan directory", zap.String("path", subPath), zap.Error(err))
                continue
            }
        } else {
            // Process file
            s.processFile(ctx, file, rootID)
        }
    }

    return nil
}
```

## Summary

The filesystem interface architecture provides:

1. **Unified Access**: Single interface for all protocols
2. **Easy Extension**: Add new protocols with minimal changes
3. **Type Safety**: Compile-time verification of implementations
4. **Testability**: Easy to mock and test
5. **Performance**: Support for caching and pooling
6. **Error Handling**: Consistent error wrapping and reporting

This design allows Catalogizer to seamlessly access files across local storage, network shares, and cloud services without protocol-specific code scattered throughout the application.
