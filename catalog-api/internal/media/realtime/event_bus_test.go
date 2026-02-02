package realtime

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// EventBus - an in-test pub/sub implementation that mirrors the channel-based
// event routing used by SMBChangeWatcher and EnhancedChangeWatcher.
// This is defined here (not in production code) because the production code
// uses raw channels; the tests validate the same pub/sub semantics.
// ---------------------------------------------------------------------------

// EventType identifies the kind of event.
type EventType string

const (
	EventFileCreated  EventType = "created"
	EventFileModified EventType = "modified"
	EventFileDeleted  EventType = "deleted"
	EventFileMoved    EventType = "moved"
)

// BusEvent is a generic event that can be published on the bus.
type BusEvent struct {
	Type      EventType
	Path      string
	Timestamp time.Time
	Payload   interface{}
}

// Subscriber is a callback that handles events.
type Subscriber func(event BusEvent)

// subscriptionEntry holds subscriber metadata.
type subscriptionEntry struct {
	id        int
	eventType EventType // empty string means subscribe to all
	fn        Subscriber
}

// EventBus implements a thread-safe publish/subscribe message bus.
type EventBus struct {
	mu          sync.RWMutex
	subscribers map[int]subscriptionEntry
	nextID      int
}

// NewEventBus creates a new EventBus.
func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[int]subscriptionEntry),
	}
}

// Subscribe registers a subscriber for a specific event type.
// Returns a subscription ID that can be used to unsubscribe.
func (eb *EventBus) Subscribe(eventType EventType, fn Subscriber) int {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	id := eb.nextID
	eb.nextID++
	eb.subscribers[id] = subscriptionEntry{
		id:        id,
		eventType: eventType,
		fn:        fn,
	}
	return id
}

// SubscribeAll registers a subscriber for all event types.
func (eb *EventBus) SubscribeAll(fn Subscriber) int {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	id := eb.nextID
	eb.nextID++
	eb.subscribers[id] = subscriptionEntry{
		id:        id,
		eventType: "", // empty = all events
		fn:        fn,
	}
	return id
}

// Unsubscribe removes a subscriber by ID.
func (eb *EventBus) Unsubscribe(id int) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	delete(eb.subscribers, id)
}

// Publish sends an event to all matching subscribers.
func (eb *EventBus) Publish(event BusEvent) {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	for _, sub := range eb.subscribers {
		if sub.eventType == "" || sub.eventType == event.Type {
			sub.fn(event)
		}
	}
}

// SubscriberCount returns the number of active subscribers.
func (eb *EventBus) SubscriberCount() int {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	return len(eb.subscribers)
}

// ---------------------------------------------------------------------------
// 1. Subscribe/publish pattern
// ---------------------------------------------------------------------------

func TestEventBus_SubscribeAndPublish(t *testing.T) {
	bus := NewEventBus()

	var received []BusEvent
	var mu sync.Mutex

	bus.Subscribe(EventFileCreated, func(event BusEvent) {
		mu.Lock()
		defer mu.Unlock()
		received = append(received, event)
	})

	event := BusEvent{
		Type:      EventFileCreated,
		Path:      "/media/movie.mkv",
		Timestamp: time.Now(),
	}
	bus.Publish(event)

	mu.Lock()
	defer mu.Unlock()
	require.Len(t, received, 1)
	assert.Equal(t, EventFileCreated, received[0].Type)
	assert.Equal(t, "/media/movie.mkv", received[0].Path)
}

func TestEventBus_SubscriberDoesNotReceiveUnrelatedEvents(t *testing.T) {
	bus := NewEventBus()

	var count int64

	bus.Subscribe(EventFileDeleted, func(event BusEvent) {
		atomic.AddInt64(&count, 1)
	})

	// Publish an event of a different type.
	bus.Publish(BusEvent{Type: EventFileCreated, Path: "/test.txt"})
	bus.Publish(BusEvent{Type: EventFileModified, Path: "/test.txt"})

	assert.Equal(t, int64(0), atomic.LoadInt64(&count))
}

