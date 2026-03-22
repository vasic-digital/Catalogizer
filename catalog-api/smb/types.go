package smb

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"go.uber.org/zap"
)

// FileInfo represents file information from SMB
type FileInfo struct {
	Name    string
	Size    int64
	ModTime time.Time
	IsDir   bool
	Mode    os.FileMode
}

// CopyOperation represents a file copy operation
type CopyOperation struct {
	SourcePath        string
	DestinationPath   string
	OverwriteExisting bool
}

// CopyResult represents the result of a copy operation
type CopyResult struct {
	Success     bool
	BytesCopied int64
	Error       error
	TimeTaken   time.Duration
}

// DirectoryTreeInfo represents directory tree information
type DirectoryTreeInfo struct {
	Path       string
	TotalFiles int
	TotalDirs  int
	TotalSize  int64
	MaxDepth   int
	Files      []*FileInfo
	Subdirs    []*DirectoryTreeInfo
}

// ConnectionPoolConfig holds configuration for the connection pool
type ConnectionPoolConfig struct {
	MaxConnections      int
	ConnectionTimeout   time.Duration
	IdleTimeout         time.Duration
	MaxLifetime         time.Duration
	HealthCheckInterval time.Duration
}

// DefaultConnectionPoolConfig returns default configuration
func DefaultConnectionPoolConfig() ConnectionPoolConfig {
	return ConnectionPoolConfig{
		MaxConnections:      10,
		ConnectionTimeout:   30 * time.Second,
		IdleTimeout:         30 * time.Second,
		MaxLifetime:         5 * time.Minute,
		HealthCheckInterval: 10 * time.Second,
	}
}

// PooledConnection wraps an SMB client with metadata
type PooledConnection struct {
	Client     *SmbClient
	Config     *SmbConfig
	CreatedAt  time.Time
	LastUsedAt time.Time
	UseCount   int64
	IsHealthy  bool
}

// IsExpired checks if the connection has exceeded its maximum lifetime
func (pc *PooledConnection) IsExpired(maxLifetime time.Duration) bool {
	return time.Since(pc.CreatedAt) > maxLifetime
}

// IsIdle checks if the connection has been idle for too long
func (pc *PooledConnection) IsIdle(idleTimeout time.Duration) bool {
	return time.Since(pc.LastUsedAt) > idleTimeout
}

// SmbConnectionPool manages multiple SMB connections with proper lifecycle
type SmbConnectionPool struct {
	connections    map[string]*PooledConnection
	maxConnections int
	config         ConnectionPoolConfig
	mu             sync.RWMutex
	logger         *zap.Logger

	// Cleanup goroutine management
	cleanupTicker *time.Ticker
	cleanupDone   chan struct{}
	isRunning     bool

	// Metrics
	totalConnections   int64
	activeConnections  int64
	expiredConnections int64
}

// NewSmbConnectionPool creates a new connection pool with cleanup
func NewSmbConnectionPool(maxConnections int) *SmbConnectionPool {
	return NewSmbConnectionPoolWithConfig(maxConnections, DefaultConnectionPoolConfig(), nil)
}

// NewSmbConnectionPoolWithConfig creates a new connection pool with custom configuration
func NewSmbConnectionPoolWithConfig(maxConnections int, config ConnectionPoolConfig, logger *zap.Logger) *SmbConnectionPool {
	if maxConnections <= 0 {
		maxConnections = 10
	}

	pool := &SmbConnectionPool{
		connections:    make(map[string]*PooledConnection),
		maxConnections: maxConnections,
		config:         config,
		logger:         logger,
		cleanupDone:    make(chan struct{}),
	}

	// Start cleanup goroutine
	pool.StartCleanup()

	if logger != nil {
		logger.Info("SMB connection pool created",
			zap.Int("max_connections", maxConnections),
			zap.Duration("idle_timeout", config.IdleTimeout),
			zap.Duration("max_lifetime", config.MaxLifetime))
	}

	return pool
}

// StartCleanup starts the background cleanup goroutine
func (p *SmbConnectionPool) StartCleanup() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.isRunning {
		return
	}

	p.cleanupTicker = time.NewTicker(p.config.HealthCheckInterval)
	p.isRunning = true

	go p.cleanupLoop()
}

// StopCleanup stops the background cleanup goroutine
func (p *SmbConnectionPool) StopCleanup() {
	p.mu.Lock()
	if !p.isRunning {
		p.mu.Unlock()
		return
	}

	p.isRunning = false
	if p.cleanupTicker != nil {
		p.cleanupTicker.Stop()
	}
	close(p.cleanupDone)
	p.mu.Unlock()

	// Wait for cleanup loop to finish
	<-p.cleanupDone
}

// cleanupLoop runs the periodic cleanup
func (p *SmbConnectionPool) cleanupLoop() {
	for {
		select {
		case <-p.cleanupTicker.C:
			p.cleanupIdleConnections()
		case <-p.cleanupDone:
			return
		}
	}
}

