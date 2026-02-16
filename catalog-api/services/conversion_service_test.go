package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewConversionService(t *testing.T) {
	service := NewConversionService(nil, nil, nil)

	assert.NotNil(t, service)
	assert.Nil(t, service.conversionRepo)
	assert.Nil(t, service.userRepo)
	assert.Nil(t, service.authService)
}

func TestConversionService_GetSupportedFormats(t *testing.T) {
	service := NewConversionService(nil, nil, nil)

	formats := service.GetSupportedFormats()

	assert.NotNil(t, formats)
	assert.Greater(t, len(formats), 0)
}

func TestConversionService_ValidateConversionRequest(t *testing.T) {
	service := NewConversionService(nil, nil, nil)

	tests := []struct {
		name   string
		input  interface{}
		expect bool
	}{
		{
			name:   "nil request",
			input:  nil,
			expect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.input == nil {
				result := service.validateConversionRequest(nil)
				assert.Equal(t, tt.expect, result)
			}
		})
	}
}

func TestConversionService_IsValidConversionType(t *testing.T) {
	service := NewConversionService(nil, nil, nil)

	tests := []struct {
		name           string
		conversionType string
		expected       bool
	}{
		{
			name:           "valid video conversion",
			conversionType: "video",
			expected:       true,
		},
		{
			name:           "valid audio conversion",
			conversionType: "audio",
			expected:       true,
		},
		{
			name:           "valid image conversion",
			conversionType: "image",
			expected:       true,
		},
		{
			name:           "valid document conversion",
			conversionType: "document",
			expected:       true,
		},
		{
			name:           "invalid conversion type",
			conversionType: "unknown",
			expected:       false,
		},
		{
			name:           "empty conversion type",
			conversionType: "",
			expected:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.isValidConversionType(tt.conversionType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConversionService_IsSupportedFormat(t *testing.T) {
	service := NewConversionService(nil, nil, nil)

	tests := []struct {
		name           string
		format         string
		conversionType string
		expected       bool
	}{
		{
			name:           "mp4 video format",
			format:         "mp4",
			conversionType: "video",
			expected:       true,
		},
		{
			name:           "mp3 audio format",
			format:         "mp3",
			conversionType: "audio",
			expected:       true,
		},
		{
			name:           "unsupported format",
			format:         "xyz",
			conversionType: "video",
			expected:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.isSupportedFormat(tt.format, tt.conversionType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConversionService_BuildFFmpegVideoArgs(t *testing.T) {
	service := NewConversionService(nil, nil, nil)

	tests := []struct {
		name     string
		source   string
		target   string
		format   string
		quality  string
		wantArgs bool
	}{
		{
			name:     "basic video conversion",
			source:   "/input/video.avi",
			target:   "/output/video.mp4",
			format:   "mp4",
			quality:  "medium",
			wantArgs: true,
		},
		{
			name:     "high quality video",
			source:   "/input/video.mkv",
			target:   "/output/video.mp4",
			format:   "mp4",
			quality:  "high",
			wantArgs: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := service.buildFFmpegVideoArgs(tt.source, tt.target, tt.format, tt.quality)
			if tt.wantArgs {
				assert.NotEmpty(t, args)
			}
		})
	}
}

func TestConversionService_BuildFFmpegAudioArgs(t *testing.T) {
	service := NewConversionService(nil, nil, nil)

	tests := []struct {
		name     string
		source   string
		target   string
		format   string
		quality  string
		wantArgs bool
	}{
		{
			name:     "basic audio conversion",
			source:   "/input/audio.wav",
			target:   "/output/audio.mp3",
			format:   "mp3",
			quality:  "medium",
			wantArgs: true,
		},
		{
			name:     "flac conversion",
			source:   "/input/audio.wav",
			target:   "/output/audio.flac",
			format:   "flac",
			quality:  "high",
			wantArgs: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := service.buildFFmpegAudioArgs(tt.source, tt.target, tt.format, tt.quality)
			if tt.wantArgs {
				assert.NotEmpty(t, args)
			}
		})
	}
}