func TestEventBus_SubscribeAllReceivesEverything(t *testing.T) {
	bus := NewEventBus()

	var received []BusEvent
	var mu sync.Mutex

	bus.SubscribeAll(func(event BusEvent) {
		mu.Lock()
		defer mu.Unlock()
		received = append(received, event)
	})

	bus.Publish(BusEvent{Type: EventFileCreated, Path: "/a"})
	bus.Publish(BusEvent{Type: EventFileDeleted, Path: "/b"})
	bus.Publish(BusEvent{Type: EventFileMoved, Path: "/c"})

	mu.Lock()
	defer mu.Unlock()
	assert.Len(t, received, 3)
}

func TestEventBus_PublishWithPayload(t *testing.T) {
	bus := NewEventBus()

	var capturedPayload interface{}
	bus.Subscribe(EventFileCreated, func(event BusEvent) {
		capturedPayload = event.Payload
	})

	bus.Publish(BusEvent{
		Type:    EventFileCreated,
		Path:    "/media/file.mp4",
		Payload: map[string]int64{"size": 1024},
	})

	require.NotNil(t, capturedPayload)
	payload := capturedPayload.(map[string]int64)
	assert.Equal(t, int64(1024), payload["size"])
}

// ---------------------------------------------------------------------------
// 2. Multiple subscribers receive events
// ---------------------------------------------------------------------------

func TestEventBus_MultipleSubscribersReceive(t *testing.T) {
	bus := NewEventBus()

	var count1, count2, count3 int64

	bus.Subscribe(EventFileCreated, func(event BusEvent) {
		atomic.AddInt64(&count1, 1)
	})
	bus.Subscribe(EventFileCreated, func(event BusEvent) {
		atomic.AddInt64(&count2, 1)
	})
	bus.Subscribe(EventFileCreated, func(event BusEvent) {
		atomic.AddInt64(&count3, 1)
	})

	bus.Publish(BusEvent{Type: EventFileCreated, Path: "/test.mp4"})

	assert.Equal(t, int64(1), atomic.LoadInt64(&count1))
	assert.Equal(t, int64(1), atomic.LoadInt64(&count2))
	assert.Equal(t, int64(1), atomic.LoadInt64(&count3))
}

func TestEventBus_MultipleSubscribersDifferentTypes(t *testing.T) {
	bus := NewEventBus()

	var createCount, deleteCount int64

	bus.Subscribe(EventFileCreated, func(event BusEvent) {
		atomic.AddInt64(&createCount, 1)
	})
	bus.Subscribe(EventFileDeleted, func(event BusEvent) {
		atomic.AddInt64(&deleteCount, 1)
	})

	bus.Publish(BusEvent{Type: EventFileCreated, Path: "/a"})
	bus.Publish(BusEvent{Type: EventFileCreated, Path: "/b"})
	bus.Publish(BusEvent{Type: EventFileDeleted, Path: "/c"})

	assert.Equal(t, int64(2), atomic.LoadInt64(&createCount))
	assert.Equal(t, int64(1), atomic.LoadInt64(&deleteCount))
}

func TestEventBus_ManySubscribers(t *testing.T) {
	bus := NewEventBus()
	subscriberCount := 50
	var totalReceived int64

	for i := 0; i < subscriberCount; i++ {
		bus.Subscribe(EventFileModified, func(event BusEvent) {
			atomic.AddInt64(&totalReceived, 1)
		})
	}

	bus.Publish(BusEvent{Type: EventFileModified, Path: "/shared.txt"})

	assert.Equal(t, int64(subscriberCount), atomic.LoadInt64(&totalReceived))
}

// ---------------------------------------------------------------------------
// 3. Unsubscribe stops delivery
// ---------------------------------------------------------------------------

