// Package detector provides format detection for media files based on file extensions and
// magic bytes.
//
// It identifies the media type (movie, TV show, music, game, software, book, etc.) of scanned
// files using a combination of extension mapping, file header analysis, and configurable
// detection rules. The detector is the first stage of the media pipeline, classifying files
// before they proceed to analysis and metadata enrichment.
package detector
