package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSyncService(t *testing.T) {
	service := NewSyncService(nil, nil, nil)

	assert.NotNil(t, service)
}

func TestSyncService_ValidateSyncEndpoint(t *testing.T) {
	service := NewSyncService(nil, nil, nil)

	tests := []struct {
		name     string
		endpoint string
		wantErr  bool
	}{
		{
			name:     "valid http endpoint",
			endpoint: "http://example.com/api",
			wantErr:  false,
		},
		{
			name:     "valid https endpoint",
			endpoint: "https://example.com/api",
			wantErr:  false,
		},
		{
			name:     "empty endpoint",
			endpoint: "",
			wantErr:  true,
		},
		{
			name:     "invalid endpoint",
			endpoint: "not-a-url",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validateSyncEndpoint(tt.endpoint)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSyncService_IsValidType(t *testing.T) {
	service := NewSyncService(nil, nil, nil)

	tests := []struct {
		name     string
		syncType string
		expected bool
	}{
		{
			name:     "valid full sync type",
			syncType: "full",
			expected: true,
		},
		{
			name:     "valid incremental sync type",
			syncType: "incremental",
			expected: true,
		},
		{
			name:     "valid selective sync type",
			syncType: "selective",
			expected: true,
		},
		{
			name:     "invalid sync type",
			syncType: "unknown",
			expected: false,
		},
		{
			name:     "empty sync type",
			syncType: "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.isValidType(tt.syncType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSyncService_ShouldSkipFile(t *testing.T) {
	service := NewSyncService(nil, nil, nil)

	tests := []struct {
		name     string
		filename string
		patterns []string
		expected bool
	}{
		{
			name:     "no patterns means no skip",
			filename: "test.txt",
			patterns: []string{},
			expected: false,
		},
		{
			name:     "matching pattern skips",
			filename: ".gitignore",
			patterns: []string{".*"},
			expected: true,
		},
		{
			name:     "non-matching pattern does not skip",
			filename: "test.txt",
			patterns: []string{"*.log"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.shouldSkipFile(tt.filename, tt.patterns)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSyncService_ShouldRunSchedule(t *testing.T) {
	service := NewSyncService(nil, nil, nil)

	tests := []struct {
		name     string
		schedule string
		expected bool
	}{
		{
			name:     "always schedule",
			schedule: "always",
			expected: true,
		},
		{
			name:     "empty schedule",
			schedule: "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.shouldRunSchedule(tt.schedule)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSyncService_CalculateChecksum(t *testing.T) {
	service := NewSyncService(nil, nil, nil)

	tests := []struct {
		name string
		data []byte
	}{
		{
			name: "basic data",
			data: []byte("hello world"),
		},
		{
			name: "empty data",
			data: []byte{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checksum := service.calculateChecksum(tt.data)
			assert.NotEmpty(t, checksum)

			// Same input should produce same output
			checksum2 := service.calculateChecksum(tt.data)
			assert.Equal(t, checksum, checksum2)
		})
	}
}
