package eventbus

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Event metadata
// ---------------------------------------------------------------------------

func TestNewEvent_MetadataCanBeEnriched(t *testing.T) {
	evt := NewEvent(EventFileCreated, "watcher", "test-payload")

	evt.WithMetadata("path", "/media/video.mp4")
	evt.WithMetadata("size", "1024")

	assert.Equal(t, "/media/video.mp4", evt.Metadata["path"])
	assert.Equal(t, "1024", evt.Metadata["size"])
}

func TestNewEvent_MetadataOverwrite(t *testing.T) {
	evt := NewEvent(EventFileCreated, "watcher", nil)

	evt.WithMetadata("key", "value1")
	assert.Equal(t, "value1", evt.Metadata["key"])

	evt.WithMetadata("key", "value2")
	assert.Equal(t, "value2", evt.Metadata["key"])
}

// ---------------------------------------------------------------------------
// Multiple subscribers same type
// ---------------------------------------------------------------------------

func TestEventBus_MultipleSubscribersSameType(t *testing.T) {
	bus := New(DefaultConfig())
	defer bus.Close()

	sub1 := bus.Subscribe(EventFileCreated)
	defer sub1.Cancel()
	sub2 := bus.Subscribe(EventFileCreated)
	defer sub2.Cancel()
	sub3 := bus.Subscribe(EventFileCreated)
	defer sub3.Cancel()

	bus.Publish(NewEvent(EventFileCreated, "watcher", "test"))

	// All 3 subscribers should receive the event
	for i, sub := range []*Subscription{sub1, sub2, sub3} {
		select {
		case received := <-sub.Channel:
			assert.Equal(t, EventFileCreated, received.Type,
				"subscriber %d should receive correct type", i)
		case <-time.After(time.Second):
			t.Fatalf("subscriber %d timed out", i)
		}
	}
}

// ---------------------------------------------------------------------------
// Cancel subscription during publish
// ---------------------------------------------------------------------------

func TestEventBus_CancelDuringPublish(t *testing.T) {
	bus := New(DefaultConfig())
	defer bus.Close()

	sub := bus.Subscribe(EventScanStarted)

	// Cancel then publish — should not panic
	sub.Cancel()
	bus.Publish(NewEvent(EventScanStarted, "test", nil))

	// Channel should be closed after cancel
	select {
	case _, open := <-sub.Channel:
		if open {
			t.Log("received buffered event after cancel (acceptable)")
		}
	default:
		// Channel drained or closed — expected
	}
}

// ---------------------------------------------------------------------------
// SubscribeMultiple — event type isolation
// ---------------------------------------------------------------------------

func TestEventBus_SubscribeMultiple_DoesNotReceiveOther(t *testing.T) {
	bus := New(DefaultConfig())
	defer bus.Close()

	sub := bus.SubscribeMultiple(EventScanStarted, EventScanCompleted)
	defer sub.Cancel()

	// Publish unrelated type
	bus.Publish(NewEvent(EventFileDeleted, "watcher", nil))

	select {
	case <-sub.Channel:
		t.Fatal("should not receive event of unsubscribed type")
	case <-time.After(50 * time.Millisecond):
		// Expected
	}
}

// ---------------------------------------------------------------------------
// Metrics — events dropped (slow subscriber)
// ---------------------------------------------------------------------------

func TestEventBus_Metrics_InitialState(t *testing.T) {
	bus := New(DefaultConfig())
	defer bus.Close()

	metrics := bus.Metrics()
	require.NotNil(t, metrics)
	assert.Equal(t, int64(0), metrics.EventsPublished)
	assert.Equal(t, int64(0), metrics.EventsDelivered)
	assert.Equal(t, int64(0), metrics.SubscribersActive)
}

func TestEventBus_Metrics_AfterMultipleEvents(t *testing.T) {
	bus := New(DefaultConfig())
	defer bus.Close()

	sub := bus.Subscribe(EventEntityCreated)
	defer sub.Cancel()

	count := 5
	for i := 0; i < count; i++ {
		bus.Publish(NewEvent(EventEntityCreated, "agg", i))
	}

	// Drain events
	for i := 0; i < count; i++ {
		select {
		case <-sub.Channel:
		case <-time.After(time.Second):
			t.Fatal("timed out draining events")
		}
	}

	metrics := bus.Metrics()
	assert.Equal(t, int64(count), metrics.EventsPublished)
	assert.Equal(t, int64(count), metrics.EventsDelivered)
}

// ---------------------------------------------------------------------------
// Catalogizer-specific event types — namespace groups
// ---------------------------------------------------------------------------

func TestEventTypes_ScanNamespace(t *testing.T) {
	scanTypes := []EventType{EventScanStarted, EventScanCompleted, EventScanFailed}
	for _, et := range scanTypes {
		assert.Contains(t, string(et), "scan.",
			"scan event types must be in scan. namespace")
	}
}

func TestEventTypes_FileNamespace(t *testing.T) {
	fileTypes := []EventType{EventFileCreated, EventFileModified, EventFileDeleted, EventFileMoved}
	for _, et := range fileTypes {
		assert.Contains(t, string(et), "file.",
			"file event types must be in file. namespace")
	}
}

func TestEventTypes_EntityNamespace(t *testing.T) {
	entityTypes := []EventType{EventEntityCreated, EventEntityUpdated}
	for _, et := range entityTypes {
		assert.Contains(t, string(et), "entity.",
			"entity event types must be in entity. namespace")
	}
}

func TestEventTypes_SystemNamespace(t *testing.T) {
	systemTypes := []EventType{EventSystemStartup, EventSystemShutdown}
	for _, et := range systemTypes {
		assert.Contains(t, string(et), "system.",
			"system event types must be in system. namespace")
	}
}

// ---------------------------------------------------------------------------
// SubscribeWithFilter — complex filters
// ---------------------------------------------------------------------------

func TestEventBus_SubscribeWithFilter_MultipleConditions(t *testing.T) {
	bus := New(DefaultConfig())
	defer bus.Close()

	// Filter: only events from "scanner" source with non-nil payload
	sub := bus.SubscribeWithFilter(EventScanCompleted, func(e *Event) bool {
		return e.Source == "scanner" && e.Payload != nil
	})
	defer sub.Cancel()

	// Should be filtered out — nil payload
	bus.Publish(NewEvent(EventScanCompleted, "scanner", nil))

	// Should be filtered out — wrong source
	bus.Publish(NewEvent(EventScanCompleted, "other", "data"))

	// Should pass — correct source and non-nil payload
	bus.Publish(NewEvent(EventScanCompleted, "scanner", "completed"))

	select {
	case received := <-sub.Channel:
		assert.Equal(t, "scanner", received.Source)
		assert.Equal(t, "completed", received.Payload)
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for filtered event")
	}
}

// ---------------------------------------------------------------------------
// Close — publish after close
// ---------------------------------------------------------------------------

func TestEventBus_PublishAfterClose(t *testing.T) {
	bus := New(DefaultConfig())
	bus.Close()

	// Publish after close should not panic
	assert.NotPanics(t, func() {
		bus.Publish(NewEvent(EventScanStarted, "test", nil))
	})
}
