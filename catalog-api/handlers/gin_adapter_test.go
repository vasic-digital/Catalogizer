package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type GinAdapterTestSuite struct {
	suite.Suite
}

func (suite *GinAdapterTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
}

func (suite *GinAdapterTestSuite) TestWrapHTTPHandler_BasicHandler() {
	stdHandler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "hello"})
	}

	ginHandler := WrapHTTPHandler(stdHandler)
	assert.NotNil(suite.T(), ginHandler)

	router := gin.New()
	router.GET("/test", ginHandler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	var resp map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "hello", resp["message"])
}

func (suite *GinAdapterTestSuite) TestWrapHTTPHandler_ErrorHandler() {
	stdHandler := func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}

	ginHandler := WrapHTTPHandler(stdHandler)

	router := gin.New()
	router.GET("/error", ginHandler)

	req := httptest.NewRequest("GET", "/error", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)
	assert.Contains(suite.T(), w.Body.String(), "internal error")
}

func (suite *GinAdapterTestSuite) TestWrapHTTPHandler_PreservesHeaders() {
	stdHandler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Custom-Header", "test-value")
		w.WriteHeader(http.StatusOK)
	}

	ginHandler := WrapHTTPHandler(stdHandler)

	router := gin.New()
	router.GET("/headers", ginHandler)

	req := httptest.NewRequest("GET", "/headers", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	assert.Equal(suite.T(), "test-value", w.Header().Get("X-Custom-Header"))
}

func (suite *GinAdapterTestSuite) TestWrapHTTPHandler_ReadsRequestBody() {
	stdHandler := func(w http.ResponseWriter, r *http.Request) {
		// Verify the request method is preserved
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"method": r.Method})
	}

	ginHandler := WrapHTTPHandler(stdHandler)

	router := gin.New()
	router.POST("/method", ginHandler)

	req := httptest.NewRequest("POST", "/method", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	var resp map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "POST", resp["method"])
}

func (suite *GinAdapterTestSuite) TestWrapHTTPHandler_PreservesQueryParams() {
	stdHandler := func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"name": name})
	}

	ginHandler := WrapHTTPHandler(stdHandler)

	router := gin.New()
	router.GET("/query", ginHandler)

	req := httptest.NewRequest("GET", "/query?name=test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	var resp map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "test", resp["name"])
}

func (suite *GinAdapterTestSuite) TestWrapHTTPHandler_NoContent() {
	stdHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}

	ginHandler := WrapHTTPHandler(stdHandler)

	router := gin.New()
	router.DELETE("/delete", ginHandler)

	req := httptest.NewRequest("DELETE", "/delete", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNoContent, w.Code)
	assert.Empty(suite.T(), w.Body.String())
}

func (suite *GinAdapterTestSuite) TestWrapHTTPHandler_RequestHeaders() {
	stdHandler := func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"auth": auth})
	}

	ginHandler := WrapHTTPHandler(stdHandler)

	router := gin.New()
	router.GET("/auth", ginHandler)

	req := httptest.NewRequest("GET", "/auth", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	var resp map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Bearer test-token", resp["auth"])
}

func (suite *GinAdapterTestSuite) TestWrapHTTPHandler_StatusCreated() {
	stdHandler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"id": "123"})
	}

	ginHandler := WrapHTTPHandler(stdHandler)

	router := gin.New()
	router.POST("/create", ginHandler)

	req := httptest.NewRequest("POST", "/create", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)
}

func TestGinAdapterTestSuite(t *testing.T) {
	suite.Run(t, new(GinAdapterTestSuite))
}
