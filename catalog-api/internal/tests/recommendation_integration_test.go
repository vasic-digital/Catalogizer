package tests

import (
	"context"
	"testing"
	"time"

	"catalog-api/internal/models"
)

// MockRecommendationService provides a mock implementation for testing
type MockRecommendationService struct {
	mockLocalItems    []models.MediaMetadata
	mockExternalItems []ExternalRecommendation
}

type ExternalRecommendation struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	URL         string  `json:"url"`
	Score       float64 `json:"score"`
	Provider    string  `json:"provider"`
}

type RecommendationRequest struct {
	MediaID       int64    `json:"media_id"`
	MediaMetadata *models.MediaMetadata `json:"media_metadata"`
	MaxResults    int      `json:"max_results"`
	Filters       map[string]interface{} `json:"filters"`
}

type RecommendationResponse struct {
	LocalItems    []models.MediaMetadata     `json:"local_items"`
	ExternalItems []ExternalRecommendation   `json:"external_items"`
	TotalFound    int                       `json:"total_found"`
	ProcessingTime time.Duration            `json:"processing_time"`
}

func NewMockRecommendationService() *MockRecommendationService {
	return &MockRecommendationService{
		mockLocalItems: []models.MediaMetadata{
			{
				ID:          1,
				Title:       "Similar Movie 1",
				Description: "A similar movie",
				Genre:       "Action",
				Year:        intPtr(2023),
				Rating:      floatPtr(8.5),
			},
			{
				ID:          2,
				Title:       "Similar Movie 2",
				Description: "Another similar movie",
				Genre:       "Action",
				Year:        intPtr(2022),
				Rating:      floatPtr(8.0),
			},
		},
		mockExternalItems: []ExternalRecommendation{
			{
				ID:          "ext_1",
				Title:       "External Recommendation 1",
				Description: "An external recommendation",
				URL:         "https://example.com/movie/1",
				Score:       0.95,
				Provider:    "TMDB",
			},
			{
				ID:          "ext_2",
				Title:       "External Recommendation 2",
				Description: "Another external recommendation",
				URL:         "https://example.com/movie/2",
				Score:       0.90,
				Provider:    "IMDB",
			},
		},
	}
}

func (m *MockRecommendationService) GetSimilarItems(ctx context.Context, req *RecommendationRequest) (*RecommendationResponse, error) {
	startTime := time.Now()

	// Simulate processing delay
	time.Sleep(10 * time.Millisecond)

	// Apply filters to local items
	filteredLocal := m.mockLocalItems
	if genre, ok := req.Filters["genre"].(string); ok && genre != "" {
		var filtered []models.MediaMetadata
		for _, item := range m.mockLocalItems {
			if item.Genre == genre {
				filtered = append(filtered, item)
			}
		}
		filteredLocal = filtered
	}

	// Limit results
	maxResults := req.MaxResults
	if maxResults <= 0 {
		maxResults = 10
	}

	if len(filteredLocal) > maxResults {
		filteredLocal = filteredLocal[:maxResults]
	}

	externalItems := m.mockExternalItems
	if len(externalItems) > maxResults {
		externalItems = externalItems[:maxResults]
	}

	return &RecommendationResponse{
		LocalItems:     filteredLocal,
		ExternalItems:  externalItems,
		TotalFound:     len(filteredLocal) + len(externalItems),
		ProcessingTime: time.Since(startTime),
	}, nil
}

// Test functions

