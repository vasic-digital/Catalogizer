package handlers

import (
	"catalogizer/internal/eventbus"
	"sync"

	root_handlers "catalogizer/handlers"

	"go.uber.org/zap"
)

// EventBusBridge subscribes to system eventbus events and forwards them to
// WebSocket clients via the existing root WebSocket handler.
//
// The bridge reads from the eventbus subscription channel in a background
// goroutine and calls wsHandler.BroadcastToClients for each matching event.
// Only events with client-facing types (scan.*, entity.*, file.*) are forwarded.
// Publication is non-blocking: BroadcastToClients drops messages for clients
// whose send buffer is full rather than blocking the bridge.
type EventBusBridge struct {
	bus       *eventbus.EventBus
	wsHandler *root_handlers.WebSocketHandler
	logger    *zap.Logger

	subCancel  func()
	closeOnce  sync.Once
	stopCh     chan struct{}
	routingMap map[eventbus.EventType]string
}

// NewEventBusBridge creates a new bridge that subscribes to the given eventbus
// and forwards relevant events to WebSocket clients.
func NewEventBusBridge(bus *eventbus.EventBus, wsHandler *root_handlers.WebSocketHandler, logger *zap.Logger) *EventBusBridge {
	if logger == nil {
		logger = zap.NewNop()
	}

	b := &EventBusBridge{
		bus:       bus,
		wsHandler: wsHandler,
		logger:    logger,
		stopCh:    make(chan struct{}),
		// Map of eventbus event types to the WebSocket message "type" field.
		// Only events listed here are forwarded to clients.
		routingMap: map[eventbus.EventType]string{
			eventbus.EventScanCompleted: "scan_completed",
			eventbus.EventScanFailed:    "scan_failed",
			eventbus.EventScanStarted:   "scan_started",
			eventbus.EventFileCreated:   "file_created",
			eventbus.EventFileModified:  "file_modified",
			eventbus.EventFileDeleted:   "file_deleted",
			eventbus.EventFileMoved:     "file_moved",
			eventbus.EventEntityCreated: "entity_created",
			eventbus.EventEntityUpdated: "entity_updated",
		},
	}

	return b
}

// Start subscribes to the eventbus and begins forwarding events to WebSocket clients.
func (b *EventBusBridge) Start() {
	sub := b.bus.SubscribeMultiple(
		eventbus.EventScanCompleted,
		eventbus.EventScanFailed,
		eventbus.EventScanStarted,
		eventbus.EventFileCreated,
		eventbus.EventFileModified,
		eventbus.EventFileDeleted,
		eventbus.EventFileMoved,
		eventbus.EventEntityCreated,
		eventbus.EventEntityUpdated,
	)
	b.subCancel = sub.Cancel

	b.logger.Info("EventBus bridge started, forwarding events to WebSocket clients")

	go b.forwardLoop(sub)
}

// Stop gracefully stops the bridge and unsubscribes from the eventbus.
func (b *EventBusBridge) Stop() {
	b.closeOnce.Do(func() {
		close(b.stopCh)
		if b.subCancel != nil {
			b.subCancel()
		}
		b.logger.Info("EventBus bridge stopped")
	})
}

// forwardLoop reads events from the subscription channel and forwards them
// to connected WebSocket clients.
func (b *EventBusBridge) forwardLoop(sub *eventbus.Subscription) {
	for {
		select {
		case <-b.stopCh:
			return
		case evt, ok := <-sub.Channel:
			if !ok {
				b.logger.Warn("EventBus subscription channel closed")
				return
			}
			b.forwardEvent(evt)
		}
	}
}

// forwardEvent converts a system event to a WebSocket JSON message and
// broadcasts it to all connected clients.
func (b *EventBusBridge) forwardEvent(evt *eventbus.Event) {
	wsType, ok := b.routingMap[evt.Type]
	if !ok {
		return // Unknown event type — don't forward
	}

	msg := map[string]interface{}{
		"type":       wsType,
		"event_type": string(evt.Type),
		"source":     evt.Source,
		"timestamp":  evt.Timestamp,
		"trace_id":   evt.TraceID,
		"payload":    evt.Payload,
	}

	// Merge metadata into the message payload when present
	if len(evt.Metadata) > 0 {
		meta := make(map[string]string)
		for k, v := range evt.Metadata {
			meta[k] = v
		}
		msg["metadata"] = meta
	}

	b.wsHandler.BroadcastToClients(msg)

	b.logger.Debug("Forwarded event to WebSocket clients",
		zap.String("event_type", string(evt.Type)),
		zap.String("ws_type", wsType))
}
