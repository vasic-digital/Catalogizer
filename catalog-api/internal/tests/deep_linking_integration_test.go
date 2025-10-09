package tests

import (
	"context"
	"strconv"
	"strings"
	"testing"
)

// MockDeepLinkingService provides a mock implementation for testing
type MockDeepLinkingService struct {
	baseURLs map[string]string
}

type DeepLinkRequest struct {
	MediaID     int64             `json:"media_id"`
	Platform    string            `json:"platform"`
	Context     string            `json:"context"`
	UTMSource   string            `json:"utm_source,omitempty"`
	UTMMedium   string            `json:"utm_medium,omitempty"`
	UTMCampaign string            `json:"utm_campaign,omitempty"`
	CustomData  map[string]string `json:"custom_data,omitempty"`
}

type DeepLinkResponse struct {
	WebURL      string            `json:"web_url"`
	AndroidURL  string            `json:"android_url"`
	IOSURL      string            `json:"ios_url"`
	DesktopURL  string            `json:"desktop_url"`
	SmartURL    string            `json:"smart_url"`
	QRCodeURL   string            `json:"qr_code_url,omitempty"`
	Analytics   map[string]string `json:"analytics,omitempty"`
}

func NewMockDeepLinkingService() *MockDeepLinkingService {
	return &MockDeepLinkingService{
		baseURLs: map[string]string{
			"web":     "https://catalogizer.app",
			"android": "catalogizer://",
			"ios":     "catalogizer://",
			"desktop": "catalogizer://",
		},
	}
}

func (m *MockDeepLinkingService) GenerateDeepLinks(ctx context.Context, req *DeepLinkRequest) (*DeepLinkResponse, error) {
	response := &DeepLinkResponse{
		WebURL:     m.generateWebURL(req),
		AndroidURL: m.generateAndroidURL(req),
		IOSURL:     m.generateIOSURL(req),
		DesktopURL: m.generateDesktopURL(req),
		SmartURL:   m.generateSmartURL(req),
		Analytics:  make(map[string]string),
	}

	// Add UTM parameters to analytics
	if req.UTMSource != "" {
		response.Analytics["utm_source"] = req.UTMSource
	}
	if req.UTMMedium != "" {
		response.Analytics["utm_medium"] = req.UTMMedium
	}
	if req.UTMCampaign != "" {
		response.Analytics["utm_campaign"] = req.UTMCampaign
	}

	// Add QR code URL if requested
	if req.CustomData != nil && req.CustomData["include_qr"] == "true" {
		response.QRCodeURL = m.generateQRCodeURL(req)
	}

	return response, nil
}

func (m *MockDeepLinkingService) generateWebURL(req *DeepLinkRequest) string {
	url := m.baseURLs["web"] + "/item/" + strconv.FormatInt(req.MediaID, 10)
	return m.addUTMParameters(url, req)
}

func (m *MockDeepLinkingService) generateAndroidURL(req *DeepLinkRequest) string {
	url := m.baseURLs["android"] + "item/" + strconv.FormatInt(req.MediaID, 10)
	return m.addUTMParameters(url, req)
}

func (m *MockDeepLinkingService) generateIOSURL(req *DeepLinkRequest) string {
	url := m.baseURLs["ios"] + "item/" + strconv.FormatInt(req.MediaID, 10)
	return m.addUTMParameters(url, req)
}

func (m *MockDeepLinkingService) generateDesktopURL(req *DeepLinkRequest) string {
	url := m.baseURLs["desktop"] + "item/" + strconv.FormatInt(req.MediaID, 10)
	return m.addUTMParameters(url, req)
}

func (m *MockDeepLinkingService) generateSmartURL(req *DeepLinkRequest) string {
	// Smart URL that redirects based on platform
	return m.baseURLs["web"] + "/smart/" + strconv.FormatInt(req.MediaID, 10)
}

func (m *MockDeepLinkingService) generateQRCodeURL(req *DeepLinkRequest) string {
	return m.baseURLs["web"] + "/qr/" + strconv.FormatInt(req.MediaID, 10) + ".png"
}

func (m *MockDeepLinkingService) addUTMParameters(url string, req *DeepLinkRequest) string {
	if req.UTMSource == "" && req.UTMMedium == "" && req.UTMCampaign == "" {
		return url
	}

	separator := "?"
	if req.UTMSource != "" {
		url += separator + "utm_source=" + req.UTMSource
		separator = "&"
	}
	if req.UTMMedium != "" {
		url += separator + "utm_medium=" + req.UTMMedium
		separator = "&"
	}
	if req.UTMCampaign != "" {
		url += separator + "utm_campaign=" + req.UTMCampaign
	}

	return url
}

// Test functions

