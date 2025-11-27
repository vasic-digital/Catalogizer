package testutils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHelper provides common testing utilities
type TestHelper struct {
	T *testing.T
}

// NewTestHelper creates a new TestHelper instance
func NewTestHelper(t *testing.T) *TestHelper {
	return &TestHelper{T: t}
}

// AssertNoError asserts that an error is nil
func (h *TestHelper) AssertNoError(err error, msgAndArgs ...interface{}) {
	assert.NoError(h.T, err, msgAndArgs...)
}

// AssertError asserts that an error is not nil
func (h *TestHelper) AssertError(err error, msgAndArgs ...interface{}) {
	assert.Error(h.T, err, msgAndArgs...)
}

// RequireNoError requires that an error is nil (fails test if not)
func (h *TestHelper) RequireNoError(err error, msgAndArgs ...interface{}) {
	require.NoError(h.T, err, msgAndArgs...)
}

// RequireError requires that an error is not nil
func (h *TestHelper) RequireError(err error, msgAndArgs ...interface{}) {
	require.Error(h.T, err, msgAndArgs...)
}

// AssertEqual asserts that two values are equal
func (h *TestHelper) AssertEqual(expected, actual interface{}, msgAndArgs ...interface{}) {
	assert.Equal(h.T, expected, actual, msgAndArgs...)
}

// AssertNotEqual asserts that two values are not equal
func (h *TestHelper) AssertNotEqual(expected, actual interface{}, msgAndArgs ...interface{}) {
	assert.NotEqual(h.T, expected, actual, msgAndArgs...)
}

// AssertTrue asserts that a condition is true
func (h *TestHelper) AssertTrue(condition bool, msgAndArgs ...interface{}) {
	assert.True(h.T, condition, msgAndArgs...)
}

// AssertFalse asserts that a condition is false
func (h *TestHelper) AssertFalse(condition bool, msgAndArgs ...interface{}) {
	assert.False(h.T, condition, msgAndArgs...)
}

// AssertNotNil asserts that a value is not nil
func (h *TestHelper) AssertNotNil(value interface{}, msgAndArgs ...interface{}) {
	assert.NotNil(h.T, value, msgAndArgs...)
}

// AssertNil asserts that a value is nil
func (h *TestHelper) AssertNil(value interface{}, msgAndArgs ...interface{}) {
	assert.Nil(h.T, value, msgAndArgs...)
}

// RequireEqual requires that two values are equal
func (h *TestHelper) RequireEqual(expected, actual interface{}, msgAndArgs ...interface{}) {
	require.Equal(h.T, expected, actual, msgAndArgs...)
}

// RequireTrue requires that a condition is true
func (h *TestHelper) RequireTrue(condition bool, msgAndArgs ...interface{}) {
	require.True(h.T, condition, msgAndArgs...)
}

// WaitForCondition waits for a condition to become true
func (h *TestHelper) WaitForCondition(condition func() bool, timeout time.Duration) bool {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	timeoutChan := time.After(timeout)

	for {
		select {
		case <-ticker.C:
			if condition() {
				return true
			}
		case <-timeoutChan:
			h.T.Errorf("Condition not met within timeout %v", timeout)
			return false
		}
	}
}