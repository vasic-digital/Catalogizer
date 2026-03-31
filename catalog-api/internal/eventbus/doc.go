// Package eventbus provides an in-process publish/subscribe event bus for decoupled
// communication between services in the Catalogizer API.
//
// It enables services to emit events (such as scan completion, file changes, or media
// detection results) without direct dependencies on consumers. Subscribers register handlers
// for specific event types, allowing the real-time WebSocket layer, aggregation service, and
// other components to react to system events asynchronously.
package eventbus