func TestEventBus_UnsubscribeStopsDelivery(t *testing.T) {
	bus := NewEventBus()

	var count int64
	subID := bus.Subscribe(EventFileCreated, func(event BusEvent) {
		atomic.AddInt64(&count, 1)
	})

	// First event is received.
	bus.Publish(BusEvent{Type: EventFileCreated, Path: "/a"})
	assert.Equal(t, int64(1), atomic.LoadInt64(&count))

	// Unsubscribe.
	bus.Unsubscribe(subID)

	// Second event should not be received.
	bus.Publish(BusEvent{Type: EventFileCreated, Path: "/b"})
	assert.Equal(t, int64(1), atomic.LoadInt64(&count))
}

func TestEventBus_UnsubscribeOneKeepsOthers(t *testing.T) {
	bus := NewEventBus()

	var count1, count2 int64
	sub1 := bus.Subscribe(EventFileCreated, func(event BusEvent) {
		atomic.AddInt64(&count1, 1)
	})
	bus.Subscribe(EventFileCreated, func(event BusEvent) {
		atomic.AddInt64(&count2, 1)
	})

	bus.Publish(BusEvent{Type: EventFileCreated, Path: "/x"})
	assert.Equal(t, int64(1), atomic.LoadInt64(&count1))
	assert.Equal(t, int64(1), atomic.LoadInt64(&count2))

	bus.Unsubscribe(sub1)

	bus.Publish(BusEvent{Type: EventFileCreated, Path: "/y"})
	assert.Equal(t, int64(1), atomic.LoadInt64(&count1)) // unchanged
	assert.Equal(t, int64(2), atomic.LoadInt64(&count2)) // incremented
}

func TestEventBus_UnsubscribeNonExistentID(t *testing.T) {
	bus := NewEventBus()
	// Should not panic.
	bus.Unsubscribe(9999)
	assert.Equal(t, 0, bus.SubscriberCount())
}

func TestEventBus_SubscriberCountAfterUnsubscribe(t *testing.T) {
	bus := NewEventBus()

	id1 := bus.Subscribe(EventFileCreated, func(event BusEvent) {})
	id2 := bus.Subscribe(EventFileDeleted, func(event BusEvent) {})
	_ = bus.Subscribe(EventFileModified, func(event BusEvent) {})

	assert.Equal(t, 3, bus.SubscriberCount())

	bus.Unsubscribe(id1)
	assert.Equal(t, 2, bus.SubscriberCount())

	bus.Unsubscribe(id2)
	assert.Equal(t, 1, bus.SubscriberCount())
}

// ---------------------------------------------------------------------------
// 4. Event bus handles concurrent publish/subscribe
// ---------------------------------------------------------------------------

func TestEventBus_ConcurrentPublish(t *testing.T) {
	bus := NewEventBus()

	var totalReceived int64
	bus.Subscribe(EventFileCreated, func(event BusEvent) {
		atomic.AddInt64(&totalReceived, 1)
	})

	var wg sync.WaitGroup
	publishCount := 100

	for i := 0; i < publishCount; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			bus.Publish(BusEvent{
				Type: EventFileCreated,
				Path: fmt.Sprintf("/file-%d.txt", idx),
			})
		}(i)
	}

	wg.Wait()
	assert.Equal(t, int64(publishCount), atomic.LoadInt64(&totalReceived))
}

func TestEventBus_ConcurrentSubscribeAndPublish(t *testing.T) {
	bus := NewEventBus()

	var totalReceived int64
	var wg sync.WaitGroup

	// Subscribe concurrently.
	subscriberCount := 20
	for i := 0; i < subscriberCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			bus.Subscribe(EventFileModified, func(event BusEvent) {
				atomic.AddInt64(&totalReceived, 1)
			})
		}()
	}
	wg.Wait()

	assert.Equal(t, subscriberCount, bus.SubscriberCount())

	// Publish concurrently.
	publishCount := 50
	for i := 0; i < publishCount; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			bus.Publish(BusEvent{
				Type: EventFileModified,
				Path: fmt.Sprintf("/doc-%d.pdf", idx),
			})
		}(i)
	}
	wg.Wait()

	expected := int64(subscriberCount * publishCount)
	assert.Equal(t, expected, atomic.LoadInt64(&totalReceived))
}

