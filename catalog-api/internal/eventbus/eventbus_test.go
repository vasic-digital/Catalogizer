package eventbus

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// allEventTypes returns all 13 Catalogizer-specific event type constants.
func allEventTypes() []struct {
	Name  string
	Value EventType
} {
	return []struct {
		Name  string
		Value EventType
	}{
		{"EventScanStarted", EventScanStarted},
		{"EventScanCompleted", EventScanCompleted},
		{"EventScanFailed", EventScanFailed},
		{"EventFileCreated", EventFileCreated},
		{"EventFileModified", EventFileModified},
		{"EventFileDeleted", EventFileDeleted},
		{"EventFileMoved", EventFileMoved},
		{"EventEntityCreated", EventEntityCreated},
		{"EventEntityUpdated", EventEntityUpdated},
		{"EventMetaRefreshed", EventMetaRefreshed},
		{"EventCacheEvicted", EventCacheEvicted},
		{"EventSystemStartup", EventSystemStartup},
		{"EventSystemShutdown", EventSystemShutdown},
	}
}

// ---------------------------------------------------------------------------
// DefaultConfig tests
// ---------------------------------------------------------------------------

func TestDefaultConfig_ReturnsNonNil(t *testing.T) {
	cfg := DefaultConfig()
	require.NotNil(t, cfg, "DefaultConfig() must return a non-nil *Config")
}

func TestDefaultConfig_HasSensibleDefaults(t *testing.T) {
	cfg := DefaultConfig()

	assert.Greater(t, cfg.BufferSize, 0, "BufferSize must be positive")
	assert.Greater(t, cfg.PublishTimeout, time.Duration(0), "PublishTimeout must be positive")
	assert.Greater(t, cfg.CleanupInterval, time.Duration(0), "CleanupInterval must be positive")
	assert.Greater(t, cfg.MaxSubscribers, 0, "MaxSubscribers must be positive")
}

// ---------------------------------------------------------------------------
// New tests
// ---------------------------------------------------------------------------

func TestNew_WithDefaultConfig(t *testing.T) {
	bus := New(DefaultConfig())
	require.NotNil(t, bus, "New() with DefaultConfig must return a non-nil *EventBus")
	defer bus.Close()
}

func TestNew_WithNilConfig(t *testing.T) {
	bus := New(nil)
	require.NotNil(t, bus, "New(nil) must return a non-nil *EventBus (uses defaults)")
	defer bus.Close()
}

func TestNew_WithCustomConfig(t *testing.T) {
	cfg := &Config{
		BufferSize:      50,
		PublishTimeout:  5 * time.Millisecond,
		CleanupInterval: 10 * time.Second,
		MaxSubscribers:  25,
	}

	bus := New(cfg)
	require.NotNil(t, bus, "New() with custom Config must return a non-nil *EventBus")
	defer bus.Close()
}

// ---------------------------------------------------------------------------
// NewEvent tests
// ---------------------------------------------------------------------------

func TestNewEvent_FieldsPopulated(t *testing.T) {
	tests := []struct {
		name      string
		eventType EventType
		source    string
		payload   interface{}
	}{
		{
			name:      "string payload",
			eventType: EventScanStarted,
			source:    "scanner",
			payload:   "starting scan",
		},
		{
			name:      "map payload",
			eventType: EventFileCreated,
			source:    "watcher",
			payload:   map[string]string{"path": "/media/video.mp4"},
		},
		{
			name:      "nil payload",
			eventType: EventSystemStartup,
			source:    "system",
			payload:   nil,
		},
		{
			name:      "int payload",
			eventType: EventCacheEvicted,
			source:    "cache",
			payload:   42,
		},
		{
			name:      "struct payload",
			eventType: EventEntityCreated,
			source:    "aggregation",
			payload:   struct{ ID int }{ID: 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := time.Now()
			evt := NewEvent(tt.eventType, tt.source, tt.payload)
			after := time.Now()

			require.NotNil(t, evt, "NewEvent must return a non-nil *Event")
			assert.Equal(t, tt.eventType, evt.Type, "event Type must match")
			assert.Equal(t, tt.source, evt.Source, "event Source must match")
			assert.Equal(t, tt.payload, evt.Payload, "event Payload must match")
			assert.NotEmpty(t, evt.ID, "event ID must be auto-generated")
			assert.NotEmpty(t, evt.TraceID, "event TraceID must be auto-generated")
			assert.NotEqual(t, evt.ID, evt.TraceID, "ID and TraceID should differ")
			assert.False(t, evt.Timestamp.IsZero(), "event Timestamp must be set")
			assert.True(t, !evt.Timestamp.Before(before) && !evt.Timestamp.After(after),
				"Timestamp must be between before and after creation")
			assert.NotNil(t, evt.Metadata, "Metadata map must be initialized")
		})
	}
}

