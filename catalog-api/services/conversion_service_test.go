package services

import (
	"testing"

	"catalogizer/models"

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
	// Verify that supported formats struct has entries in each category
	assert.Greater(t, len(formats.Video.Input), 0)
	assert.Greater(t, len(formats.Audio.Input), 0)
	assert.Greater(t, len(formats.Document.Input), 0)
	assert.Greater(t, len(formats.Image.Input), 0)
}

func TestConversionService_ValidateConversionRequest(t *testing.T) {
	service := NewConversionService(nil, nil, nil)

	tests := []struct {
		name   string
		input  *models.ConversionRequest
		expect bool
	}{
		{
			name: "empty request",
			input: &models.ConversionRequest{
				SourcePath:     "",
				TargetPath:     "",
				SourceFormat:   "",
				TargetFormat:   "",
				ConversionType: "",
			},
			expect: false,
		},
		{
			name: "missing source path",
			input: &models.ConversionRequest{
				SourcePath:     "",
				TargetPath:     "/output/video.mp4",
				SourceFormat:   "avi",
				TargetFormat:   "mp4",
				ConversionType: "video",
			},
			expect: false,
		},
		{
			name: "missing target path",
			input: &models.ConversionRequest{
				SourcePath:     "/input/video.avi",
				TargetPath:     "",
				SourceFormat:   "avi",
				TargetFormat:   "mp4",
				ConversionType: "video",
			},
			expect: false,
		},
		{
			name: "missing source format",
			input: &models.ConversionRequest{
				SourcePath:     "/input/video.avi",
				TargetPath:     "/output/video.mp4",
				SourceFormat:   "",
				TargetFormat:   "mp4",
				ConversionType: "video",
			},
			expect: false,
		},
		{
			name: "missing target format",
			input: &models.ConversionRequest{
				SourcePath:     "/input/video.avi",
				TargetPath:     "/output/video.mp4",
				SourceFormat:   "avi",
				TargetFormat:   "",
				ConversionType: "video",
			},
			expect: false,
		},
		{
			name: "missing conversion type",
			input: &models.ConversionRequest{
				SourcePath:     "/input/video.avi",
				TargetPath:     "/output/video.mp4",
				SourceFormat:   "avi",
				TargetFormat:   "mp4",
				ConversionType: "",
			},
			expect: false,
		},
		{
			name: "invalid conversion type",
			input: &models.ConversionRequest{
				SourcePath:     "/input/video.avi",
				TargetPath:     "/output/video.mp4",
				SourceFormat:   "avi",
				TargetFormat:   "mp4",
				ConversionType: "unknown",
			},
			expect: false,
		},
		{
			name: "unsupported format",
			input: &models.ConversionRequest{
				SourcePath:     "/input/video.xyz",
				TargetPath:     "/output/video.abc",
				SourceFormat:   "xyz",
				TargetFormat:   "abc",
				ConversionType: "video",
			},
			expect: false,
		},
		{
			name: "valid video conversion",
			input: &models.ConversionRequest{
				SourcePath:     "/input/video.avi",
				TargetPath:     "/output/video.mp4",
				SourceFormat:   "avi",
				TargetFormat:   "mp4",
				ConversionType: "video",
			},
			expect: true,
		},
		{
			name: "valid audio conversion",
			input: &models.ConversionRequest{
				SourcePath:     "/input/audio.wav",
				TargetPath:     "/output/audio.mp3",
				SourceFormat:   "wav",
				TargetFormat:   "mp3",
				ConversionType: "audio",
			},
			expect: true,
		},
		{
			name: "valid image conversion",
			input: &models.ConversionRequest{
				SourcePath:     "/input/image.png",
				TargetPath:     "/output/image.jpg",
				SourceFormat:   "png",
				TargetFormat:   "jpg",
				ConversionType: "image",
			},
			expect: true,
		},
		{
			name: "valid document conversion",
			input: &models.ConversionRequest{
				SourcePath:     "/input/doc.pdf",
				TargetPath:     "/output/doc.txt",
				SourceFormat:   "pdf",
				TargetFormat:   "txt",
				ConversionType: "document",
			},
			expect: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.validateConversionRequest(tt.input)
			assert.Equal(t, tt.expect, result)
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
		sourceFormat   string
		targetFormat   string
		conversionType string
		expected       bool
	}{
		{
			name:           "mp4 video format",
			sourceFormat:   "mp4",
			targetFormat:   "mp4",
			conversionType: "video",
			expected:       true,
		},
		{
			name:           "mp3 audio format",
			sourceFormat:   "mp3",
			targetFormat:   "mp3",
			conversionType: "audio",
			expected:       true,
		},
		{
			name:           "unsupported format",
			sourceFormat:   "xyz",
			targetFormat:   "xyz",
			conversionType: "video",
			expected:       false,
		},
		{
			name:           "jpg image format",
			sourceFormat:   "jpg",
			targetFormat:   "png",
			conversionType: "image",
			expected:       true,
		},
		{
			name:           "png image format uppercase",
			sourceFormat:   "PNG",
			targetFormat:   "JPG",
			conversionType: "image",
			expected:       true,
		},
		{
			name:           "epub document format",
			sourceFormat:   "epub",
			targetFormat:   "pdf",
			conversionType: "document",
			expected:       true,
		},
		{
			name:           "source supported target unsupported video",
			sourceFormat:   "mp4",
			targetFormat:   "xyz",
			conversionType: "video",
			expected:       false,
		},
		{
			name:           "source unsupported target supported video",
			sourceFormat:   "xyz",
			targetFormat:   "mp4",
			conversionType: "video",
			expected:       false,
		},
		{
			name:           "invalid conversion type",
			sourceFormat:   "mp4",
			targetFormat:   "mp4",
			conversionType: "invalid",
			expected:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.isSupportedFormat(tt.conversionType, tt.sourceFormat, tt.targetFormat)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConversionService_BuildFFmpegVideoArgs(t *testing.T) {
	service := NewConversionService(nil, nil, nil)

	tests := []struct {
		name     string
		job      *models.ConversionJob
		expected []string
	}{
		{
			name: "basic video conversion medium quality",
			job: &models.ConversionJob{
				SourcePath:   "/input/video.avi",
				TargetPath:   "/output/video.mp4",
				TargetFormat: "mp4",
				Quality:      "medium",
			},
			expected: []string{"-i", "/input/video.avi", "-y", "-crf", "23", "-preset", "medium", "/output/video.mp4"},
		},
		{
			name: "low quality video",
			job: &models.ConversionJob{
				SourcePath:   "/input/video.avi",
				TargetPath:   "/output/video.mp4",
				TargetFormat: "mp4",
				Quality:      "low",
			},
			expected: []string{"-i", "/input/video.avi", "-y", "-crf", "28", "-preset", "fast", "/output/video.mp4"},
		},
		{
			name: "high quality video",
			job: &models.ConversionJob{
				SourcePath:   "/input/video.avi",
				TargetPath:   "/output/video.mp4",
				TargetFormat: "mp4",
				Quality:      "high",
			},
			expected: []string{"-i", "/input/video.avi", "-y", "-crf", "18", "-preset", "slow", "/output/video.mp4"},
		},
		{
			name: "lossless quality video",
			job: &models.ConversionJob{
				SourcePath:   "/input/video.avi",
				TargetPath:   "/output/video.mp4",
				TargetFormat: "mp4",
				Quality:      "lossless",
			},
			expected: []string{"-i", "/input/video.avi", "-y", "-crf", "0", "-preset", "veryslow", "/output/video.mp4"},
		},
		{
			name: "default quality when unknown",
			job: &models.ConversionJob{
				SourcePath:   "/input/video.avi",
				TargetPath:   "/output/video.mp4",
				TargetFormat: "mp4",
				Quality:      "unknown",
			},
			expected: []string{"-i", "/input/video.avi", "-y", "-crf", "23", "-preset", "medium", "/output/video.mp4"},
		},
		{
			name: "with resolution setting",
			job: &models.ConversionJob{
				SourcePath:   "/input/video.avi",
				TargetPath:   "/output/video.mp4",
				TargetFormat: "mp4",
				Quality:      "medium",
				Settings:     ptr(`{"resolution": "1920x1080"}`),
			},
			expected: []string{"-i", "/input/video.avi", "-y", "-crf", "23", "-preset", "medium", "-s", "1920x1080", "/output/video.mp4"},
		},
		{
			name: "with framerate setting",
			job: &models.ConversionJob{
				SourcePath:   "/input/video.avi",
				TargetPath:   "/output/video.mp4",
				TargetFormat: "mp4",
				Quality:      "medium",
				Settings:     ptr(`{"framerate": "30"}`),
			},
			expected: []string{"-i", "/input/video.avi", "-y", "-crf", "23", "-preset", "medium", "-r", "30", "/output/video.mp4"},
		},
		{
			name: "with bitrate setting",
			job: &models.ConversionJob{
				SourcePath:   "/input/video.avi",
				TargetPath:   "/output/video.mp4",
				TargetFormat: "mp4",
				Quality:      "medium",
				Settings:     ptr(`{"bitrate": "2M"}`),
			},
			expected: []string{"-i", "/input/video.avi", "-y", "-crf", "23", "-preset", "medium", "-b:v", "2M", "/output/video.mp4"},
		},
		{
			name: "with all settings",
			job: &models.ConversionJob{
				SourcePath:   "/input/video.avi",
				TargetPath:   "/output/video.mp4",
				TargetFormat: "mp4",
				Quality:      "medium",
				Settings:     ptr(`{"resolution": "1280x720", "framerate": "24", "bitrate": "1M"}`),
			},
			expected: []string{"-i", "/input/video.avi", "-y", "-crf", "23", "-preset", "medium", "-s", "1280x720", "-r", "24", "-b:v", "1M", "/output/video.mp4"},
		},
		{
			name: "invalid settings json",
			job: &models.ConversionJob{
				SourcePath:   "/input/video.avi",
				TargetPath:   "/output/video.mp4",
				TargetFormat: "mp4",
				Quality:      "medium",
				Settings:     ptr(`invalid json`),
			},
			expected: []string{"-i", "/input/video.avi", "-y", "-crf", "23", "-preset", "medium", "/output/video.mp4"},
		},
		{
			name: "nil settings",
			job: &models.ConversionJob{
				SourcePath:   "/input/video.avi",
				TargetPath:   "/output/video.mp4",
				TargetFormat: "mp4",
				Quality:      "medium",
				Settings:     nil,
			},
			expected: []string{"-i", "/input/video.avi", "-y", "-crf", "23", "-preset", "medium", "/output/video.mp4"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := service.buildFFmpegVideoArgs(tt.job)
			assert.Equal(t, tt.expected, args)
		})
	}
}

func TestConversionService_BuildFFmpegAudioArgs(t *testing.T) {
	service := NewConversionService(nil, nil, nil)

	tests := []struct {
		name     string
		job      *models.ConversionJob
		expected []string
	}{
		{
			name: "basic audio conversion medium quality",
			job: &models.ConversionJob{
				SourcePath:   "/input/audio.wav",
				TargetPath:   "/output/audio.mp3",
				TargetFormat: "mp3",
				Quality:      "medium",
			},
			expected: []string{"-i", "/input/audio.wav", "-y", "-ab", "192k", "/output/audio.mp3"},
		},
		{
			name: "low quality audio",
			job: &models.ConversionJob{
				SourcePath:   "/input/audio.wav",
				TargetPath:   "/output/audio.mp3",
				TargetFormat: "mp3",
				Quality:      "low",
			},
			expected: []string{"-i", "/input/audio.wav", "-y", "-ab", "96k", "/output/audio.mp3"},
		},
		{
			name: "high quality audio",
			job: &models.ConversionJob{
				SourcePath:   "/input/audio.wav",
				TargetPath:   "/output/audio.mp3",
				TargetFormat: "mp3",
				Quality:      "high",
			},
			expected: []string{"-i", "/input/audio.wav", "-y", "-ab", "320k", "/output/audio.mp3"},
		},
		{
			name: "lossless quality audio (flac)",
			job: &models.ConversionJob{
				SourcePath:   "/input/audio.wav",
				TargetPath:   "/output/audio.flac",
				TargetFormat: "flac",
				Quality:      "lossless",
			},
			expected: []string{"-i", "/input/audio.wav", "-y", "-c:a", "flac", "/output/audio.flac"},
		},
		{
			name: "default quality when unknown",
			job: &models.ConversionJob{
				SourcePath:   "/input/audio.wav",
				TargetPath:   "/output/audio.mp3",
				TargetFormat: "mp3",
				Quality:      "unknown",
			},
			expected: []string{"-i", "/input/audio.wav", "-y", "-ab", "192k", "/output/audio.mp3"},
		},
		{
			name: "with sample rate setting",
			job: &models.ConversionJob{
				SourcePath:   "/input/audio.wav",
				TargetPath:   "/output/audio.mp3",
				TargetFormat: "mp3",
				Quality:      "medium",
				Settings:     ptr(`{"sample_rate": "44100"}`),
			},
			expected: []string{"-i", "/input/audio.wav", "-y", "-ab", "192k", "-ar", "44100", "/output/audio.mp3"},
		},
		{
			name: "with channels setting",
			job: &models.ConversionJob{
				SourcePath:   "/input/audio.wav",
				TargetPath:   "/output/audio.mp3",
				TargetFormat: "mp3",
				Quality:      "medium",
				Settings:     ptr(`{"channels": "2"}`),
			},
			expected: []string{"-i", "/input/audio.wav", "-y", "-ab", "192k", "-ac", "2", "/output/audio.mp3"},
		},
		{
			name: "with both settings",
			job: &models.ConversionJob{
				SourcePath:   "/input/audio.wav",
				TargetPath:   "/output/audio.mp3",
				TargetFormat: "mp3",
				Quality:      "medium",
				Settings:     ptr(`{"sample_rate": "48000", "channels": "1"}`),
			},
			expected: []string{"-i", "/input/audio.wav", "-y", "-ab", "192k", "-ar", "48000", "-ac", "1", "/output/audio.mp3"},
		},
		{
			name: "invalid settings json",
			job: &models.ConversionJob{
				SourcePath:   "/input/audio.wav",
				TargetPath:   "/output/audio.mp3",
				TargetFormat: "mp3",
				Quality:      "medium",
				Settings:     ptr(`invalid json`),
			},
			expected: []string{"-i", "/input/audio.wav", "-y", "-ab", "192k", "/output/audio.mp3"},
		},
		{
			name: "nil settings",
			job: &models.ConversionJob{
				SourcePath:   "/input/audio.wav",
				TargetPath:   "/output/audio.mp3",
				TargetFormat: "mp3",
				Quality:      "medium",
				Settings:     nil,
			},
			expected: []string{"-i", "/input/audio.wav", "-y", "-ab", "192k", "/output/audio.mp3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := service.buildFFmpegAudioArgs(tt.job)
			assert.Equal(t, tt.expected, args)
		})
	}
}

func TestConversionService_BuildImageMagickArgs(t *testing.T) {
	service := NewConversionService(nil, nil, nil)

	tests := []struct {
		name           string
		job            *models.ConversionJob
		expectedArgLen int
	}{
		{
			name: "basic image conversion",
			job: &models.ConversionJob{
				SourcePath: "/input/image.png",
				TargetPath: "/output/image.jpg",
			},
			expectedArgLen: 2,
		},
		{
			name: "with resize setting",
			job: &models.ConversionJob{
				SourcePath: "/input/image.png",
				TargetPath: "/output/image.jpg",
				Settings:   ptr(`{"resize": "1920x1080"}`),
			},
			expectedArgLen: 4,
		},
		{
			name: "with quality setting",
			job: &models.ConversionJob{
				SourcePath: "/input/image.png",
				TargetPath: "/output/image.jpg",
				Settings:   ptr(`{"quality": "90"}`),
			},
			expectedArgLen: 4,
		},
		{
			name: "with compress setting",
			job: &models.ConversionJob{
				SourcePath: "/input/image.png",
				TargetPath: "/output/image.jpg",
				Settings:   ptr(`{"compress": true}`),
			},
			expectedArgLen: 4,
		},
		{
			name: "with all settings",
			job: &models.ConversionJob{
				SourcePath: "/input/image.png",
				TargetPath: "/output/image.jpg",
				Settings:   ptr(`{"resize": "800x600", "quality": "85", "compress": true}`),
			},
			expectedArgLen: 8,
		},
		{
			name: "invalid settings json",
			job: &models.ConversionJob{
				SourcePath: "/input/image.png",
				TargetPath: "/output/image.jpg",
				Settings:   ptr(`invalid json`),
			},
			expectedArgLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := service.buildImageMagickArgs(tt.job)
			assert.Len(t, args, tt.expectedArgLen)
		})
	}
}

func TestConversionService_IsEbookConversion(t *testing.T) {
	service := NewConversionService(nil, nil, nil)

	tests := []struct {
		name         string
		job          *models.ConversionJob
		expectedBool bool
	}{
		{"epub to mobi", &models.ConversionJob{SourceFormat: "epub", TargetFormat: "mobi"}, true},
		{"mobi to epub", &models.ConversionJob{SourceFormat: "mobi", TargetFormat: "epub"}, true},
		{"pdf to epub", &models.ConversionJob{SourceFormat: "pdf", TargetFormat: "epub"}, true},
		{"epub to pdf", &models.ConversionJob{SourceFormat: "epub", TargetFormat: "pdf"}, true},
		{"epub to txt", &models.ConversionJob{SourceFormat: "epub", TargetFormat: "txt"}, true},
		{"txt to epub", &models.ConversionJob{SourceFormat: "txt", TargetFormat: "epub"}, true},
		{"pdf to jpg", &models.ConversionJob{SourceFormat: "pdf", TargetFormat: "jpg"}, false},
		{"mp4 to webm", &models.ConversionJob{SourceFormat: "mp4", TargetFormat: "webm"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.isEbookConversion(tt.job)
			assert.Equal(t, tt.expectedBool, result)
		})
	}
}

func TestConversionService_IsPDFConversion(t *testing.T) {
	service := NewConversionService(nil, nil, nil)

	tests := []struct {
		name         string
		job          *models.ConversionJob
		expectedBool bool
	}{
		{"pdf to jpg", &models.ConversionJob{SourceFormat: "pdf", TargetFormat: "jpg"}, true},
		{"pdf to png", &models.ConversionJob{SourceFormat: "pdf", TargetFormat: "png"}, true},
		{"pdf to txt", &models.ConversionJob{SourceFormat: "pdf", TargetFormat: "txt"}, true},
		{"pdf to html", &models.ConversionJob{SourceFormat: "pdf", TargetFormat: "html"}, true},
		{"jpg to pdf", &models.ConversionJob{SourceFormat: "jpg", TargetFormat: "pdf"}, true},
		{"epub to pdf", &models.ConversionJob{SourceFormat: "epub", TargetFormat: "pdf"}, true},
		{"doc to pdf", &models.ConversionJob{SourceFormat: "doc", TargetFormat: "pdf"}, true},
		{"mp4 to webm", &models.ConversionJob{SourceFormat: "mp4", TargetFormat: "webm"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.isPDFConversion(tt.job)
			assert.Equal(t, tt.expectedBool, result)
		})
	}
}

func ptr(s string) *string {
	return &s
}
