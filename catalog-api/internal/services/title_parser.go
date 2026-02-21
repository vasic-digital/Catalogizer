package services

import (
	vasicparser "digital.vasic.entities/pkg/parser"
)

// ParsedTitle holds structured metadata extracted from a directory or file name.
// This is a type alias for digital.vasic.entities/pkg/parser.ParsedTitle so all
// Catalogizer code that already references services.ParsedTitle continues to work.
type ParsedTitle = vasicparser.ParsedTitle

// ParseMovieTitle extracts title and year from movie directory/file names.
// Delegates to digital.vasic.entities/pkg/parser.
func ParseMovieTitle(dirname string) ParsedTitle {
	return vasicparser.ParseMovieTitle(dirname)
}

// ParseTVShow extracts show name, season, and episode from TV show directory/file names.
// Delegates to digital.vasic.entities/pkg/parser.
func ParseTVShow(dirname string) ParsedTitle {
	return vasicparser.ParseTVShow(dirname)
}

// ParseMusicAlbum extracts artist, album, and year from music directory names.
// Delegates to digital.vasic.entities/pkg/parser.
func ParseMusicAlbum(dirname string) ParsedTitle {
	return vasicparser.ParseMusicAlbum(dirname)
}

// ParseGameTitle extracts title and platform from game directory names.
// Delegates to digital.vasic.entities/pkg/parser.
func ParseGameTitle(dirname string) ParsedTitle {
	return vasicparser.ParseGameTitle(dirname)
}

// ParseSoftwareTitle extracts name and version from software directory names.
// Delegates to digital.vasic.entities/pkg/parser.
func ParseSoftwareTitle(dirname string) ParsedTitle {
	return vasicparser.ParseSoftwareTitle(dirname)
}

// CleanTitle replaces dots and underscores with spaces, trims whitespace,
// and collapses multiple consecutive spaces into one.
// Delegates to digital.vasic.entities/pkg/parser.
func CleanTitle(raw string) string {
	return vasicparser.CleanTitle(raw)
}

// ExtractYear finds a 4-digit year (1900-2099) in the given string.
// Returns nil if no valid year is found.
// Delegates to digital.vasic.entities/pkg/parser.
func ExtractYear(s string) *int {
	return vasicparser.ExtractYear(s)
}
