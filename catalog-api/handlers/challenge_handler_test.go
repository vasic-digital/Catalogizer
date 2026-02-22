package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"catalogizer/services"
	"digital.vasic.challenges/pkg/challenge"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// mockChallengeService implements challengeService for testing.
type mockChallengeService struct {
	listChallengesFunc func() []services.ChallengeSummary
	runChallengeFunc   func(ctx context.Context, id string) (*challenge.Result, error)
	runAllFunc         func(ctx context.Context) ([]*challenge.Result, error)
	runByCategoryFunc  func(ctx context.Context, category string) ([]*challenge.Result, error)
	getResultsFunc     func() []*challenge.Result
}

func (m *mockChallengeService) ListChallenges() []services.ChallengeSummary {
	if m.listChallengesFunc != nil {
		return m.listChallengesFunc()
	}
	return nil
}

func (m *mockChallengeService) RunChallenge(ctx context.Context, id string) (*challenge.Result, error) {
	if m.runChallengeFunc != nil {
		return m.runChallengeFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockChallengeService) RunAll(ctx context.Context) ([]*challenge.Result, error) {
	if m.runAllFunc != nil {
		return m.runAllFunc(ctx)
	}
	return nil, nil
}

func (m *mockChallengeService) RunByCategory(ctx context.Context, category string) ([]*challenge.Result, error) {
	if m.runByCategoryFunc != nil {
		return m.runByCategoryFunc(ctx, category)
	}
	return nil, nil
}

func (m *mockChallengeService) GetResults() []*challenge.Result {
	if m.getResultsFunc != nil {
		return m.getResultsFunc()
	}
	return nil
}

type ChallengeHandlerTestSuite struct {
	suite.Suite
	handler *ChallengeHandler
	mock    *mockChallengeService
	router  *gin.Engine
}

func (suite *ChallengeHandlerTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
}

func (suite *ChallengeHandlerTestSuite) SetupTest() {
	suite.mock = &mockChallengeService{}
	suite.handler = NewChallengeHandler(suite.mock)
	suite.router = gin.New()
	suite.router.GET("/api/v1/challenges", suite.handler.ListChallenges)
	suite.router.GET("/api/v1/challenges/:id", suite.handler.GetChallenge)
	suite.router.POST("/api/v1/challenges/:id/run", suite.handler.RunChallenge)
	suite.router.POST("/api/v1/challenges/run/all", suite.handler.RunAll)
	suite.router.POST("/api/v1/challenges/run/category/:category", suite.handler.RunByCategory)
	suite.router.GET("/api/v1/challenges/results", suite.handler.GetResults)
}

func (suite *ChallengeHandlerTestSuite) TestListChallenges() {
	expected := []services.ChallengeSummary{
		{ID: "ch-001", Name: "Test Challenge", Description: "Test", Category: "test", Dependencies: []string{}},
	}
	suite.mock.listChallengesFunc = func() []services.ChallengeSummary {
		return expected
	}

	req := httptest.NewRequest("GET", "/api/v1/challenges", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	// TODO: parse response body and verify
}

func (suite *ChallengeHandlerTestSuite) TestGetChallenge_EmptyID() {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "id", Value: ""}}
	c.Request = httptest.NewRequest("GET", "/api/v1/challenges/", nil)

	suite.handler.GetChallenge(c)
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func (suite *ChallengeHandlerTestSuite) TestGetChallenge_MissingID() {
	// ID param missing from path but route expects :id, so this test is not possible
	// Instead test with empty string ID? The handler validates id == "" after extraction.
	// We'll need to call handler directly with gin context.
	// Let's test using direct handler call.
}

func (suite *ChallengeHandlerTestSuite) TestRunChallenge_EmptyID() {
	// Test via direct handler call
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = []gin.Param{}
	c.Request = httptest.NewRequest("POST", "/api/v1/challenges//run", nil)

	suite.handler.RunChallenge(c)
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func (suite *ChallengeHandlerTestSuite) TestRunByCategory_EmptyCategory() {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = []gin.Param{}
	c.Request = httptest.NewRequest("POST", "/api/v1/challenges/run/category/", nil)

	suite.handler.RunByCategory(c)
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func TestChallengeHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(ChallengeHandlerTestSuite))
}
