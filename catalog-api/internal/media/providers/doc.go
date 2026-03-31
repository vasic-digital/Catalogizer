// Package providers implements external metadata providers for enriching media entities
// with information from third-party services.
//
// Fully implemented providers include TMDB and OMDB for movies and TV shows, OpenLibrary
// for books, and MusicBrainz for music. Additional providers (IGDB, GiantBomb, etc.) have
// graceful degradation: missing API keys or unavailable services do not block the pipeline.
// Each provider implements a common interface enabling uniform access from the media pipeline.
package providers