func TestNewEvent_UniqueIDs(t *testing.T) {
	ids := make(map[string]bool)
	count := 100

	for i := 0; i < count; i++ {
		evt := NewEvent(EventScanStarted, "test", nil)
		assert.False(t, ids[evt.ID], "event ID must be unique, duplicate found: %s", evt.ID)
		ids[evt.ID] = true
	}
}

// ---------------------------------------------------------------------------
// Event type constant tests
// ---------------------------------------------------------------------------

func TestEventTypeConstants_Count(t *testing.T) {
	types := allEventTypes()
	assert.Equal(t, 13, len(types), "must have exactly 13 Catalogizer event type constants")
}

func TestEventTypeConstants_Unique(t *testing.T) {
	types := allEventTypes()
	seen := make(map[EventType]string)

	for _, tt := range types {
		if prev, exists := seen[tt.Value]; exists {
			t.Errorf("duplicate event type value %q: %s and %s", tt.Value, prev, tt.Name)
		}
		seen[tt.Value] = tt.Name
	}
}

func TestEventTypeConstants_DotNotation(t *testing.T) {
	types := allEventTypes()

	for _, tt := range types {
		t.Run(tt.Name, func(t *testing.T) {
			val := string(tt.Value)
			assert.NotEmpty(t, val, "event type must not be empty")
			assert.Contains(t, val, ".", "event type must use dot-notation")

			parts := strings.Split(val, ".")
			assert.Equal(t, 2, len(parts),
				"event type %q must have exactly two dot-separated segments", val)

			for _, part := range parts {
				assert.NotEmpty(t, part,
					"dot-notation segments must not be empty in %q", val)
				assert.Equal(t, strings.ToLower(part), part,
					"dot-notation segments must be lowercase in %q", val)
			}
		})
	}
}

func TestEventTypeConstants_ExpectedValues(t *testing.T) {
	tests := []struct {
		name     string
		constant EventType
		expected string
	}{
		{"EventScanStarted", EventScanStarted, "scan.started"},
		{"EventScanCompleted", EventScanCompleted, "scan.completed"},
		{"EventScanFailed", EventScanFailed, "scan.failed"},
		{"EventFileCreated", EventFileCreated, "file.created"},
		{"EventFileModified", EventFileModified, "file.modified"},
		{"EventFileDeleted", EventFileDeleted, "file.deleted"},
		{"EventFileMoved", EventFileMoved, "file.moved"},
		{"EventEntityCreated", EventEntityCreated, "entity.created"},
		{"EventEntityUpdated", EventEntityUpdated, "entity.updated"},
		{"EventMetaRefreshed", EventMetaRefreshed, "metadata.refreshed"},
		{"EventCacheEvicted", EventCacheEvicted, "cache.evicted"},
		{"EventSystemStartup", EventSystemStartup, "system.startup"},
		{"EventSystemShutdown", EventSystemShutdown, "system.shutdown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, EventType(tt.expected), tt.constant,
				"%s must equal %q", tt.name, tt.expected)
		})
	}
}

// ---------------------------------------------------------------------------
// Publish / Subscribe tests
// ---------------------------------------------------------------------------

func TestEventBus_Subscribe_ReceivesPublishedEvent(t *testing.T) {
	bus := New(DefaultConfig())
	defer bus.Close()

	sub := bus.Subscribe(EventScanStarted)
	defer sub.Cancel()

	payload := map[string]string{"root": "/media"}
	evt := NewEvent(EventScanStarted, "scanner", payload)
	bus.Publish(evt)

	select {
	case received := <-sub.Channel:
		require.NotNil(t, received)
		assert.Equal(t, evt.ID, received.ID)
		assert.Equal(t, EventScanStarted, received.Type)
		assert.Equal(t, "scanner", received.Source)
		assert.Equal(t, payload, received.Payload)
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for event")
	}
}

func TestEventBus_Subscribe_DoesNotReceiveOtherTypes(t *testing.T) {
	bus := New(DefaultConfig())
	defer bus.Close()

	sub := bus.Subscribe(EventScanStarted)
	defer sub.Cancel()

	evt := NewEvent(EventFileCreated, "watcher", nil)
	bus.Publish(evt)

	select {
	case received := <-sub.Channel:
		t.Fatalf("should not receive event of different type, got: %+v", received)
	case <-time.After(50 * time.Millisecond):
		// Expected: no event received.
	}
}