func TestDeepLinkingBasicGeneration(t *testing.T) {
	service := NewMockDeepLinkingService()
	ctx := context.Background()

	req := &DeepLinkRequest{
		MediaID:  123,
		Platform: "web",
		Context:  "detail_screen",
	}

	response, err := service.GenerateDeepLinks(ctx, req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if response == nil {
		t.Fatal("Expected response, got nil")
	}

	// Verify all platform URLs are generated
	if response.WebURL == "" {
		t.Error("Expected WebURL to be generated")
	}
	if response.AndroidURL == "" {
		t.Error("Expected AndroidURL to be generated")
	}
	if response.IOSURL == "" {
		t.Error("Expected IOSURL to be generated")
	}
	if response.DesktopURL == "" {
		t.Error("Expected DesktopURL to be generated")
	}
	if response.SmartURL == "" {
		t.Error("Expected SmartURL to be generated")
	}
}

func TestDeepLinkingUTMParameters(t *testing.T) {
	service := NewMockDeepLinkingService()
	ctx := context.Background()

	req := &DeepLinkRequest{
		MediaID:     123,
		Platform:    "web",
		UTMSource:   "email",
		UTMMedium:   "newsletter",
		UTMCampaign: "winter_2024",
	}

	response, err := service.GenerateDeepLinks(ctx, req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify UTM parameters are included in URLs
	expectedParams := []string{"utm_source=email", "utm_medium=newsletter", "utm_campaign=winter_2024"}

	for _, param := range expectedParams {
		if !containsString(response.WebURL, param) {
			t.Errorf("Expected WebURL to contain %s, got %s", param, response.WebURL)
		}
		if !containsString(response.AndroidURL, param) {
			t.Errorf("Expected AndroidURL to contain %s, got %s", param, response.AndroidURL)
		}
	}

	// Verify analytics tracking
	if response.Analytics == nil {
		t.Error("Expected analytics to be populated")
	}
	if response.Analytics["utm_source"] != "email" {
		t.Error("Expected utm_source in analytics")
	}
}

func TestDeepLinkingQRCodeGeneration(t *testing.T) {
	service := NewMockDeepLinkingService()
	ctx := context.Background()

	req := &DeepLinkRequest{
		MediaID:  123,
		Platform: "web",
		CustomData: map[string]string{
			"include_qr": "true",
		},
	}

	response, err := service.GenerateDeepLinks(ctx, req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if response.QRCodeURL == "" {
		t.Error("Expected QR code URL to be generated when requested")
	}

	if !containsString(response.QRCodeURL, ".png") {
		t.Error("Expected QR code URL to have .png extension")
	}
}

func TestDeepLinkingPlatformSpecificity(t *testing.T) {
	service := NewMockDeepLinkingService()
	ctx := context.Background()

	req := &DeepLinkRequest{
		MediaID:  123,
		Platform: "android",
	}

	response, err := service.GenerateDeepLinks(ctx, req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify platform-specific URL formats
	if !containsString(response.WebURL, "https://") {
		t.Error("Expected web URL to use HTTPS protocol")
	}

	if !containsString(response.AndroidURL, "catalogizer://") {
		t.Error("Expected Android URL to use custom scheme")
	}

	if !containsString(response.IOSURL, "catalogizer://") {
		t.Error("Expected iOS URL to use custom scheme")
	}
}

func TestDeepLinkingContextualLinks(t *testing.T) {
	service := NewMockDeepLinkingService()
	ctx := context.Background()

	contexts := []string{"detail_screen", "search_results", "recommendations", "share"}

	for _, context := range contexts {
		req := &DeepLinkRequest{
			MediaID:  123,
			Platform: "web",
			Context:  context,
		}

		response, err := service.GenerateDeepLinks(ctx, req)
		if err != nil {
			t.Fatalf("Expected no error for context %s, got %v", context, err)
		}

		if response == nil {
			t.Fatalf("Expected response for context %s, got nil", context)
		}

		// Verify that links are generated regardless of context
		if response.WebURL == "" {
			t.Errorf("Expected WebURL for context %s", context)
		}
	}
}

func TestDeepLinkingPerformance(t *testing.T) {
	service := NewMockDeepLinkingService()
	ctx := context.Background()

	// Test multiple concurrent requests
	const numRequests = 50
	results := make(chan error, numRequests)

	for i := 0; i < numRequests; i++ {
		go func(id int) {
			req := &DeepLinkRequest{
				MediaID:  int64(id + 1),
				Platform: "web",
			}

			_, err := service.GenerateDeepLinks(ctx, req)
			results <- err
		}(i)
	}

	// Collect results
	for i := 0; i < numRequests; i++ {
		if err := <-results; err != nil {
			t.Errorf("Request %d failed: %v", i, err)
		}
	}
}

func TestDeepLinkingValidation(t *testing.T) {
	service := NewMockDeepLinkingService()
	ctx := context.Background()

	// Test with various invalid inputs
	testCases := []struct {
		name    string
		request *DeepLinkRequest
	}{
		{
			name: "zero media ID",
			request: &DeepLinkRequest{
				MediaID:  0,
				Platform: "web",
			},
		},
		{
			name: "empty platform",
			request: &DeepLinkRequest{
				MediaID:  123,
				Platform: "",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			response, err := service.GenerateDeepLinks(ctx, tc.request)
			if err != nil {
				t.Fatalf("Expected no error for %s, got %v", tc.name, err)
			}

			// Service should still generate links even with edge cases
			if response == nil {
				t.Errorf("Expected response for %s, got nil", tc.name)
			}
		})
	}
}

// Helper functions
func containsString(s, substr string) bool {
	return strings.Contains(s, substr)
}