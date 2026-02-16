package services

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewPlaylistService(t *testing.T) {
	var mockDB *sql.DB
	mockLogger := zap.NewNop()

	service := NewPlaylistService(mockDB, mockLogger)

	assert.NotNil(t, service)
}

func TestPlaylistService_BuildSmartPlaylistQuery(t *testing.T) {
	var mockDB *sql.DB
	mockLogger := zap.NewNop()
	service := NewPlaylistService(mockDB, mockLogger)

	tests := []struct {
		name         string
		criteria     *SmartPlaylistCriteria
		wantNotEmpty bool
	}{
		{
			name: "genre equals rule",
			criteria: &SmartPlaylistCriteria{
				Rules: []SmartRule{
					{Field: "genre", Operator: "equals", Value: "Rock"},
				},
				Logic: "AND",
				Limit: 50,
				Order: "added_desc",
			},
			wantNotEmpty: true,
		},
		{
			name: "artist contains rule",
			criteria: &SmartPlaylistCriteria{
				Rules: []SmartRule{
					{Field: "artist", Operator: "contains", Value: "Beatles"},
				},
				Logic: "AND",
			},
			wantNotEmpty: true,
		},
		{
			name: "multiple rules with OR logic",
			criteria: &SmartPlaylistCriteria{
				Rules: []SmartRule{
					{Field: "genre", Operator: "equals", Value: "Rock"},
					{Field: "year", Operator: "greater_than", Value: 2000},
				},
				Logic: "OR",
				Limit: 100,
			},
			wantNotEmpty: true,
		},
		{
			name: "year greater than rule",
			criteria: &SmartPlaylistCriteria{
				Rules: []SmartRule{
					{Field: "year", Operator: "greater_than", Value: 2020},
				},
				Logic: "AND",
			},
			wantNotEmpty: true,
		},
		{
			name: "rating greater than rule",
			criteria: &SmartPlaylistCriteria{
				Rules: []SmartRule{
					{Field: "rating", Operator: "greater_than", Value: 4},
				},
				Logic: "AND",
			},
			wantNotEmpty: true,
		},
		{
			name: "empty rules",
			criteria: &SmartPlaylistCriteria{
				Rules: []SmartRule{},
				Logic: "AND",
			},
			wantNotEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, args := service.buildSmartPlaylistQuery(tt.criteria)
			if tt.wantNotEmpty {
				assert.NotEmpty(t, query)
			}
			assert.NotNil(t, args)
		})
	}
}

func TestPlaylistService_BuildRuleCondition(t *testing.T) {
	var mockDB *sql.DB
	mockLogger := zap.NewNop()
	service := NewPlaylistService(mockDB, mockLogger)

	tests := []struct {
		name          string
		rule          SmartRule
		wantCondition bool
	}{
		{
			name:          "genre equals",
			rule:          SmartRule{Field: "genre", Operator: "equals", Value: "Jazz"},
			wantCondition: true,
		},
		{
			name:          "genre contains",
			rule:          SmartRule{Field: "genre", Operator: "contains", Value: "Rock"},
			wantCondition: true,
		},
		{
			name:          "artist equals",
			rule:          SmartRule{Field: "artist", Operator: "equals", Value: "Queen"},
			wantCondition: true,
		},
		{
			name:          "year equals",
			rule:          SmartRule{Field: "year", Operator: "equals", Value: 2024},
			wantCondition: true,
		},
		{
			name:          "year less than",
			rule:          SmartRule{Field: "year", Operator: "less_than", Value: 2000},
			wantCondition: true,
		},
		{
			name:          "unknown field",
			rule:          SmartRule{Field: "unknown", Operator: "equals", Value: "test"},
			wantCondition: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			argIndex := 1
			condition, args := service.buildRuleCondition(tt.rule, &argIndex)
			if tt.wantCondition {
				assert.NotEmpty(t, condition)
				assert.NotEmpty(t, args)
			} else {
				assert.Empty(t, condition)
			}
		})
	}
}

func TestPlaylistService_GetOrderClause(t *testing.T) {
	var mockDB *sql.DB
	mockLogger := zap.NewNop()
	service := NewPlaylistService(mockDB, mockLogger)

	tests := []struct {
		name     string
		order    string
		expected string
	}{
		{
			name:     "added descending",
			order:    "added_desc",
			expected: "mi.created_at DESC",
		},
		{
			name:     "added ascending",
			order:    "added_asc",
			expected: "mi.created_at ASC",
		},
		{
			name:     "play count descending",
			order:    "play_count_desc",
			expected: "mi.play_count DESC",
		},
		{
			name:     "rating descending",
			order:    "rating_desc",
			expected: "mi.rating DESC",
		},
		{
			name:     "random",
			order:    "random",
			expected: "RANDOM()",
		},
		{
			name:     "title ascending",
			order:    "title_asc",
			expected: "mi.title ASC",
		},
		{
			name:     "artist ascending",
			order:    "artist_asc",
			expected: "mi.artist ASC, mi.album ASC, mi.track_number ASC",
		},
		{
			name:     "unknown order defaults to created_at desc",
			order:    "unknown",
			expected: "mi.created_at DESC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.getOrderClause(tt.order)
			assert.Equal(t, tt.expected, result)
		})
	}
}
