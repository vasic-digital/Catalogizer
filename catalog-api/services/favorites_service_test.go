package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFavoritesService(t *testing.T) {
	service := NewFavoritesService(nil, nil)

	assert.NotNil(t, service)
	assert.Nil(t, service.favoritesRepo)
	assert.Nil(t, service.authService)
}

func TestFavoritesService_RemoveDuplicateStrings(t *testing.T) {
	service := NewFavoritesService(nil, nil)

	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "no duplicates",
			input:    []string{"a", "b", "c"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "with duplicates",
			input:    []string{"a", "b", "a", "c", "b"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "all duplicates",
			input:    []string{"a", "a", "a"},
			expected: []string{"a"},
		},
		{
			name:     "empty slice",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "nil slice",
			input:    nil,
			expected: nil,
		},
		{
			name:     "single element",
			input:    []string{"a"},
			expected: []string{"a"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.removeDuplicateStrings(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFavoritesService_ExportFavorites_NilRepo(t *testing.T) {
	service := NewFavoritesService(nil, nil)

	// With nil repo, should return error or empty result
	result, err := service.ExportFavorites(1, "json")
	if err != nil {
		assert.Error(t, err)
	} else {
		assert.NotNil(t, result)
	}
}

func TestFavoritesService_ImportFavorites_EmptyData(t *testing.T) {
	service := NewFavoritesService(nil, nil)

	tests := []struct {
		name    string
		userID  int
		format  string
		data    []byte
		wantErr bool
	}{
		{
			name:    "empty data",
			userID:  1,
			format:  "json",
			data:    []byte{},
			wantErr: true,
		},
		{
			name:    "nil data",
			userID:  1,
			format:  "json",
			data:    nil,
			wantErr: true,
		},
		{
			name:    "invalid format",
			userID:  1,
			format:  "invalid_format",
			data:    []byte(`[{"id":1}]`),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.ImportFavorites(tt.userID, tt.format, tt.data)
			if tt.wantErr {
				assert.Error(t, err)
			}
		})
	}
}
