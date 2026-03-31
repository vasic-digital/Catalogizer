// Package services provides infrastructure-level service implementations for the Catalogizer API.
//
// It includes the universal scanner for traversing storage roots across multiple protocols,
// the aggregation service for post-scan media entity creation, the title parser for extracting
// structured metadata from filenames, and the detection service that orchestrates the media
// analysis pipeline. These services handle the core background processing that transforms
// raw filesystem data into organized media entities.
package services