func TestEventBus_ConcurrentSubscribeUnsubscribePublish(t *testing.T) {
	bus := NewEventBus()

	var wg sync.WaitGroup
	iterations := 100

	for i := 0; i < iterations; i++ {
		wg.Add(3)

		// Subscribe goroutine.
		go func() {
			defer wg.Done()
			id := bus.Subscribe(EventFileCreated, func(event BusEvent) {})
			// Sometimes immediately unsubscribe.
			bus.Unsubscribe(id)
		}()

		// Publish goroutine.
		go func(idx int) {
			defer wg.Done()
			bus.Publish(BusEvent{
				Type: EventFileCreated,
				Path: fmt.Sprintf("/concurrent-%d", idx),
			})
		}(i)

		// Unsubscribe goroutine (harmless if ID doesn't exist).
		go func(idx int) {
			defer wg.Done()
			bus.Unsubscribe(idx)
		}(i)
	}

	wg.Wait()
	// No panic or data race means the test passes.
}

// ---------------------------------------------------------------------------
// 5. Event types are properly routed
// ---------------------------------------------------------------------------

func TestEventBus_EventTypeRouting(t *testing.T) {
	bus := NewEventBus()

	counters := map[EventType]*int64{
		EventFileCreated:  new(int64),
		EventFileModified: new(int64),
		EventFileDeleted:  new(int64),
		EventFileMoved:    new(int64),
	}

	for eventType, counter := range counters {
		et := eventType
		cnt := counter
		bus.Subscribe(et, func(event BusEvent) {
			atomic.AddInt64(cnt, 1)
		})
	}

	// Publish various event types.
	bus.Publish(BusEvent{Type: EventFileCreated, Path: "/new.mp4"})
	bus.Publish(BusEvent{Type: EventFileCreated, Path: "/new2.mp4"})
	bus.Publish(BusEvent{Type: EventFileModified, Path: "/changed.mp4"})
	bus.Publish(BusEvent{Type: EventFileDeleted, Path: "/removed.mp4"})
	bus.Publish(BusEvent{Type: EventFileDeleted, Path: "/removed2.mp4"})
	bus.Publish(BusEvent{Type: EventFileDeleted, Path: "/removed3.mp4"})
	bus.Publish(BusEvent{Type: EventFileMoved, Path: "/moved.mp4"})

	assert.Equal(t, int64(2), atomic.LoadInt64(counters[EventFileCreated]))
	assert.Equal(t, int64(1), atomic.LoadInt64(counters[EventFileModified]))
	assert.Equal(t, int64(3), atomic.LoadInt64(counters[EventFileDeleted]))
	assert.Equal(t, int64(1), atomic.LoadInt64(counters[EventFileMoved]))
}

func TestEventBus_RoutingWithMixedSubscribers(t *testing.T) {
	bus := NewEventBus()

	var specificCount, allCount int64

	// Specific subscriber for created events.
	bus.Subscribe(EventFileCreated, func(event BusEvent) {
		atomic.AddInt64(&specificCount, 1)
	})

	// Catch-all subscriber.
	bus.SubscribeAll(func(event BusEvent) {
		atomic.AddInt64(&allCount, 1)
	})

	bus.Publish(BusEvent{Type: EventFileCreated, Path: "/a"})
	bus.Publish(BusEvent{Type: EventFileDeleted, Path: "/b"})
	bus.Publish(BusEvent{Type: EventFileModified, Path: "/c"})

	assert.Equal(t, int64(1), atomic.LoadInt64(&specificCount))
	assert.Equal(t, int64(3), atomic.LoadInt64(&allCount))
}

func TestEventBus_UnknownEventTypeIsIgnored(t *testing.T) {
	bus := NewEventBus()

	var count int64
	bus.Subscribe(EventFileCreated, func(event BusEvent) {
		atomic.AddInt64(&count, 1)
	})

	// Publish an event with an unregistered type.
	bus.Publish(BusEvent{Type: "unknown_type", Path: "/test"})

	assert.Equal(t, int64(0), atomic.LoadInt64(&count))
}

