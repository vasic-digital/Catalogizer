// Package database provides media-specific database operations for the detection pipeline
// in the Catalogizer API.
//
// It handles persistence of media detection results, analysis metadata, and detection rules
// to the database. This package bridges the media pipeline's in-memory processing with
// durable storage, supporting the aggregation service's post-scan entity creation and the
// directory analysis system.
package database
