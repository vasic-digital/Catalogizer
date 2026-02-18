package manager

import (
	"context"
	"fmt"
	"io"
	"sync"

	"digital.vasic.assets/pkg/event"
	"digital.vasic.assets/pkg/resolver"
	"digital.vasic.assets/pkg/store"
)

// workItem represents a pending resolution task.
type workItem struct {
	request *resolver.ResolveRequest
}

// workerPool manages background asset resolution goroutines.
type workerPool struct {
	queue    chan workItem
	resolver resolver.Resolver
	store    store.Store
	eventBus event.EventBus
	logger   io.Writer
	wg       sync.WaitGroup
	ctx      context.Context
	cancel   context.CancelFunc
}

func newWorkerPool(count int, r resolver.Resolver, s store.Store, bus event.EventBus, logger io.Writer) *workerPool {
	ctx, cancel := context.WithCancel(context.Background())
	wp := &workerPool{
		queue:    make(chan workItem, 256),
		resolver: r,
		store:    s,
		eventBus: bus,
		logger:   logger,
		ctx:      ctx,
		cancel:   cancel,
	}

	for i := 0; i < count; i++ {
		wp.wg.Add(1)
		go wp.run()
	}

	return wp
}

func (wp *workerPool) submit(item workItem) {
	select {
	case wp.queue <- item:
	case <-wp.ctx.Done():
	}
}

func (wp *workerPool) run() {
	defer wp.wg.Done()

	for {
		select {
		case <-wp.ctx.Done():
			return
		case item, ok := <-wp.queue:
			if !ok {
				return
			}
			wp.process(item)
		}
	}
}

func (wp *workerPool) process(item workItem) {
	req := item.request

	wp.publishEvent(event.AssetResolving, req)

	result, err := wp.resolver.Resolve(wp.ctx, req)
	if err != nil {
		wp.logf("resolve failed for %s: %v", req.AssetID, err)
		wp.publishEvent(event.AssetFailed, req)
		return
	}
	defer result.Content.Close()

	err = wp.store.Put(wp.ctx, req.AssetID, result.Content, &store.Info{
		ContentType: result.ContentType,
		Size:        result.Size,
	})
	if err != nil {
		wp.logf("store failed for %s: %v", req.AssetID, err)
		wp.publishEvent(event.AssetFailed, req)
		return
	}

	wp.publishEvent(event.AssetReady, req)
}

func (wp *workerPool) publishEvent(eventType event.Type, req *resolver.ResolveRequest) {
	if wp.eventBus == nil {
		return
	}
	wp.eventBus.Publish(event.Event{
		Type:      eventType,
		AssetID:   req.AssetID,
		AssetType: req.AssetType,
		Metadata: map[string]string{
			"entity_type": req.EntityType,
			"entity_id":   req.EntityID,
		},
	})
}

func (wp *workerPool) logf(format string, args ...interface{}) {
	if wp.logger != nil {
		fmt.Fprintf(wp.logger, "[assets] "+format+"\n", args...)
	}
}

func (wp *workerPool) stop() {
	wp.cancel()
	close(wp.queue)
	wp.wg.Wait()
}
