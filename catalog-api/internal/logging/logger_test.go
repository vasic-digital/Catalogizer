package logging

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestInit(t *testing.T) {
	tests := []struct {
		name    string
		env     string
		wantErr bool
	}{
		{
			name:    "development environment",
			env:     "development",
			wantErr: false,
		},
		{
			name:    "production environment",
			env:     "production",
			wantErr: false,
		},
		{
			name:    "dev shorthand",
			env:     "dev",
			wantErr: false,
		},
		{
			name:    "empty environment defaults to production",
			env:     "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Init(tt.env)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, Logger)
				assert.NotNil(t, SugaredLogger)
				_ = Sync() // Clean up
			}
		})
	}
}

func TestInitWithConfig(t *testing.T) {
	config := zap.NewDevelopmentConfig()
	config.DisableCaller = true

	err := InitWithConfig(config)
	require.NoError(t, err)
	assert.NotNil(t, Logger)
	assert.NotNil(t, SugaredLogger)

	_ = Sync() // Clean up
}

func TestLoggingFunctions(t *testing.T) {
	// Initialize logger
	err := Init("development")
	require.NoError(t, err)
	defer func() { _ = Sync() }()

	// Test that logging functions don't panic
	t.Run("Debug", func(t *testing.T) {
		assert.NotPanics(t, func() {
			Debug("debug message", String("key", "value"))
		})
	})

	t.Run("Info", func(t *testing.T) {
		assert.NotPanics(t, func() {
			Info("info message", Int("count", 42))
		})
	})

	t.Run("Warn", func(t *testing.T) {
		assert.NotPanics(t, func() {
			Warn("warning message", Bool("flag", true))
		})
	})

	t.Run("Error", func(t *testing.T) {
		assert.NotPanics(t, func() {
			Error("error message", ErrorField(errors.New("test error")))
		})
	})

	t.Run("Formatted logging", func(t *testing.T) {
		assert.NotPanics(t, func() {
			Debugf("debug %s %d", "message", 1)
			Infof("info %s %d", "message", 2)
			Warnf("warning %s %d", "message", 3)
			Errorf("error %s %d", "message", 4)
		})
	})
}

func TestFieldHelpers(t *testing.T) {
	t.Run("String field", func(t *testing.T) {
		field := String("key", "value")
		assert.Equal(t, "key", field.Key)
		assert.Equal(t, "value", field.String)
	})

	t.Run("Int field", func(t *testing.T) {
		field := Int("count", 42)
		assert.Equal(t, "count", field.Key)
		assert.Equal(t, int64(42), field.Integer)
	})

	t.Run("Int64 field", func(t *testing.T) {
		field := Int64("size", 9223372036854775807)
		assert.Equal(t, "size", field.Key)
		assert.Equal(t, int64(9223372036854775807), field.Integer)
	})

	t.Run("Bool field", func(t *testing.T) {
		field := Bool("enabled", true)
		assert.Equal(t, "enabled", field.Key)
		assert.Equal(t, int64(1), field.Integer) // zap stores bool as int64
	})

	t.Run("Error field", func(t *testing.T) {
		testErr := errors.New("test error")
		field := ErrorField(testErr)
		assert.Equal(t, "error", field.Key)
		assert.Equal(t, testErr, field.Interface)
	})

	t.Run("Any field", func(t *testing.T) {
		field := Any("data", map[string]string{"key": "value"})
		assert.Equal(t, "data", field.Key)
		assert.Equal(t, map[string]string{"key": "value"}, field.Interface)
	})
}

func TestWith(t *testing.T) {
	err := Init("development")
	require.NoError(t, err)
	defer func() { _ = Sync() }()

	childLogger := With(String("component", "test"))
	assert.NotNil(t, childLogger)
}

func TestWithFields(t *testing.T) {
	err := Init("development")
	require.NoError(t, err)
	defer func() { _ = Sync() }()

	childLogger := WithFields(map[string]interface{}{
		"service": "test-service",
		"version": "1.0.0",
	})
	assert.NotNil(t, childLogger)
}

func TestSyncWithNilLogger(t *testing.T) {
	// Reset logger
	Logger = nil
	SugaredLogger = nil

	// Should not panic
	assert.NotPanics(t, func() {
		err := Sync()
		assert.NoError(t, err)
	})
}

func TestLoggingWithNilLogger(t *testing.T) {
	// Reset logger
	Logger = nil
	SugaredLogger = nil

	// All logging functions should not panic when logger is nil
	t.Run("structured logging", func(t *testing.T) {
		assert.NotPanics(t, func() {
			Debug("debug", String("key", "value"))
			Info("info", String("key", "value"))
			Warn("warn", String("key", "value"))
			Error("error", String("key", "value"))
		})
	})

	t.Run("formatted logging", func(t *testing.T) {
		assert.NotPanics(t, func() {
			Debugf("debug %s", "test")
			Infof("info %s", "test")
			Warnf("warn %s", "test")
			Errorf("error %s", "test")
		})
	})
}

func TestWithNilLogger(t *testing.T) {
	// Reset logger
	Logger = nil
	SugaredLogger = nil

	t.Run("With returns nil", func(t *testing.T) {
		child := With(String("key", "value"))
		assert.Nil(t, child)
	})

	t.Run("WithFields returns nil", func(t *testing.T) {
		child := WithFields(map[string]interface{}{"key": "value"})
		assert.Nil(t, child)
	})
}
