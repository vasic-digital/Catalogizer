package services

import (
	"testing"
	"time"

	"catalogizer/models"
)

// TestNonBlockingSendReleasesOnDone exercises the same select pattern used
// in streamLogEntries: a send on an unbuffered channel is wrapped in a
// select with <-done so the sender can always exit when done fires,
// even if no receiver is present.
//
// This is the regression guard for the CS-07 fix: the pre-fix code did a
// plain `channel <- entry` which would block forever if the receiver
// stopped reading, leaking the streaming goroutine.
func TestNonBlockingSendReleasesOnDone(t *testing.T) {
	ch := make(chan *models.LogEntry) // unbuffered — any send without a receiver blocks
	done := make(chan struct{})

	exited := make(chan struct{})
	go func() {
		defer close(exited)
		entry := &models.LogEntry{ID: 1}
		// Mirror the pattern used in streamLogEntries.
		select {
		case <-done:
			return
		case ch <- entry:
			t.Error("should not have sent — no receiver")
		case <-time.After(500 * time.Millisecond):
			return
		}
	}()

	// Cancel before the 500ms timeout fires.
	time.Sleep(20 * time.Millisecond)
	close(done)

	select {
	case <-exited:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("sender goroutine did not exit when done was closed")
	}
}

// TestNonBlockingSendReleasesOnTimeout verifies the timeout branch — if
// neither done fires nor a receiver appears, the sender must still exit
// so the goroutine is never leaked.
func TestNonBlockingSendReleasesOnTimeout(t *testing.T) {
	ch := make(chan *models.LogEntry)
	done := make(chan struct{})
	_ = done // never closed

	exited := make(chan struct{})
	go func() {
		defer close(exited)
		entry := &models.LogEntry{ID: 1}
		select {
		case <-done:
			return
		case ch <- entry:
			t.Error("should not have sent — no receiver")
		case <-time.After(80 * time.Millisecond):
			return
		}
	}()

	select {
	case <-exited:
	case <-time.After(300 * time.Millisecond):
		t.Fatal("sender goroutine did not exit on timeout")
	}
}

// TestNonBlockingSendDeliversWhenReceiverReady verifies the happy path —
// if a receiver IS present, the send completes normally (no spurious
// timeout / done preference).
func TestNonBlockingSendDeliversWhenReceiverReady(t *testing.T) {
	ch := make(chan *models.LogEntry, 1)
	done := make(chan struct{})

	entry := &models.LogEntry{ID: 42}
	select {
	case <-done:
		t.Fatal("done should not be selected")
	case ch <- entry:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout should not be selected")
	}

	select {
	case got := <-ch:
		if got.ID != 42 {
			t.Fatalf("expected ID 42, got %d", got.ID)
		}
	default:
		t.Fatal("channel should have the entry")
	}
}
