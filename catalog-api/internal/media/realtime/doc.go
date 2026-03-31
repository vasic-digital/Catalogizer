// Package realtime provides real-time file watching with debounced change detection for the
// Catalogizer API.
//
// It monitors storage roots for filesystem changes and publishes events through the event bus
// to notify connected WebSocket clients. Change detection is debounced to avoid flooding
// subscribers with rapid successive updates. This package bridges the filesystem watcher with
// the WebSocket layer, enabling live updates in the web and desktop frontends.
package realtime