// cleanupIdleConnections removes expired and idle connections
func (p *SmbConnectionPool) cleanupIdleConnections() {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now()
	var expired []string

	for key, conn := range p.connections {
		// Check if connection is expired or idle
		if conn.IsExpired(p.config.MaxLifetime) || conn.IsIdle(p.config.IdleTimeout) {
			expired = append(expired, key)
			p.expiredConnections++

			if p.logger != nil {
				p.logger.Debug("Closing expired/idle SMB connection",
					zap.String("key", key),
					zap.Time("created_at", conn.CreatedAt),
					zap.Time("last_used", conn.LastUsedAt),
					zap.Duration("lifetime", now.Sub(conn.CreatedAt)),
					zap.Duration("idle_time", now.Sub(conn.LastUsedAt)))
			}

			// Close the connection
			if conn.Client != nil {
				conn.Client.Close()
			}
		}
	}

	// Remove expired connections from map
	for _, key := range expired {
		delete(p.connections, key)
		p.activeConnections--
	}

	if len(expired) > 0 && p.logger != nil {
		p.logger.Info("Cleaned up SMB connections",
			zap.Int("expired_count", len(expired)),
			zap.Int("remaining", len(p.connections)))
	}
}

// GetConnection gets or creates an SMB connection with proper lifecycle management
func (p *SmbConnectionPool) GetConnection(key string, config *SmbConfig) (*SmbClient, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Check for existing connection
	if pooledConn, exists := p.connections[key]; exists {
		// Check if connection is still valid
		if !pooledConn.IsExpired(p.config.MaxLifetime) {
			if err := pooledConn.Client.TestConnection(); err == nil {
				// Connection is healthy, update metadata
				pooledConn.LastUsedAt = time.Now()
				pooledConn.UseCount++
				pooledConn.IsHealthy = true
				return pooledConn.Client, nil
			}
		}

		// Connection is stale or expired, close it
		if p.logger != nil {
			p.logger.Debug("Closing stale SMB connection",
				zap.String("key", key),
				zap.Bool("expired", pooledConn.IsExpired(p.config.MaxLifetime)))
		}

		pooledConn.Client.Close()
		delete(p.connections, key)
		p.activeConnections--
	}

	// Create new connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), p.config.ConnectionTimeout)
	defer cancel()

	client, err := p.createConnectionWithContext(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create SMB connection: %w", err)
	}

	// Store in pool if there's space
	if len(p.connections) < p.maxConnections {
		p.connections[key] = &PooledConnection{
			Client:     client,
			Config:     config,
			CreatedAt:  time.Now(),
			LastUsedAt: time.Now(),
			UseCount:   1,
			IsHealthy:  true,
		}
		p.activeConnections++
		p.totalConnections++

		if p.logger != nil {
			p.logger.Debug("New SMB connection added to pool",
				zap.String("key", key),
				zap.Int("pool_size", len(p.connections)),
				zap.Int("max_size", p.maxConnections))
		}
	} else if p.logger != nil {
		p.logger.Warn("SMB connection pool at capacity, connection not stored",
			zap.String("key", key),
			zap.Int("max_size", p.maxConnections))
	}

	return client, nil
}

// createConnectionWithContext creates an SMB connection with context support
func (p *SmbConnectionPool) createConnectionWithContext(ctx context.Context, config *SmbConfig) (*SmbClient, error) {
	// Validate config
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// Create a channel for the connection result
	type result struct {
		client *SmbClient
		err    error
	}
	resultChan := make(chan result, 1)

	// Create connection in a goroutine
	go func() {
		defer func() {
			if r := recover(); r != nil {
				resultChan <- result{client: nil, err: fmt.Errorf("panic during connection: %v", r)}
			}
		}()
		client, err := NewSmbClient(config)
		resultChan <- result{client: client, err: err}
	}()

	// Wait for either the connection or context cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case res := <-resultChan:
		return res.client, res.err
	}
}

// CloseAll closes all connections in the pool and stops cleanup
func (p *SmbConnectionPool) CloseAll() {
	// Stop cleanup first
	p.StopCleanup()

	p.mu.Lock()
	defer p.mu.Unlock()

	for key, conn := range p.connections {
		if conn.Client != nil {
			conn.Client.Close()
		}
		delete(p.connections, key)
	}

	p.activeConnections = 0

	if p.logger != nil {
		p.logger.Info("All SMB connections closed",
			zap.Int64("total_created", p.totalConnections),
			zap.Int64("total_expired", p.expiredConnections))
	}
}

// GetStats returns pool statistics
func (p *SmbConnectionPool) GetStats() ConnectionPoolStats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return ConnectionPoolStats{
		ActiveConnections:  p.activeConnections,
		TotalConnections:   p.totalConnections,
		ExpiredConnections: p.expiredConnections,
		PoolSize:           int64(len(p.connections)),
		MaxConnections:     int64(p.maxConnections),
	}
}

// ConnectionPoolStats holds pool statistics
type ConnectionPoolStats struct {
	ActiveConnections  int64
	TotalConnections   int64
	ExpiredConnections int64
	PoolSize           int64
	MaxConnections     int64
}

// String returns a string representation of stats
func (s ConnectionPoolStats) String() string {
	return fmt.Sprintf("PoolStats{active=%d, total=%d, expired=%d, size=%d/%d}",
		s.ActiveConnections, s.TotalConnections, s.ExpiredConnections,
		s.PoolSize, s.MaxConnections)
}

// ForceCleanup immediately triggers a cleanup of idle connections
func (p *SmbConnectionPool) ForceCleanup() {
	p.cleanupIdleConnections()
}
