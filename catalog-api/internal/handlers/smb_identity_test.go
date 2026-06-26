package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"catalogizer/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// ---------------------------------------------------------------------------
// ListIdentities tests
// ---------------------------------------------------------------------------

func TestSMBListIdentities_Empty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	handler := NewSMBDiscoveryHandler(nil, logger)

	router := gin.New()
	router.GET("/api/v1/smb/identities", handler.ListIdentities)

	// No env vars set → empty slice, 200 OK
	req := httptest.NewRequest("GET", "/api/v1/smb/identities", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var got []ListIdentityInfo
	err := json.Unmarshal(w.Body.Bytes(), &got)
	require.NoError(t, err)
	assert.Empty(t, got, "expected no identities when env is not set")
}

func TestSMBListIdentities_ReturnsCredentialsOnly(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	handler := NewSMBDiscoveryHandler(nil, logger)

	router := gin.New()
	router.GET("/api/v1/smb/identities", handler.ListIdentities)

	// Set two credential identities; one should be filtered out by TYPE.
	// Identity 1: credentials (should appear)
	os.Setenv("CATALOGIZER_IDENTITY_COUNT", "3")
	os.Setenv("CATALOGIZER_IDENTITY_1_TYPE", "credentials")
	os.Setenv("CATALOGIZER_IDENTITY_1_USERNAME", "alice")
	os.Setenv("CATALOGIZER_IDENTITY_1_PASSWORD", "secret-alice")
	os.Setenv("CATALOGIZER_IDENTITY_1_DOMAIN", "WORKGROUP")
	// Identity 2: api_token (should be filtered out — not SMB NTLM)
	os.Setenv("CATALOGIZER_IDENTITY_2_TYPE", "api_token")
	os.Setenv("CATALOGIZER_IDENTITY_2_USERNAME", "bob")
	// Identity 3: credentials (should appear)
	os.Setenv("CATALOGIZER_IDENTITY_3_TYPE", "credentials")
	os.Setenv("CATALOGIZER_IDENTITY_3_USERNAME", "carol")
	os.Setenv("CATALOGIZER_IDENTITY_3_PASSWORD", "secret-carol")

	defer func() {
		os.Unsetenv("CATALOGIZER_IDENTITY_COUNT")
		os.Unsetenv("CATALOGIZER_IDENTITY_1_TYPE")
		os.Unsetenv("CATALOGIZER_IDENTITY_1_USERNAME")
		os.Unsetenv("CATALOGIZER_IDENTITY_1_PASSWORD")
		os.Unsetenv("CATALOGIZER_IDENTITY_1_DOMAIN")
		os.Unsetenv("CATALOGIZER_IDENTITY_2_TYPE")
		os.Unsetenv("CATALOGIZER_IDENTITY_2_USERNAME")
		os.Unsetenv("CATALOGIZER_IDENTITY_3_TYPE")
		os.Unsetenv("CATALOGIZER_IDENTITY_3_USERNAME")
		os.Unsetenv("CATALOGIZER_IDENTITY_3_PASSWORD")
	}()

	req := httptest.NewRequest("GET", "/api/v1/smb/identities", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var got []ListIdentityInfo
	err := json.Unmarshal(w.Body.Bytes(), &got)
	require.NoError(t, err)

	// Only credential identities should be returned (indices 1 and 3), NOT 2.
	require.Len(t, got, 2, "expected exactly 2 credential identities")

	assert.Equal(t, 1, got[0].Index)
	assert.Equal(t, "alice", got[0].Username)
	assert.Equal(t, "alice", got[0].Label, "label should be the username, never a secret")
	assert.Equal(t, "credentials", got[0].Kind)

	assert.Equal(t, 3, got[1].Index)
	assert.Equal(t, "carol", got[1].Username)
	assert.Equal(t, "carol", got[1].Label)
	assert.Equal(t, "credentials", got[1].Kind)
}

func TestSMBListIdentities_NeverExposesPassword(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	handler := NewSMBDiscoveryHandler(nil, logger)

	router := gin.New()
	router.GET("/api/v1/smb/identities", handler.ListIdentities)

	os.Setenv("CATALOGIZER_IDENTITY_COUNT", "1")
	os.Setenv("CATALOGIZER_IDENTITY_1_TYPE", "credentials")
	os.Setenv("CATALOGIZER_IDENTITY_1_USERNAME", "dboy")
	os.Setenv("CATALOGIZER_IDENTITY_1_PASSWORD", "supersecret")
	defer func() {
		os.Unsetenv("CATALOGIZER_IDENTITY_COUNT")
		os.Unsetenv("CATALOGIZER_IDENTITY_1_TYPE")
		os.Unsetenv("CATALOGIZER_IDENTITY_1_USERNAME")
		os.Unsetenv("CATALOGIZER_IDENTITY_1_PASSWORD")
	}()

	req := httptest.NewRequest("GET", "/api/v1/smb/identities", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	bodyStr := w.Body.String()
	// The password "supersecret" MUST NOT appear anywhere in the JSON response.
	assert.NotContains(t, bodyStr, "supersecret",
		"ListIdentities MUST NOT expose password values (§11.4.10)")
	// The literal "password" key MUST NOT appear.
	assert.NotContains(t, bodyStr, "password",
		"JSON response must not contain a 'password' key")
}

func TestSMBListIdentities_GuestIsNotIncluded(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	handler := NewSMBDiscoveryHandler(nil, logger)

	router := gin.New()
	router.GET("/api/v1/smb/identities", handler.ListIdentities)

	// LoadSMBIdentitiesFromEnv returns ONLY credential-type identities — the
	// implicit guest identity is not in the env config.
	os.Setenv("CATALOGIZER_IDENTITY_COUNT", "1")
	os.Setenv("CATALOGIZER_IDENTITY_1_TYPE", "credentials")
	os.Setenv("CATALOGIZER_IDENTITY_1_USERNAME", "testuser")
	os.Setenv("CATALOGIZER_IDENTITY_1_PASSWORD", "testpass")
	defer func() {
		os.Unsetenv("CATALOGIZER_IDENTITY_COUNT")
		os.Unsetenv("CATALOGIZER_IDENTITY_1_TYPE")
		os.Unsetenv("CATALOGIZER_IDENTITY_1_USERNAME")
		os.Unsetenv("CATALOGIZER_IDENTITY_1_PASSWORD")
	}()

	req := httptest.NewRequest("GET", "/api/v1/smb/identities", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var got []ListIdentityInfo
	err := json.Unmarshal(w.Body.Bytes(), &got)
	require.NoError(t, err)

	require.Len(t, got, 1, "guest should NOT appear as a listed identity")
	assert.Equal(t, "testuser", got[0].Username)
}

// ---------------------------------------------------------------------------
// ProbeHost tests
// ---------------------------------------------------------------------------

func TestSMBProbeHost_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	svc := services.NewSMBDiscoveryService(logger)
	handler := NewSMBDiscoveryHandler(svc, logger)

	router := gin.New()
	router.POST("/api/v1/smb/probe", handler.ProbeHost)

	req := httptest.NewRequest("POST", "/api/v1/smb/probe",
		bytes.NewBufferString("not-json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var body map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &body)
	require.NoError(t, err)
	assert.Contains(t, body["error"], "Invalid request")
}

func TestSMBProbeHost_MissingHost(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	svc := services.NewSMBDiscoveryService(logger)
	handler := NewSMBDiscoveryHandler(svc, logger)

	router := gin.New()
	router.POST("/api/v1/smb/probe", handler.ProbeHost)

	body := `{}`
	req := httptest.NewRequest("POST", "/api/v1/smb/probe",
		bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Contains(t, resp["error"].(string), "Host")
}

func TestSMBProbeHost_ProbesUnreachableHost(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	svc := services.NewSMBDiscoveryService(logger)
	handler := NewSMBDiscoveryHandler(svc, logger)

	router := gin.New()
	router.POST("/api/v1/smb/probe", handler.ProbeHost)

	body := `{"host": "192.0.2.1"}`
	req := httptest.NewRequest("POST", "/api/v1/smb/probe",
		bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code,
		"an unreachable host should produce a 500, not a false 200")

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Contains(t, resp["error"].(string), "Failed to probe host")
	// The result object should still be in the response even on error.
	assert.NotNil(t, resp["result"])
}

func TestSMBProbeHost_NoIdentitiesStillProbesGuest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	svc := services.NewSMBDiscoveryService(logger)
	handler := NewSMBDiscoveryHandler(svc, logger)

	router := gin.New()
	router.POST("/api/v1/smb/probe", handler.ProbeHost)

	// No identities configured → still tries guest (which will fail on a
	// nonexistent host, but the handler must accept the request and respond
	// faithfully).
	body := `{"host": "10.255.255.1"}`
	req := httptest.NewRequest("POST", "/api/v1/smb/probe",
		bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// We expect a 500 (network error) but the result payload must be present
	// with Authenticated=false.
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Contains(t, resp["error"].(string), "Failed to probe host")

	// Parse the embedded result object.
	resultRaw, ok := resp["result"].(map[string]interface{})
	require.True(t, ok, "result should be a JSON object")
	assert.False(t, resultRaw["authenticated"].(bool),
		"no identity should authenticate against an unreachable host")
	assert.Equal(t, "10.255.255.1", resultRaw["host"])
}

func TestSMBProbeHost_ResponseShape(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zap.NewNop()
	svc := services.NewSMBDiscoveryService(logger)
	handler := NewSMBDiscoveryHandler(svc, logger)

	router := gin.New()
	router.POST("/api/v1/smb/probe", handler.ProbeHost)

	// Set one identity so ProbeHostWithIdentities has it to attempt.
	os.Setenv("CATALOGIZER_IDENTITY_COUNT", "1")
	os.Setenv("CATALOGIZER_IDENTITY_1_TYPE", "credentials")
	os.Setenv("CATALOGIZER_IDENTITY_1_USERNAME", "prober")
	os.Setenv("CATALOGIZER_IDENTITY_1_PASSWORD", "proberpass")
	defer func() {
		os.Unsetenv("CATALOGIZER_IDENTITY_COUNT")
		os.Unsetenv("CATALOGIZER_IDENTITY_1_TYPE")
		os.Unsetenv("CATALOGIZER_IDENTITY_1_USERNAME")
		os.Unsetenv("CATALOGIZER_IDENTITY_1_PASSWORD")
	}()

	body := `{"host": "10.0.0.99"}`
	req := httptest.NewRequest("POST", "/api/v1/smb/probe",
		bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	// Verify the result shape: host, authenticated (false), and the zero-value
	// identity fields when no identity bound (identity_index=0, label="").
	resultRaw, ok := resp["result"].(map[string]interface{})
	require.True(t, ok, "result must be a JSON object")
	assert.Equal(t, "10.0.0.99", resultRaw["host"])
	assert.False(t, resultRaw["authenticated"].(bool))
	assert.Equal(t, 0, int(resultRaw["identity_index"].(float64)),
		"when no identity authenticated, identity_index is the Go zero-value 0")
	assert.Equal(t, "", resultRaw["identity_label"],
		"when no identity authenticated, identity_label is the Go zero-value empty string")
	// shares should be nil/empty when auth failed.
	shares, ok := resultRaw["shares"]
	assert.True(t, ok)
	assert.Nil(t, shares)
}
