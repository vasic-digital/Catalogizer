package event

import (
	"sync"
	"sync/atomic"
	"testing"

	"digital.vasic.assets/pkg/asset"
	"github.com/stretchr/testify/assert"
)

func TestInMemoryBus_PublishSubscribe(t *testing.T) {
	bus := NewInMemoryBus()
	received := make(chan Event, 1)

	bus.Subscribe(func(evt Event) {
		received <- evt
	})

	bus.Publish(Event{
		Type:    AssetReady,
		AssetID: "test-id",
	})

	evt := <-received
	assert.Equal(t, AssetReady, evt.Type)
	assert.Equal(t, asset.ID("test-id"), evt.AssetID)
}

func TestInMemoryBus_MultipleSubscribers(t *testing.T) {
	bus := NewInMemoryBus()
	var count atomic.Int32

	for i := 0; i < 3; i++ {
		bus.Subscribe(func(evt Event) {
			count.Add(1)
		})
	}

	bus.Publish(Event{Type: AssetReady})

	assert.Equal(t, int32(3), count.Load())
}

func TestInMemoryBus_Unsubscribe(t *testing.T) {
	bus := NewInMemoryBus()
	var count atomic.Int32

	unsub := bus.Subscribe(func(evt Event) {
		count.Add(1)
	})

	bus.Publish(Event{Type: AssetReady})
	assert.Equal(t, int32(1), count.Load())

	unsub()

	bus.Publish(Event{Type: AssetReady})
	assert.Equal(t, int32(1), count.Load())
}

func TestInMemoryBus_ConcurrentPublish(t *testing.T) {
	bus := NewInMemoryBus()
	var count atomic.Int32

	bus.Subscribe(func(evt Event) {
		count.Add(1)
	})

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			bus.Publish(Event{Type: AssetReady})
		}()
	}
	wg.Wait()

	assert.Equal(t, int32(100), count.Load())
}