func TestRecommendationServiceBasicFunctionality(t *testing.T) {
	service := NewMockRecommendationService()
	ctx := context.Background()

	req := &RecommendationRequest{
		MediaID: 123,
		MediaMetadata: &models.MediaMetadata{
			ID:    123,
			Title: "Test Movie",
			Genre: "Action",
			Year:  intPtr(2023),
		},
		MaxResults: 5,
		Filters:    map[string]interface{}{},
	}

	response, err := service.GetSimilarItems(ctx, req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if response == nil {
		t.Fatal("Expected response, got nil")
	}

	if len(response.LocalItems) == 0 {
		t.Error("Expected local items, got none")
	}

	if len(response.ExternalItems) == 0 {
		t.Error("Expected external items, got none")
	}

	if response.TotalFound != len(response.LocalItems)+len(response.ExternalItems) {
		t.Errorf("Expected total found to match sum of items, got %d", response.TotalFound)
	}

	if response.ProcessingTime <= 0 {
		t.Error("Expected processing time > 0")
	}
}

func TestRecommendationServiceFiltering(t *testing.T) {
	service := NewMockRecommendationService()
	ctx := context.Background()

	req := &RecommendationRequest{
		MediaID: 123,
		MediaMetadata: &models.MediaMetadata{
			ID:    123,
			Title: "Test Movie",
			Genre: "Action",
		},
		MaxResults: 10,
		Filters: map[string]interface{}{
			"genre": "Action",
		},
	}

	response, err := service.GetSimilarItems(ctx, req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// All local items should match the genre filter
	for _, item := range response.LocalItems {
		if item.Genre != "Action" {
			t.Errorf("Expected genre 'Action', got '%s'", item.Genre)
		}
	}
}

func TestRecommendationServiceMaxResults(t *testing.T) {
	service := NewMockRecommendationService()
	ctx := context.Background()

	req := &RecommendationRequest{
		MediaID:    123,
		MaxResults: 1,
		Filters:    map[string]interface{}{},
	}

	response, err := service.GetSimilarItems(ctx, req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(response.LocalItems) > 1 {
		t.Errorf("Expected max 1 local item, got %d", len(response.LocalItems))
	}

	if len(response.ExternalItems) > 1 {
		t.Errorf("Expected max 1 external item, got %d", len(response.ExternalItems))
	}
}

func TestRecommendationServicePerformance(t *testing.T) {
	service := NewMockRecommendationService()
	ctx := context.Background()

	// Test multiple concurrent requests
	const numRequests = 10
	results := make(chan error, numRequests)

	for i := 0; i < numRequests; i++ {
		go func() {
			req := &RecommendationRequest{
				MediaID:    123,
				MaxResults: 5,
				Filters:    map[string]interface{}{},
			}

			_, err := service.GetSimilarItems(ctx, req)
			results <- err
		}()
	}

	// Collect results
	for i := 0; i < numRequests; i++ {
		if err := <-results; err != nil {
			t.Errorf("Request %d failed: %v", i, err)
		}
	}
}

func TestRecommendationResponseStructure(t *testing.T) {
	service := NewMockRecommendationService()
	ctx := context.Background()

	req := &RecommendationRequest{
		MediaID:    123,
		MaxResults: 10,
		Filters:    map[string]interface{}{},
	}

	response, err := service.GetSimilarItems(ctx, req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify response structure
	if response.LocalItems == nil {
		t.Error("Expected LocalItems to be initialized")
	}

	if response.ExternalItems == nil {
		t.Error("Expected ExternalItems to be initialized")
	}

	// Verify local items have required fields
	for i, item := range response.LocalItems {
		if item.Title == "" {
			t.Errorf("Local item %d missing title", i)
		}
		if item.ID == 0 {
			t.Errorf("Local item %d missing ID", i)
		}
	}

	// Verify external items have required fields
	for i, item := range response.ExternalItems {
		if item.Title == "" {
			t.Errorf("External item %d missing title", i)
		}
		if item.URL == "" {
			t.Errorf("External item %d missing URL", i)
		}
		if item.Provider == "" {
			t.Errorf("External item %d missing provider", i)
		}
		if item.Score <= 0 || item.Score > 1 {
			t.Errorf("External item %d has invalid score: %f", i, item.Score)
		}
	}
}

// Helper functions
func intPtr(i int) *int {
	return &i
}

func floatPtr(f float64) *float64 {
	return &f
}