package services

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewSyncOperationManager(t *testing.T) {
	tests := []struct {
		name    string
		timeout time.Duration
		want    time.Duration
	}{
		{
			name:    "with custom timeout",
			timeout: 30 * time.Minute,
			want:    30 * time.Minute,
		},
		{
			name:    "with zero timeout uses default",
			timeout: 0,
			want:    30 * time.Minute,
		},
		{
			name:    "with negative timeout uses default",
			timeout: -1 * time.Minute,
			want:    30 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewSyncOperationManager(tt.timeout)
			assert.NotNil(t, manager)
			assert.NotNil(t, manager.operations)
			assert.Equal(t, tt.want, manager.timeout)
		})
	}
}

func TestSyncOperationManager_Register(t *testing.T) {
	manager := NewSyncOperationManager(5 * time.Minute)

	sessionID := 1
	endpointID := 100
	userID := 50

	op, ctx := manager.Register(sessionID, endpointID, userID)

	assert.NotNil(t, op)
	assert.NotNil(t, ctx)
	assert.Equal(t, sessionID, op.ID)
	assert.Equal(t, sessionID, op.SessionID)
	assert.Equal(t, endpointID, op.EndpointID)
	assert.Equal(t, userID, op.UserID)
	assert.Equal(t, SyncStateIdle, op.State)
	assert.NotNil(t, op.CancelFunc)
	assert.False(t, op.StartTime.IsZero())

	// Verify operation is stored
	retrievedOp, exists := manager.Get(sessionID)
	assert.True(t, exists)
	assert.Equal(t, op.ID, retrievedOp.ID)
}

func TestSyncOperationManager_Get(t *testing.T) {
	manager := NewSyncOperationManager(5 * time.Minute)

	// Get non-existent operation
	op, exists := manager.Get(999)
	assert.False(t, exists)
	assert.Nil(t, op)

	// Register and get
	manager.Register(1, 100, 50)
	op, exists = manager.Get(1)
	assert.True(t, exists)
	assert.NotNil(t, op)
}

func TestSyncOperationManager_Cancel(t *testing.T) {
	manager := NewSyncOperationManager(5 * time.Minute)

	// Cancel non-existent operation
	err := manager.Cancel(999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Register and cancel
	manager.Register(1, 100, 50)
	err = manager.Cancel(1)
	assert.NoError(t, err)

	// Verify state is cancelled
	op, exists := manager.Get(1)
	assert.True(t, exists)
	assert.Equal(t, SyncStateCancelled, op.State)
	assert.NotNil(t, op.EndTime)
}

func TestSyncOperationManager_Complete(t *testing.T) {
	manager := NewSyncOperationManager(5 * time.Minute)

	// Complete non-existent operation (should not panic)
	manager.Complete(999)

	// Register and complete
	manager.Register(1, 100, 50)
	manager.Complete(1)

	// Verify state is completed
	op, exists := manager.Get(1)
	assert.True(t, exists)
	assert.Equal(t, SyncStateCompleted, op.State)
	assert.NotNil(t, op.EndTime)
}

func TestSyncOperationManager_Fail(t *testing.T) {
	manager := NewSyncOperationManager(5 * time.Minute)
	testErr := assert.AnError

	// Fail non-existent operation (should not panic)
	manager.Fail(999, testErr)

	// Register and fail
	manager.Register(1, 100, 50)
	manager.Fail(1, testErr)

	// Verify state is failed
	op, exists := manager.Get(1)
	assert.True(t, exists)
	assert.Equal(t, SyncStateFailed, op.State)
	assert.Equal(t, testErr, op.Error)
	assert.NotNil(t, op.EndTime)
}

func TestSyncOperationManager_UpdateProgress(t *testing.T) {
	manager := NewSyncOperationManager(5 * time.Minute)
	manager.Register(1, 100, 50)

	// Update progress
	manager.UpdateProgress(1, 50, 5, 10, 100)

	op, exists := manager.Get(1)
	assert.True(t, exists)
	assert.Equal(t, 50, op.Processed)
	assert.Equal(t, 5, op.Failed)
	assert.Equal(t, 10, op.Skipped)
	assert.Equal(t, 100, op.TotalFiles)
	assert.Equal(t, 65, op.Progress) // (50+5+10)*100/100 = 65%
}

func TestSyncOperationManager_UpdateProgress_NotFound(t *testing.T) {
	manager := NewSyncOperationManager(5 * time.Minute)

	// Should not panic
	manager.UpdateProgress(999, 50, 5, 10, 100)
}

func TestSyncOperationManager_GetAllActive(t *testing.T) {
	manager := NewSyncOperationManager(5 * time.Minute)

	// Initially no active operations
	active := manager.GetAllActive()
	assert.Empty(t, active)

	// Register operations
	manager.Register(1, 100, 50)
	manager.Register(2, 101, 51)
	manager.Register(3, 102, 52)

	// Complete one
	manager.Complete(2)

	// Get active
	active = manager.GetAllActive()
	assert.Len(t, active, 2)
}

func TestSyncOperationManager_Cleanup(t *testing.T) {
	manager := NewSyncOperationManager(5 * time.Minute)

	// Register and complete operations
	manager.Register(1, 100, 50)
	manager.Register(2, 101, 51)
	manager.Register(3, 102, 52)

	manager.Complete(1)
	manager.Complete(2)

	// Wait a bit
	time.Sleep(10 * time.Millisecond)

	// Cleanup operations older than 5ms
	manager.Cleanup(5 * time.Millisecond)

	// Completed operations should be removed
	_, exists1 := manager.Get(1)
	_, exists2 := manager.Get(2)
	_, exists3 := manager.Get(3)

	assert.False(t, exists1, "completed operation 1 should be cleaned up")
	assert.False(t, exists2, "completed operation 2 should be cleaned up")
	assert.True(t, exists3, "active operation 3 should still exist")
}

func TestSyncStateMachine_Transition(t *testing.T) {
	tests := []struct {
		name      string
		from      SyncState
		to        SyncState
		wantErr   bool
		wantState SyncState
	}{
		{
			name:      "valid: idle to connecting",
			from:      SyncStateIdle,
			to:        SyncStateConnecting,
			wantErr:   false,
			wantState: SyncStateConnecting,
		},
		{
			name:      "valid: connecting to scanning",
			from:      SyncStateConnecting,
			to:        SyncStateScanning,
			wantErr:   false,
			wantState: SyncStateScanning,
		},
		{
			name:    "invalid: idle to transferring",
			from:    SyncStateIdle,
			to:      SyncStateTransferring,
			wantErr: true,
		},
		{
			name:      "valid: transferring to failed",
			from:      SyncStateTransferring,
			to:        SyncStateFailed,
			wantErr:   false,
			wantState: SyncStateFailed,
		},
		{
			name:      "valid: to completed sets end time",
			from:      SyncStateVerifying,
			to:        SyncStateCompleted,
			wantErr:   false,
			wantState: SyncStateCompleted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			op := &SyncOperation{
				ID:    1,
				State: tt.from,
			}
			sm := NewSyncStateMachine(op)

			err := sm.Transition(tt.to)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantState, op.State)
				if tt.to == SyncStateCompleted || tt.to == SyncStateFailed || tt.to == SyncStateCancelled {
					assert.NotNil(t, op.EndTime)
				}
			}
		})
	}
}

