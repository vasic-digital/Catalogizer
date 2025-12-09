package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"catalogizer/config"
	"catalogizer/database"
	"catalogizer/handlers"
	root_handlers "catalogizer/handlers"
	"catalogizer/internal/middleware"
	"catalogizer/internal/services"
	root_repository "catalogizer/repository"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type RecommendationTestSuite struct {
	suite.Suite
	router               *gin.Engine
	recommendationHandler *handlers.RecommendationHandler
	simpleRecHandler     *root_handlers.SimpleRecommendationHandler
	db                   *database.DB
}

func (suite *RecommendationTestSuite) TearDownSuite() {
	if suite.db != nil {
		suite.db.Close()
	}
}

func (suite *RecommendationTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	
	// Create a test database
	dbConfig := &config.DatabaseConfig{
		Path: ":memory:",
	}
	var err error
	db, err := database.NewConnection(dbConfig)
	suite.Require().NoError(err)
	// Don't close the database here - it will be closed in TearDownSuite
	suite.db = db
	
	// Create test logger
	logger, _ := zap.NewDevelopment()
	
	// Initialize services
	mediaRecognitionService := services.NewMediaRecognitionService(db.DB, logger, nil, nil, "", "", "", "", "", "")
	duplicateDetectionService := services.NewDuplicateDetectionService(db.DB, logger, nil)
	fileRepository := root_repository.NewFileRepository(db)
	recommendationService := services.NewRecommendationService(
		mediaRecognitionService,
		duplicateDetectionService,
		fileRepository,
		db.DB,
	)
	
	// Create handlers
	suite.recommendationHandler = handlers.NewRecommendationHandler(recommendationService)
	suite.simpleRecHandler = root_handlers.NewSimpleRecommendationHandler()
	
	// Setup router
	suite.router = gin.New()
	suite.router.Use(middleware.Logger(logger))
	
	// Add routes
	api := suite.router.Group("/api/v1")
	
	// Simple test routes
	api.GET("/recommendations/test", suite.simpleRecHandler.GetSimpleRecommendation)
	api.GET("/recommendations/error", suite.simpleRecHandler.GetTest)
	
	// Full recommendation routes
	recGroup := api.Group("/recommendations")
	recGroup.GET("/similar/:media_id", suite.recommendationHandler.GetSimilarItems)
	recGroup.GET("/trending", suite.recommendationHandler.GetTrendingItems)
	recGroup.GET("/personalized/:user_id", suite.recommendationHandler.GetPersonalizedRecommendations)
}

func (suite *RecommendationTestSuite) TestSimpleRecommendationEndpoint() {
	req, _ := http.NewRequest("GET", "/api/v1/recommendations/test", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Simple recommendation works!", response["message"])
}

func (suite *RecommendationTestSuite) TestSimilarItemsEndpoint() {
	req, _ := http.NewRequest("GET", "/api/v1/recommendations/similar/123", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	
	// Should return 200 even if no similar items found
	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

func (suite *RecommendationTestSuite) TestTrendingItemsEndpoint() {
	req, _ := http.NewRequest("GET", "/api/v1/recommendations/trending", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	
	// Should return 200 even if no trending items found
	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

func (suite *RecommendationTestSuite) TestPersonalizedRecommendationsEndpoint() {
	req, _ := http.NewRequest("GET", "/api/v1/recommendations/personalized/456", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	
	// Should return 200 even if no recommendations found
	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

func TestRecommendationSuite(t *testing.T) {
	suite.Run(t, new(RecommendationTestSuite))
}