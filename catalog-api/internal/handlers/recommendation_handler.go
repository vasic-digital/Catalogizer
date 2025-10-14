package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"catalogizer/internal/services"
)

type RecommendationHandler struct {
	recommendationService *services.RecommendationService
	deepLinkingService    *services.DeepLinkingService
}

func NewRecommendationHandler(
	recommendationService *services.RecommendationService,
	deepLinkingService *services.DeepLinkingService,
) *RecommendationHandler {
	return &RecommendationHandler{
		recommendationService: recommendationService,
		deepLinkingService:    deepLinkingService,
	}
}

// GetSimilarItems handles GET /api/v1/media/{id}/similar
func (rh *RecommendationHandler) GetSimilarItems(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	mediaID := vars["id"]

	if mediaID == "" {
		http.Error(w, "Media ID is required", http.StatusBadRequest)
		return
	}

	// Parse query parameters
	maxLocal := 10
	if maxLocalStr := r.URL.Query().Get("max_local"); maxLocalStr != "" {
		if val, err := strconv.Atoi(maxLocalStr); err == nil && val > 0 {
			maxLocal = val
		}
	}

	maxExternal := 5
	if maxExternalStr := r.URL.Query().Get("max_external"); maxExternalStr != "" {
		if val, err := strconv.Atoi(maxExternalStr); err == nil && val > 0 {
			maxExternal = val
		}
	}

	includeExternal := r.URL.Query().Get("include_external") == "true"

	similarityThreshold := 0.3
	if thresholdStr := r.URL.Query().Get("similarity_threshold"); thresholdStr != "" {
		if val, err := strconv.ParseFloat(thresholdStr, 64); err == nil && val >= 0 && val <= 1 {
			similarityThreshold = val
		}
	}

	// Parse filters
	filters := &services.RecommendationFilters{}

	if genreFilter := r.URL.Query().Get("genre"); genreFilter != "" {
		filters.GenreFilter = []string{genreFilter}
	}

	if yearStart := r.URL.Query().Get("year_start"); yearStart != "" {
		if yearEnd := r.URL.Query().Get("year_end"); yearEnd != "" {
			if start, err1 := strconv.Atoi(yearStart); err1 == nil {
				if end, err2 := strconv.Atoi(yearEnd); err2 == nil {
					filters.YearRange = &services.YearRange{
						StartYear: start,
						EndYear:   end,
					}
				}
			}
		}
	}

	if minRating := r.URL.Query().Get("min_rating"); minRating != "" {
		if maxRating := r.URL.Query().Get("max_rating"); maxRating != "" {
			if min, err1 := strconv.ParseFloat(minRating, 64); err1 == nil {
				if max, err2 := strconv.ParseFloat(maxRating, 64); err2 == nil {
					filters.RatingRange = &services.RatingRange{
						MinRating: min,
						MaxRating: max,
					}
				}
			}
		}
	}

	if language := r.URL.Query().Get("language"); language != "" {
		filters.LanguageFilter = []string{language}
	}

	filters.ExcludeWatched = r.URL.Query().Get("exclude_watched") == "true"
	filters.ExcludeOwned = r.URL.Query().Get("exclude_owned") == "true"

	if minConfidence := r.URL.Query().Get("min_confidence"); minConfidence != "" {
		if val, err := strconv.ParseFloat(minConfidence, 64); err == nil && val >= 0 && val <= 1 {
			filters.MinConfidence = val
		}
	}

	// TODO: Get actual media metadata from database/service
	// For now, we'll create a mock request
	req := &services.SimilarItemsRequest{
		MediaID:             mediaID,
		MaxLocalItems:       maxLocal,
		MaxExternalItems:    maxExternal,
		IncludeExternal:     includeExternal,
		SimilarityThreshold: similarityThreshold,
		Filters:             filters,
	}

	response, err := rh.recommendationService.GetSimilarItems(r.Context(), req)
	if err != nil {
		http.Error(w, "Failed to get similar items: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// PostSimilarItems handles POST /api/v1/media/similar
func (rh *RecommendationHandler) PostSimilarItems(w http.ResponseWriter, r *http.Request) {
	var req services.SimilarItemsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.MediaID == "" && req.MediaMetadata == nil {
		http.Error(w, "Either media_id or media_metadata is required", http.StatusBadRequest)
		return
	}

	response, err := rh.recommendationService.GetSimilarItems(r.Context(), &req)
	if err != nil {
		http.Error(w, "Failed to get similar items: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GenerateDeepLinks handles POST /api/v1/links/generate
func (rh *RecommendationHandler) GenerateDeepLinks(w http.ResponseWriter, r *http.Request) {
	var req services.DeepLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.MediaID == "" && req.MediaMetadata == nil {
		http.Error(w, "Either media_id or media_metadata is required", http.StatusBadRequest)
		return
	}

	if req.Action == "" {
		req.Action = "detail"
	}

	response, err := rh.deepLinkingService.GenerateDeepLinks(r.Context(), &req)
	if err != nil {
		http.Error(w, "Failed to generate deep links: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetMediaWithSimilarItems handles GET /api/v1/media/{id}/detail-with-similar
func (rh *RecommendationHandler) GetMediaWithSimilarItems(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	mediaID := vars["id"]

	if mediaID == "" {
		http.Error(w, "Media ID is required", http.StatusBadRequest)
		return
	}

	// Parse parameters for similar items
	maxLocal := 10
	if maxLocalStr := r.URL.Query().Get("max_similar"); maxLocalStr != "" {
		if val, err := strconv.Atoi(maxLocalStr); err == nil && val > 0 {
			maxLocal = val
		}
	}

	includeExternal := r.URL.Query().Get("include_external") == "true"

	// TODO: Get actual media metadata from database/service
	// For now, we'll create a mock response
	response := &MediaDetailWithSimilarResponse{
		MediaID:       mediaID,
		MediaMetadata: nil, // Would be populated from database
		SimilarItems:  nil, // Will be populated below
		Links:         nil, // Will be populated below
	}

	// Get similar items
	similarReq := &services.SimilarItemsRequest{
		MediaID:             mediaID,
		MaxLocalItems:       maxLocal,
		MaxExternalItems:    5,
		IncludeExternal:     includeExternal,
		SimilarityThreshold: 0.3,
	}

	similarItems, err := rh.recommendationService.GetSimilarItems(r.Context(), similarReq)
	if err == nil {
		response.SimilarItems = similarItems
	}

	// Generate deep links
	linkReq := &services.DeepLinkRequest{
		MediaID: mediaID,
		Action:  "detail",
		Context: extractLinkContext(r),
	}

	links, err := rh.deepLinkingService.GenerateDeepLinks(r.Context(), linkReq)
	if err == nil {
		response.Links = links
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// TrackLinkClick handles POST /api/v1/links/track
func (rh *RecommendationHandler) TrackLinkClick(w http.ResponseWriter, r *http.Request) {
	var event services.LinkTrackingEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	if event.TrackingID == "" {
		http.Error(w, "Tracking ID is required", http.StatusBadRequest)
		return
	}

	// Set additional fields from request
	event.UserAgent = r.Header.Get("User-Agent")
	event.IPAddress = getClientIP(r)

	err := rh.deepLinkingService.TrackLinkEvent(r.Context(), &event)
	if err != nil {
		http.Error(w, "Failed to track link event: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// GetLinkAnalytics handles GET /api/v1/links/{tracking_id}/analytics
func (rh *RecommendationHandler) GetLinkAnalytics(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	trackingID := vars["tracking_id"]

	if trackingID == "" {
		http.Error(w, "Tracking ID is required", http.StatusBadRequest)
		return
	}

	analytics, err := rh.deepLinkingService.GetLinkAnalytics(r.Context(), trackingID)
	if err != nil {
		http.Error(w, "Failed to get analytics: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(analytics)
}

// BatchGenerateLinks handles POST /api/v1/links/batch
func (rh *RecommendationHandler) BatchGenerateLinks(w http.ResponseWriter, r *http.Request) {
	var requests []*services.DeepLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&requests); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	if len(requests) == 0 {
		http.Error(w, "At least one request is required", http.StatusBadRequest)
		return
	}

	if len(requests) > 50 { // Limit batch size
		http.Error(w, "Too many requests (max 50)", http.StatusBadRequest)
		return
	}

	responses, err := rh.deepLinkingService.GenerateBatchLinks(r.Context(), requests)
	if err != nil {
		http.Error(w, "Failed to generate batch links: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"links":     responses,
		"processed": len(responses),
		"requested": len(requests),
	})
}

// GenerateSmartLink handles POST /api/v1/links/smart
func (rh *RecommendationHandler) GenerateSmartLink(w http.ResponseWriter, r *http.Request) {
	var req services.DeepLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.MediaID == "" && req.MediaMetadata == nil {
		http.Error(w, "Either media_id or media_metadata is required", http.StatusBadRequest)
		return
	}

	// Set context from request if not provided
	if req.Context == nil {
		req.Context = extractLinkContext(r)
	}

	response, err := rh.deepLinkingService.GenerateSmartLink(r.Context(), &req)
	if err != nil {
		http.Error(w, "Failed to generate smart link: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetRecommendationTrends handles GET /api/v1/recommendations/trends
func (rh *RecommendationHandler) GetRecommendationTrends(w http.ResponseWriter, r *http.Request) {
	mediaType := r.URL.Query().Get("media_type")
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "week"
	}

	limit := 20
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if val, err := strconv.Atoi(limitStr); err == nil && val > 0 && val <= 100 {
			limit = val
		}
	}

	// Mock trending recommendations
	trends := &RecommendationTrends{
		Period:    period,
		MediaType: mediaType,
		Items:     generateMockTrendingItems(mediaType, limit),
		UpdatedAt: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(trends)
}

// Helper types
type MediaDetailWithSimilarResponse struct {
	MediaID       string                         `json:"media_id"`
	MediaMetadata interface{}                    `json:"media_metadata"` // Would be proper type
	SimilarItems  *services.SimilarItemsResponse `json:"similar_items"`
	Links         *services.DeepLinkResponse     `json:"links"`
}

type RecommendationTrends struct {
	Period    string      `json:"period"`
	MediaType string      `json:"media_type,omitempty"`
	Items     []TrendItem `json:"items"`
	UpdatedAt time.Time   `json:"updated_at"`
}

type TrendItem struct {
	MediaID             string  `json:"media_id"`
	Title               string  `json:"title"`
	Subtitle            string  `json:"subtitle,omitempty"`
	CoverArt            string  `json:"cover_art,omitempty"`
	TrendScore          float64 `json:"trend_score"`
	RecommendationCount int     `json:"recommendation_count"`
	ViewCount           int     `json:"view_count"`
	Rating              float64 `json:"rating,omitempty"`
}

// Helper functions
func extractLinkContext(r *http.Request) *services.LinkContext {
	context := &services.LinkContext{
		ReferrerPage: r.Header.Get("Referer"),
	}

	// Extract user context from headers or query params
	if userID := r.Header.Get("X-User-ID"); userID != "" {
		context.UserID = userID
	}
	if deviceID := r.Header.Get("X-Device-ID"); deviceID != "" {
		context.DeviceID = deviceID
	}
	if sessionID := r.Header.Get("X-Session-ID"); sessionID != "" {
		context.SessionID = sessionID
	}

	// Determine platform from User-Agent
	userAgent := strings.ToLower(r.Header.Get("User-Agent"))
	switch {
	case strings.Contains(userAgent, "android"):
		context.Platform = "android"
	case strings.Contains(userAgent, "iphone") || strings.Contains(userAgent, "ipad"):
		context.Platform = "ios"
	case strings.Contains(userAgent, "catalogizer-desktop"):
		context.Platform = "desktop"
	default:
		context.Platform = "web"
	}

	// Extract UTM parameters
	query := r.URL.Query()
	if query.Get("utm_source") != "" {
		context.UTMParams = &services.UTMParameters{
			Source:   query.Get("utm_source"),
			Medium:   query.Get("utm_medium"),
			Campaign: query.Get("utm_campaign"),
			Term:     query.Get("utm_term"),
			Content:  query.Get("utm_content"),
		}
	}

	return context
}

func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		return strings.Split(forwarded, ",")[0]
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fallback to RemoteAddr
	return r.RemoteAddr
}

func generateMockTrendingItems(mediaType string, limit int) []TrendItem {
	items := make([]TrendItem, 0, limit)

	for i := 0; i < limit; i++ {
		item := TrendItem{
			MediaID:             fmt.Sprintf("trending_%s_%d", mediaType, i),
			TrendScore:          0.9 - float64(i)*0.02,
			RecommendationCount: 100 - i*5,
			ViewCount:           1000 - i*50,
			Rating:              8.5 - float64(i)*0.1,
		}

		switch mediaType {
		case "video":
			item.Title = fmt.Sprintf("Trending Movie %d", i+1)
			item.Subtitle = "Action/Adventure"
		case "audio":
			item.Title = fmt.Sprintf("Trending Song %d", i+1)
			item.Subtitle = "Popular Artist"
		case "book":
			item.Title = fmt.Sprintf("Trending Book %d", i+1)
			item.Subtitle = "Bestselling Author"
		default:
			item.Title = fmt.Sprintf("Trending Item %d", i+1)
		}

		items = append(items, item)
	}

	return items
}