func TestSyncStateMachine_CanTransition(t *testing.T) {
	sm := NewSyncStateMachine(&SyncOperation{State: SyncStateIdle})

	assert.True(t, sm.CanTransition(SyncStateConnecting))
	assert.True(t, sm.CanTransition(SyncStateCancelled))
	assert.False(t, sm.CanTransition(SyncStateScanning))
	assert.False(t, sm.CanTransition(SyncStateCompleted))
}

func TestSyncMetrics_RecordOperation(t *testing.T) {
	metrics := &SyncMetrics{}

	// Record successful operation
	metrics.RecordOperation(true, 1024, 5*time.Second)
	assert.Equal(t, int64(1), metrics.TotalOperations)
	assert.Equal(t, int64(1), metrics.CompletedCount)
	assert.Equal(t, int64(0), metrics.FailedCount)
	assert.Equal(t, int64(1024), metrics.TotalBytes)

	// Record failed operation
	metrics.RecordOperation(false, 2048, 3*time.Second)
	assert.Equal(t, int64(2), metrics.TotalOperations)
	assert.Equal(t, int64(1), metrics.CompletedCount)
	assert.Equal(t, int64(1), metrics.FailedCount)
	assert.Equal(t, int64(3072), metrics.TotalBytes)

	// Check average duration
	expectedAvg := (5*time.Second + 3*time.Second) / 2
	assert.Equal(t, expectedAvg, metrics.AverageDuration)
}

func TestSyncMetrics_RecordCancellation(t *testing.T) {
	metrics := &SyncMetrics{}

	metrics.RecordCancellation()
	assert.Equal(t, int64(1), metrics.TotalOperations)
	assert.Equal(t, int64(1), metrics.CancelledCount)
}

func TestSyncMetrics_GetStats(t *testing.T) {
	metrics := &SyncMetrics{}
	metrics.RecordOperation(true, 1024, 5*time.Second)
	metrics.RecordOperation(false, 2048, 3*time.Second)

	stats := metrics.GetStats()

	assert.Equal(t, int64(2), stats["total_operations"])
	assert.Equal(t, int64(1), stats["completed"])
	assert.Equal(t, int64(1), stats["failed"])
	assert.Equal(t, int64(3072), stats["total_bytes"])
	assert.NotNil(t, stats["average_duration"])
	assert.NotNil(t, stats["last_sync"])
}

func TestSafeSyncContext(t *testing.T) {
	t.Run("successful execution", func(t *testing.T) {
		ctx := context.Background()
		err := SafeSyncContext(ctx, 5*time.Second, func(ctx context.Context) error {
			return nil
		})
		assert.NoError(t, err)
	})

	t.Run("execution error", func(t *testing.T) {
		ctx := context.Background()
		testErr := assert.AnError
		err := SafeSyncContext(ctx, 5*time.Second, func(ctx context.Context) error {
			return testErr
		})
		assert.Error(t, err)
		assert.Equal(t, testErr, err)
	})

	t.Run("timeout", func(t *testing.T) {
		ctx := context.Background()
		err := SafeSyncContext(ctx, 10*time.Millisecond, func(ctx context.Context) error {
			time.Sleep(100 * time.Millisecond)
			return nil
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "timed out")
	})

	t.Run("no timeout", func(t *testing.T) {
		ctx := context.Background()
		err := SafeSyncContext(ctx, 0, func(ctx context.Context) error {
			return nil
		})
		assert.NoError(t, err)
	})
}

func TestDefaultSyncTimeoutConfig(t *testing.T) {
	config := DefaultSyncTimeoutConfig()

	assert.Equal(t, 30*time.Second, config.ConnectionTimeout)
	assert.Equal(t, 5*time.Minute, config.ScanTimeout)
	assert.Equal(t, 30*time.Minute, config.TransferTimeout)
	assert.Equal(t, 5*time.Minute, config.VerifyTimeout)
}
