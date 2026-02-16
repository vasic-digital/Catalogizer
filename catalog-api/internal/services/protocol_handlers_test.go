package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewLocalProtocolHandler(t *testing.T) {
	mockLogger := zap.NewNop()
	handler := NewLocalProtocolHandler(mockLogger)

	assert.NotNil(t, handler)
}

func TestNewSMBProtocolHandler(t *testing.T) {
	mockLogger := zap.NewNop()
	handler := NewSMBProtocolHandler(mockLogger)

	assert.NotNil(t, handler)
}

func TestNewFTPProtocolHandler(t *testing.T) {
	mockLogger := zap.NewNop()
	handler := NewFTPProtocolHandler(mockLogger)

	assert.NotNil(t, handler)
}

func TestNewNFSProtocolHandler(t *testing.T) {
	mockLogger := zap.NewNop()
	handler := NewNFSProtocolHandler(mockLogger)

	assert.NotNil(t, handler)
}

func TestNewWebDAVProtocolHandler(t *testing.T) {
	mockLogger := zap.NewNop()
	handler := NewWebDAVProtocolHandler(mockLogger)

	assert.NotNil(t, handler)
}

func TestProtocolHandlers_GetMoveWindow(t *testing.T) {
	mockLogger := zap.NewNop()

	tests := []struct {
		name     string
		handler  ProtocolHandler
		expected time.Duration
	}{
		{
			name:     "local handler 2 seconds",
			handler:  NewLocalProtocolHandler(mockLogger),
			expected: 2 * time.Second,
		},
		{
			name:     "SMB handler 10 seconds",
			handler:  NewSMBProtocolHandler(mockLogger),
			expected: 10 * time.Second,
		},
		{
			name:     "FTP handler 30 seconds",
			handler:  NewFTPProtocolHandler(mockLogger),
			expected: 30 * time.Second,
		},
		{
			name:     "NFS handler 5 seconds",
			handler:  NewNFSProtocolHandler(mockLogger),
			expected: 5 * time.Second,
		},
		{
			name:     "WebDAV handler 15 seconds",
			handler:  NewWebDAVProtocolHandler(mockLogger),
			expected: 15 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.handler.GetMoveWindow()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestProtocolHandlers_SupportsRealTimeNotification(t *testing.T) {
	mockLogger := zap.NewNop()

	tests := []struct {
		name     string
		handler  ProtocolHandler
		expected bool
	}{
		{
			name:     "local supports realtime",
			handler:  NewLocalProtocolHandler(mockLogger),
			expected: true,
		},
		{
			name:     "SMB does not support realtime",
			handler:  NewSMBProtocolHandler(mockLogger),
			expected: false,
		},
		{
			name:     "FTP does not support realtime",
			handler:  NewFTPProtocolHandler(mockLogger),
			expected: false,
		},
		{
			name:     "NFS does not support realtime",
			handler:  NewNFSProtocolHandler(mockLogger),
			expected: false,
		},
		{
			name:     "WebDAV does not support realtime",
			handler:  NewWebDAVProtocolHandler(mockLogger),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.handler.SupportsRealTimeNotification()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewProtocolHandlerFactory(t *testing.T) {
	mockLogger := zap.NewNop()
	factory := NewProtocolHandlerFactory(mockLogger)

	assert.NotNil(t, factory)
}

func TestProtocolHandlerFactory_GetSupportedProtocols(t *testing.T) {
	mockLogger := zap.NewNop()
	factory := NewProtocolHandlerFactory(mockLogger)

	protocols := factory.GetSupportedProtocols()
	assert.NotEmpty(t, protocols)
	assert.Contains(t, protocols, "local")
	assert.Contains(t, protocols, "smb")
	assert.Contains(t, protocols, "ftp")
	assert.Contains(t, protocols, "nfs")
	assert.Contains(t, protocols, "webdav")
}

func TestProtocolHandlerFactory_CreateHandler(t *testing.T) {
	mockLogger := zap.NewNop()
	factory := NewProtocolHandlerFactory(mockLogger)

	tests := []struct {
		name     string
		protocol string
		wantErr  bool
	}{
		{
			name:     "local handler",
			protocol: "local",
			wantErr:  false,
		},
		{
			name:     "smb handler",
			protocol: "smb",
			wantErr:  false,
		},
		{
			name:     "ftp handler",
			protocol: "ftp",
			wantErr:  false,
		},
		{
			name:     "nfs handler",
			protocol: "nfs",
			wantErr:  false,
		},
		{
			name:     "webdav handler",
			protocol: "webdav",
			wantErr:  false,
		},
		{
			name:     "unsupported protocol",
			protocol: "sftp",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, err := factory.CreateHandler(tt.protocol)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, handler)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, handler)
			}
		})
	}
}

func TestGetProtocolCapabilities(t *testing.T) {
	mockLogger := zap.NewNop()

	tests := []struct {
		name                         string
		protocol                     string
		wantErr                      bool
		expectedRealTime             bool
		expectedAtomicMove           bool
	}{
		{
			name:               "local capabilities",
			protocol:           "local",
			wantErr:            false,
			expectedRealTime:   true,
			expectedAtomicMove: true,
		},
		{
			name:               "smb capabilities",
			protocol:           "smb",
			wantErr:            false,
			expectedRealTime:   false,
			expectedAtomicMove: false,
		},
		{
			name:               "nfs capabilities",
			protocol:           "nfs",
			wantErr:            false,
			expectedRealTime:   false,
			expectedAtomicMove: true,
		},
		{
			name:     "unsupported protocol",
			protocol: "sftp",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			caps, err := GetProtocolCapabilities(tt.protocol, mockLogger)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, caps)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, caps)
				assert.Equal(t, tt.protocol, caps.Protocol)
				assert.Equal(t, tt.expectedRealTime, caps.SupportsRealTimeNotification)
				assert.Equal(t, tt.expectedAtomicMove, caps.SupportsAtomicMove)
				assert.Equal(t, !tt.expectedRealTime, caps.RequiresPolling)
			}
		})
	}
}
