package testutils

import (
	"testing"

	"github.com/stretchr/testify/mock"
)

// MockRepository provides a mock repository for service testing
type MockRepository struct {
	mock.Mock
}

// ServiceTestTemplate provides common test setup for services
type ServiceTestTemplate struct {
	t        *testing.T
	mockRepo *MockRepository
	cleanup  func()
}

// NewServiceTestTemplate creates a new service test template
func NewServiceTestTemplate(t *testing.T) *ServiceTestTemplate {
	mockRepo := &MockRepository{}

	return &ServiceTestTemplate{
		t:        t,
		mockRepo: mockRepo,
		cleanup: func() {
			// Cleanup if needed
		},
	}
}

// Cleanup cleans up test resources
func (stt *ServiceTestTemplate) Cleanup() {
	if stt.cleanup != nil {
		stt.cleanup()
	}
}

// ServiceTestSuite defines interface for service test suites
type ServiceTestSuite interface {
	SetupTest()
	TeardownTest()
	TestCreate()
	TestGetByID()
	TestGetAll()
	TestUpdate()
	TestDelete()
	TestBusinessLogic()
}

// RunServiceTestSuite runs a complete service test suite
func RunServiceTestSuite(t *testing.T, suite ServiceTestSuite) {
	t.Run("Setup", func(t *testing.T) {
		suite.SetupTest()
	})

	t.Run("Create", func(t *testing.T) {
		suite.TestCreate()
	})

	t.Run("GetByID", func(t *testing.T) {
		suite.TestGetByID()
	})

	t.Run("GetAll", func(t *testing.T) {
		suite.TestGetAll()
	})

	t.Run("Update", func(t *testing.T) {
		suite.TestUpdate()
	})

	t.Run("Delete", func(t *testing.T) {
		suite.TestDelete()
	})

	t.Run("BusinessLogic", func(t *testing.T) {
		suite.TestBusinessLogic()
	})

	t.Run("Teardown", func(t *testing.T) {
		suite.TeardownTest()
	})
}

// TestCase represents a single test case
type TestCase struct {
	Name    string
	Setup   func()
	Action  func() error
	Assert  func(t *testing.T, err error)
	Cleanup func()
}

// RunTestCases runs multiple test cases
func RunTestCases(t *testing.T, testCases []TestCase) {
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			if tc.Setup != nil {
				tc.Setup()
			}

			if tc.Cleanup != nil {
				defer tc.Cleanup()
			}

			err := tc.Action()
			tc.Assert(t, err)
		})
	}
}
