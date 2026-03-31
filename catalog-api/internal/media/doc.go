// Package media orchestrates the media detection and analysis pipeline for the Catalogizer API.
//
// It coordinates the flow from file scanning through format detection, content analysis,
// and metadata enrichment. The pipeline processes scanned files through the detector,
// analyzer, and provider sub-packages to classify media files, extract technical metadata,
// and fetch external information from services like TMDB, OMDB, OpenLibrary, and MusicBrainz.
package media