func TestEventBus_EmptyBusPublish(t *testing.T) {
	bus := NewEventBus()

	// Publishing to an empty bus should not panic.
	bus.Publish(BusEvent{Type: EventFileCreated, Path: "/test"})
	assert.Equal(t, 0, bus.SubscriberCount())
}

// ---------------------------------------------------------------------------
// Integration with ChangeEvent channel pattern
// ---------------------------------------------------------------------------

func TestChangeEventChannel_ProducerConsumer(t *testing.T) {
	// Test the same channel-based pub/sub pattern used by SMBChangeWatcher.
	queue := make(chan ChangeEvent, 100)

	var wg sync.WaitGroup
	var received []ChangeEvent
	var mu sync.Mutex

	// Consumer
	wg.Add(1)
	go func() {
		defer wg.Done()
		for event := range queue {
			mu.Lock()
			received = append(received, event)
			mu.Unlock()
		}
	}()

	// Producer
	events := []ChangeEvent{
		{Path: "/a.txt", Operation: "created", Timestamp: time.Now()},
		{Path: "/b.txt", Operation: "modified", Timestamp: time.Now()},
		{Path: "/c.txt", Operation: "deleted", Timestamp: time.Now()},
	}
	for _, e := range events {
		queue <- e
	}
	close(queue)

	wg.Wait()

	mu.Lock()
	defer mu.Unlock()
	require.Len(t, received, 3)
	assert.Equal(t, "created", received[0].Operation)
	assert.Equal(t, "modified", received[1].Operation)
	assert.Equal(t, "deleted", received[2].Operation)
}

func TestEnhancedChangeEventChannel_ConcurrentProducers(t *testing.T) {
	// Test the EnhancedChangeEvent channel under concurrent producer load.
	queue := make(chan EnhancedChangeEvent, 1000)
	producerCount := 10
	eventsPerProducer := 50
	totalExpected := producerCount * eventsPerProducer

	var wg sync.WaitGroup
	var received int64

	// Consumer
	done := make(chan struct{})
	go func() {
		for range queue {
			atomic.AddInt64(&received, 1)
			if atomic.LoadInt64(&received) == int64(totalExpected) {
				close(done)
				return
			}
		}
	}()

	// Producers
	for p := 0; p < producerCount; p++ {
		wg.Add(1)
		go func(producerID int) {
			defer wg.Done()
			for i := 0; i < eventsPerProducer; i++ {
				queue <- EnhancedChangeEvent{
					Path:      fmt.Sprintf("/producer-%d/file-%d.txt", producerID, i),
					SmbRoot:   "test_root",
					Operation: "created",
					Timestamp: time.Now(),
					Size:      int64(i * 1024),
					IsDir:     false,
				}
			}
		}(p)
	}

	wg.Wait()

	// Wait for consumer to finish with a timeout.
	select {
	case <-done:
		// OK
	case <-time.After(5 * time.Second):
		t.Fatalf("Timed out waiting for consumer; received %d of %d events",
			atomic.LoadInt64(&received), totalExpected)
	}

	assert.Equal(t, int64(totalExpected), atomic.LoadInt64(&received))
}

func TestChangeEventChannel_FullQueueDrop(t *testing.T) {
	// Mirrors the debounceChange pattern where events are dropped when the queue is full.
	queue := make(chan ChangeEvent, 5)

	// Fill the queue.
	for i := 0; i < 5; i++ {
		queue <- ChangeEvent{
			Path:      fmt.Sprintf("/file-%d", i),
			Operation: "created",
			Timestamp: time.Now(),
		}
	}

	// Attempt to send one more with non-blocking select (as in the production code).
	dropped := false
	event := ChangeEvent{Path: "/overflow", Operation: "created", Timestamp: time.Now()}
	select {
	case queue <- event:
		// Should not reach here since queue is full.
		t.Error("Expected queue to be full")
	default:
		dropped = true
	}

	assert.True(t, dropped, "Event should have been dropped when queue is full")
}
