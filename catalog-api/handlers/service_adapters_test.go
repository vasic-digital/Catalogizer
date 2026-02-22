package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ServiceAdaptersTestSuite struct {
	suite.Suite
}

// AuthServiceAdapter tests

func (suite *ServiceAdaptersTestSuite) TestAuthServiceAdapter_NilInner() {
	adapter := &AuthServiceAdapter{Inner: nil}
	assert.NotNil(suite.T(), adapter)
	assert.Nil(suite.T(), adapter.Inner)
}

func (suite *ServiceAdaptersTestSuite) TestAuthServiceAdapter_CheckPermission_NilInner() {
	adapter := &AuthServiceAdapter{Inner: nil}
	// CheckPermission accesses userRepo which is nil, causing a panic
	assert.Panics(suite.T(), func() {
		adapter.CheckPermission(1, "admin")
	})
}

func (suite *ServiceAdaptersTestSuite) TestAuthServiceAdapter_GetCurrentUser_NilInner() {
	// GetCurrentUser calls validateToken which uses jwtSecret (nil byte slice)
	// On a nil *AuthService, this may not panic if the method handles nil gracefully
	adapter := &AuthServiceAdapter{Inner: nil}
	user, err := adapter.GetCurrentUser("invalid-token")
	// With a nil AuthService, jwtSecret is nil, so token parsing fails with error
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), user)
}

func (suite *ServiceAdaptersTestSuite) TestAuthServiceAdapter_HashPassword_NilInner() {
	// HashPasswordForUser calls generateSalt which creates crypto random bytes
	// then hashPassword which uses bcrypt - should work on nil receiver
	adapter := &AuthServiceAdapter{Inner: nil}
	hash, err := adapter.HashPassword("password123")
	// This should succeed since the method doesn't access struct fields for hashing
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), hash)
}

func (suite *ServiceAdaptersTestSuite) TestAuthServiceAdapter_ValidatePassword_NilInner() {
	// ValidatePassword only checks string length, no struct fields accessed
	adapter := &AuthServiceAdapter{Inner: nil}
	err := adapter.ValidatePassword("Password123!")
	assert.NoError(suite.T(), err)
}

func (suite *ServiceAdaptersTestSuite) TestAuthServiceAdapter_ValidatePassword_TooShort() {
	adapter := &AuthServiceAdapter{Inner: nil}
	err := adapter.ValidatePassword("short")
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "at least 8 characters")
}

func (suite *ServiceAdaptersTestSuite) TestAuthServiceAdapter_HashData_NilInner() {
	// HashData uses sha256 directly, no struct fields accessed
	adapter := &AuthServiceAdapter{Inner: nil}
	result := adapter.HashData("test-data")
	assert.NotEmpty(suite.T(), result)
}

func (suite *ServiceAdaptersTestSuite) TestAuthServiceAdapter_HashData_Deterministic() {
	adapter := &AuthServiceAdapter{Inner: nil}
	result1 := adapter.HashData("same-data")
	result2 := adapter.HashData("same-data")
	assert.Equal(suite.T(), result1, result2)
}

func (suite *ServiceAdaptersTestSuite) TestAuthServiceAdapter_HashData_DifferentInputs() {
	adapter := &AuthServiceAdapter{Inner: nil}
	result1 := adapter.HashData("data-one")
	result2 := adapter.HashData("data-two")
	assert.NotEqual(suite.T(), result1, result2)
}

// ConfigurationServiceAdapter tests

func (suite *ServiceAdaptersTestSuite) TestConfigurationServiceAdapter_NilInner() {
	adapter := &ConfigurationServiceAdapter{Inner: nil}
	assert.NotNil(suite.T(), adapter)
	assert.Nil(suite.T(), adapter.Inner)
}

func (suite *ServiceAdaptersTestSuite) TestConfigurationServiceAdapter_TestConfiguration_NilConfig() {
	adapter := &ConfigurationServiceAdapter{Inner: nil}
	result, err := adapter.TestConfiguration(nil)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), result.IsValid)
}

func (suite *ServiceAdaptersTestSuite) TestConfigurationServiceAdapter_GetConfigurationSchema_NilInner() {
	adapter := &ConfigurationServiceAdapter{Inner: nil}
	schema, err := adapter.GetConfigurationSchema()
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), schema)
}

// ErrorReportingServiceAdapter tests

func (suite *ServiceAdaptersTestSuite) TestErrorReportingServiceAdapter_NilInner() {
	adapter := &ErrorReportingServiceAdapter{Inner: nil}
	assert.NotNil(suite.T(), adapter)
	assert.Nil(suite.T(), adapter.Inner)
}

func (suite *ServiceAdaptersTestSuite) TestErrorReportingServiceAdapter_ReportError_NilInner() {
	adapter := &ErrorReportingServiceAdapter{Inner: nil}
	assert.Panics(suite.T(), func() {
		adapter.ReportError(1, nil)
	})
}

// LogManagementServiceAdapter tests

func (suite *ServiceAdaptersTestSuite) TestLogManagementServiceAdapter_NilInner() {
	adapter := &LogManagementServiceAdapter{Inner: nil}
	assert.NotNil(suite.T(), adapter)
	assert.Nil(suite.T(), adapter.Inner)
}

func (suite *ServiceAdaptersTestSuite) TestLogManagementServiceAdapter_CollectLogs_NilInner() {
	adapter := &LogManagementServiceAdapter{Inner: nil}
	assert.Panics(suite.T(), func() {
		adapter.CollectLogs(1, nil)
	})
}

func TestServiceAdaptersTestSuite(t *testing.T) {
	suite.Run(t, new(ServiceAdaptersTestSuite))
}