func TestEventBus_SubscribeMultiple(t *testing.T) {
	bus := New(DefaultConfig())
	defer bus.Close()

	sub := bus.SubscribeMultiple(EventScanStarted, EventScanCompleted)
	defer sub.Cancel()

	evt1 := NewEvent(EventScanStarted, "scanner", nil)
	evt2 := NewEvent(EventScanCompleted, "scanner", nil)
	bus.Publish(evt1)
	bus.Publish(evt2)

	received := make([]*Event, 0, 2)
	for i := 0; i < 2; i++ {
		select {
		case e := <-sub.Channel:
			received = append(received, e)
		case <-time.After(time.Second):
			t.Fatalf("timed out waiting for event %d", i+1)
		}
	}

	assert.Equal(t, 2, len(received))
	assert.Equal(t, evt1.ID, received[0].ID)
	assert.Equal(t, evt2.ID, received[1].ID)
}

func TestEventBus_SubscribeAll(t *testing.T) {
	bus := New(DefaultConfig())
	defer bus.Close()

	sub := bus.SubscribeAll()
	defer sub.Cancel()

	types := []EventType{EventScanStarted, EventFileCreated, EventEntityCreated}
	for _, et := range types {
		bus.Publish(NewEvent(et, "test", nil))
	}

	for i, et := range types {
		select {
		case received := <-sub.Channel:
			assert.Equal(t, et, received.Type, "event %d type mismatch", i)
		case <-time.After(time.Second):
			t.Fatalf("timed out waiting for event %d (%s)", i, et)
		}
	}
}

func TestEventBus_SubscribeWithFilter(t *testing.T) {
	bus := New(DefaultConfig())
	defer bus.Close()

	// Filter that only accepts events from "scanner" source.
	sub := bus.SubscribeWithFilter(EventScanStarted, func(e *Event) bool {
		return e.Source == "scanner"
	})
	defer sub.Cancel()

	bus.Publish(NewEvent(EventScanStarted, "other", nil))
	bus.Publish(NewEvent(EventScanStarted, "scanner", "accepted"))

	select {
	case received := <-sub.Channel:
		assert.Equal(t, "scanner", received.Source)
		assert.Equal(t, "accepted", received.Payload)
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for filtered event")
	}
}

// ---------------------------------------------------------------------------
// Unsubscribe tests
// ---------------------------------------------------------------------------

func TestEventBus_Unsubscribe(t *testing.T) {
	bus := New(DefaultConfig())
	defer bus.Close()

	sub := bus.Subscribe(EventFileCreated)
	assert.Equal(t, 1, bus.SubscriberCount(EventFileCreated))

	sub.Cancel()
	assert.Equal(t, 0, bus.SubscriberCount(EventFileCreated))
}

// ---------------------------------------------------------------------------
// Metrics tests
// ---------------------------------------------------------------------------

func TestEventBus_Metrics(t *testing.T) {
	bus := New(DefaultConfig())
	defer bus.Close()

	sub := bus.Subscribe(EventScanStarted)
	defer sub.Cancel()

	bus.Publish(NewEvent(EventScanStarted, "test", nil))

	// Drain the event to ensure delivery is recorded.
	select {
	case <-sub.Channel:
	case <-time.After(time.Second):
		t.Fatal("timed out")
	}

	metrics := bus.Metrics()
	require.NotNil(t, metrics)
	assert.Equal(t, int64(1), metrics.EventsPublished)
	assert.Equal(t, int64(1), metrics.EventsDelivered)
	assert.Equal(t, int64(1), metrics.SubscribersActive)
	assert.GreaterOrEqual(t, metrics.SubscribersTotal, int64(1))
}

// ---------------------------------------------------------------------------
// Event ordering tests
// ---------------------------------------------------------------------------

func TestEventBus_EventOrdering(t *testing.T) {
	bus := New(DefaultConfig())
	defer bus.Close()

	sub := bus.Subscribe(EventFileCreated)
	defer sub.Cancel()

	count := 50
	for i := 0; i < count; i++ {
		bus.Publish(NewEvent(EventFileCreated, "watcher", i))
	}

	for i := 0; i < count; i++ {
		select {
		case received := <-sub.Channel:
			assert.Equal(t, i, received.Payload,
				"events must be delivered in publish order")
		case <-time.After(time.Second):
			t.Fatalf("timed out waiting for event %d of %d", i, count)
		}
	}
}

// ---------------------------------------------------------------------------
// Close tests
// ---------------------------------------------------------------------------

func TestEventBus_Close_StopsDelivery(t *testing.T) {
	bus := New(DefaultConfig())
	sub := bus.Subscribe(EventScanStarted)

	err := bus.Close()
	assert.NoError(t, err)

	// After close, channel should be closed.
	_, open := <-sub.Channel
	assert.False(t, open, "subscriber channel must be closed after bus.Close()")
}

func TestEventBus_Close_Idempotent(t *testing.T) {
	bus := New(DefaultConfig())
	assert.NoError(t, bus.Close())
	assert.NoError(t, bus.Close(), "Close() must be idempotent")
}

// ---------------------------------------------------------------------------
// Subscriber count tests
// ---------------------------------------------------------------------------

func TestEventBus_SubscriberCount(t *testing.T) {
	bus := New(DefaultConfig())
	defer bus.Close()

	assert.Equal(t, 0, bus.SubscriberCount(EventScanStarted))

	sub1 := bus.Subscribe(EventScanStarted)
	assert.Equal(t, 1, bus.SubscriberCount(EventScanStarted))

	sub2 := bus.Subscribe(EventScanStarted)
	assert.Equal(t, 2, bus.SubscriberCount(EventScanStarted))

	// Different type should not affect count.
	sub3 := bus.Subscribe(EventFileCreated)
	assert.Equal(t, 2, bus.SubscriberCount(EventScanStarted))
	assert.Equal(t, 1, bus.SubscriberCount(EventFileCreated))

	sub1.Cancel()
	assert.Equal(t, 1, bus.SubscriberCount(EventScanStarted))

	sub2.Cancel()
	sub3.Cancel()
}

func TestEventBus_TotalSubscribers(t *testing.T) {
	bus := New(DefaultConfig())
	defer bus.Close()

	assert.Equal(t, 0, bus.TotalSubscribers())

	sub1 := bus.Subscribe(EventScanStarted)
	sub2 := bus.Subscribe(EventFileCreated)
	assert.Equal(t, 2, bus.TotalSubscribers())

	sub1.Cancel()
	assert.Equal(t, 1, bus.TotalSubscribers())

	sub2.Cancel()
	assert.Equal(t, 0, bus.TotalSubscribers())
}

// ---------------------------------------------------------------------------
// Concurrency tests
// ---------------------------------------------------------------------------

func TestEventBus_ConcurrentPublish(t *testing.T) {
	bus := New(DefaultConfig())
	defer bus.Close()

	sub := bus.SubscribeAll()
	defer sub.Cancel()

	count := 100
	var wg sync.WaitGroup
	wg.Add(count)

	for i := 0; i < count; i++ {
		go func(n int) {
			defer wg.Done()
			bus.Publish(NewEvent(EventFileCreated, "goroutine", n))
		}(i)
	}

	wg.Wait()
	// Allow time for all publishes to be delivered.
	time.Sleep(50 * time.Millisecond)

	received := 0
	for {
		select {
		case <-sub.Channel:
			received++
		default:
			assert.Equal(t, count, received,
				"all concurrently published events must be delivered")
			return
		}
	}
}

// ---------------------------------------------------------------------------
// Middleware tests
// ---------------------------------------------------------------------------

func TestEventBus_Middleware(t *testing.T) {
	bus := New(DefaultConfig())
	defer bus.Close()

	// Add middleware that enriches events with metadata.
	bus.Use(func(e *Event) *Event {
		e.WithMetadata("enriched", "true")
		return e
	})

	sub := bus.Subscribe(EventEntityCreated)
	defer sub.Cancel()

	bus.Publish(NewEvent(EventEntityCreated, "aggregation", nil))

	select {
	case received := <-sub.Channel:
		assert.Equal(t, "true", received.Metadata["enriched"],
			"middleware must enrich events before delivery")
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for enriched event")
	}
}

func TestEventBus_Middleware_DropEvent(t *testing.T) {
	bus := New(DefaultConfig())
	defer bus.Close()

	// Middleware that drops all events.
	bus.Use(func(e *Event) *Event {
		return nil
	})

	sub := bus.Subscribe(EventScanStarted)
	defer sub.Cancel()

	bus.Publish(NewEvent(EventScanStarted, "test", nil))

	select {
	case <-sub.Channel:
		t.Fatal("event should have been dropped by middleware")
	case <-time.After(50 * time.Millisecond):
		// Expected: middleware dropped the event.
	}

	metrics := bus.Metrics()
	assert.Equal(t, int64(0), metrics.EventsPublished,
		"dropped events should not count as published")
}

// ---------------------------------------------------------------------------
// Wait tests
// ---------------------------------------------------------------------------

func TestEventBus_Wait(t *testing.T) {
	bus := New(DefaultConfig())
	defer bus.Close()

	go func() {
		time.Sleep(10 * time.Millisecond)
		bus.Publish(NewEvent(EventScanCompleted, "scanner", "done"))
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	received, err := bus.Wait(ctx, EventScanCompleted)
	require.NoError(t, err)
	require.NotNil(t, received)
	assert.Equal(t, EventScanCompleted, received.Type)
	assert.Equal(t, "done", received.Payload)
}

func TestEventBus_Wait_ContextCancelled(t *testing.T) {
	bus := New(DefaultConfig())
	defer bus.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	_, err := bus.Wait(ctx, EventScanCompleted)
	assert.Error(t, err, "Wait must return error when context is cancelled")
}
