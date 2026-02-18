package services

import (
	"regexp"
	"strconv"
	"strings"
)

// ParsedTitle holds structured metadata extracted from a directory or file name.
type ParsedTitle struct {
	Title        string
	Year         *int
	QualityHints []string
	Season       *int
	Episode      *int
	Artist       string
	Album        string
	TrackNumber  *int
	Platform     string
	Version      string
}

var (
	// Year patterns
	yearParenRe   = regexp.MustCompile(`\((\d{4})\)`)
	yearBracketRe = regexp.MustCompile(`\[(\d{4})\]`)
	yearInlineRe  = regexp.MustCompile(`(?:^|[\s._-])(\d{4})(?:[\s._-]|$)`)

	// Movie: "The Matrix (1999)", "The.Matrix.1999.1080p.BluRay", "The Matrix [1999]"
	movieYearParenRe   = regexp.MustCompile(`^(.+?)[\s._-]*\((\d{4})\)`)
	movieYearBracketRe = regexp.MustCompile(`^(.+?)[\s._-]*\[(\d{4})\]`)
	movieYearDotRe     = regexp.MustCompile(`^(.+?)[\s._-]+(\d{4})[\s._-]`)

	// TV Show patterns
	tvSxxExxRe   = regexp.MustCompile(`(?i)^(.+?)[\s._-]+S(\d{1,2})E(\d{1,2})`)
	tvNxNNRe     = regexp.MustCompile(`(?i)^(.+?)[\s._-]+(\d{1,2})x(\d{2,3})`)
	tvSeasonRe   = regexp.MustCompile(`(?i)^(.+?)[\s._-]+(?:Season|S)[\s._-]*(\d{1,2})(?:[\s._-]+Episode[\s._-]*(\d{1,3}))?`)
	tvCompleteRe = regexp.MustCompile(`(?i)^(.+?)[\s._-]+(?:Complete|COMPLETE)`)

	// Music: "Artist - Album (Year)", path separator "Artist/Album"
	musicDashRe  = regexp.MustCompile(`^(.+?)\s*-\s*(.+?)(?:\s*\((\d{4})\)\s*)?$`)
	musicSlashRe = regexp.MustCompile(`^([^/]+)/([^/]+)$`)

	// Quality indicators
	qualityIndicators = []struct {
		label string
		re    *regexp.Regexp
	}{
		{"2160p", regexp.MustCompile(`(?i)2160p`)},
		{"4K", regexp.MustCompile(`(?i)\b4K\b`)},
		{"1080p", regexp.MustCompile(`(?i)1080p`)},
		{"720p", regexp.MustCompile(`(?i)720p`)},
		{"480p", regexp.MustCompile(`(?i)480p`)},
		{"BluRay", regexp.MustCompile(`(?i)(?:Blu[\s._-]?Ray|BDRip|BRRip)`)},
		{"WEB-DL", regexp.MustCompile(`(?i)(?:WEB[\s._-]*DL|WEBRip)`)},
		{"HDRip", regexp.MustCompile(`(?i)HDRip`)},
		{"DVDRip", regexp.MustCompile(`(?i)(?:DVDRip|DVD[\s._-]?Rip)`)},
		{"REMUX", regexp.MustCompile(`(?i)REMUX`)},
		{"HDR", regexp.MustCompile(`(?i)\bHDR(?:10)?\b`)},
		{"DTS", regexp.MustCompile(`(?i)\bDTS\b`)},
		{"Atmos", regexp.MustCompile(`(?i)\bAtmos\b`)},
	}

	// Game platform indicators
	gamePlatformRe = regexp.MustCompile(`(?i)\b(?:PC|Windows|Linux|Mac|macOS|PS[2-5]|PlayStation[\s._-]*[2-5]?|Xbox(?:[\s._-]*(?:One|360|Series[\s._-]*[XS]))?|Switch|Nintendo|GOG|Steam)\b`)

	// Software version: "v3.0.20", "3.0.20", "24.04"
	softwareVersionRe = regexp.MustCompile(`(?:^|[\s._-])v?(\d+(?:\.\d+)+)(?:[\s._-]|$)`)

	// Track number: leading "01 - ", "01.", "Track 01"
	trackNumberRe = regexp.MustCompile(`(?i)(?:^|[\s._-])(?:Track[\s._-]*)?(0?\d{1,2})(?:[\s._-]+|$)`)

	// Dots/underscores replacement
	dotUnderscoreRe = regexp.MustCompile(`[._]+`)

	// Collapse multiple spaces
	multiSpaceRe = regexp.MustCompile(`\s{2,}`)
)

// ParseMovieTitle extracts title and year from movie directory/file names.
// Handles patterns like "The Matrix (1999)", "The.Matrix.1999.1080p.BluRay",
// "The Matrix [1999]".
func ParseMovieTitle(dirname string) ParsedTitle {
	var result ParsedTitle

	// Try parenthesized year: "Title (2023)"
	if m := movieYearParenRe.FindStringSubmatch(dirname); m != nil {
		result.Title = CleanTitle(m[1])
		if y, err := strconv.Atoi(m[2]); err == nil && y >= 1900 && y <= 2099 {
			result.Year = &y
		}
	} else if m := movieYearBracketRe.FindStringSubmatch(dirname); m != nil {
		// Try bracketed year: "Title [2023]"
		result.Title = CleanTitle(m[1])
		if y, err := strconv.Atoi(m[2]); err == nil && y >= 1900 && y <= 2099 {
			result.Year = &y
		}
	} else if m := movieYearDotRe.FindStringSubmatch(dirname); m != nil {
		// Try dotted year: "The.Matrix.1999.1080p"
		result.Title = CleanTitle(m[1])
		if y, err := strconv.Atoi(m[2]); err == nil && y >= 1900 && y <= 2099 {
			result.Year = &y
		}
	} else {
		result.Title = CleanTitle(dirname)
		result.Year = ExtractYear(dirname)
	}

	result.QualityHints = extractQualityHints(dirname)
	return result
}

// ParseTVShow extracts show name, season, and episode from TV show directory/file names.
// Handles patterns like "Breaking.Bad.S01E02", "Breaking Bad - Season 1",
// "S01E02 - Pilot", "01x02".
func ParseTVShow(dirname string) ParsedTitle {
	var result ParsedTitle

	// Try S01E02 format
	if m := tvSxxExxRe.FindStringSubmatch(dirname); m != nil {
		result.Title = CleanTitle(m[1])
		if s, err := strconv.Atoi(m[2]); err == nil {
			result.Season = &s
		}
		if e, err := strconv.Atoi(m[3]); err == nil {
			result.Episode = &e
		}
	} else if m := tvNxNNRe.FindStringSubmatch(dirname); m != nil {
		// Try 1x02 format
		result.Title = CleanTitle(m[1])
		if s, err := strconv.Atoi(m[2]); err == nil {
			result.Season = &s
		}
		if e, err := strconv.Atoi(m[3]); err == nil {
			result.Episode = &e
		}
	} else if m := tvSeasonRe.FindStringSubmatch(dirname); m != nil {
		// Try "Season 1" or "Season 1 Episode 2" format
		result.Title = CleanTitle(m[1])
		if s, err := strconv.Atoi(m[2]); err == nil {
			result.Season = &s
		}
		if len(m) > 3 && m[3] != "" {
			if e, err := strconv.Atoi(m[3]); err == nil {
				result.Episode = &e
			}
		}
	} else if m := tvCompleteRe.FindStringSubmatch(dirname); m != nil {
		// Try "Complete" suffix
		result.Title = CleanTitle(m[1])
	} else {
		result.Title = CleanTitle(dirname)
	}

	result.QualityHints = extractQualityHints(dirname)
	return result
}

// ParseMusicAlbum extracts artist, album, and year from music directory names.
// Handles patterns like "Pink Floyd - The Wall (1979)", "Pink Floyd/The Wall".
func ParseMusicAlbum(dirname string) ParsedTitle {
	var result ParsedTitle

	// Try "Artist - Album (Year)" format
	if m := musicDashRe.FindStringSubmatch(dirname); m != nil {
		result.Artist = strings.TrimSpace(m[1])
		album := strings.TrimSpace(m[2])
		// Remove trailing year in parentheses from album if captured separately
		album = yearParenRe.ReplaceAllString(album, "")
		result.Album = strings.TrimSpace(album)
		result.Title = result.Album
		if m[3] != "" {
			if y, err := strconv.Atoi(m[3]); err == nil && y >= 1900 && y <= 2099 {
				result.Year = &y
			}
		}
	} else if m := musicSlashRe.FindStringSubmatch(dirname); m != nil {
		// Try "Artist/Album" format (path separator)
		result.Artist = strings.TrimSpace(m[1])
		result.Album = strings.TrimSpace(m[2])
		result.Title = result.Album
		result.Year = ExtractYear(dirname)
	} else {
		result.Title = CleanTitle(dirname)
		result.Year = ExtractYear(dirname)
	}

	return result
}

// ParseGameTitle extracts title and platform from game directory names.
// Handles patterns like "Half-Life 2 (PC)", "The Legend of Zelda [Switch]".
func ParseGameTitle(dirname string) ParsedTitle {
	var result ParsedTitle

	// Extract platform
	if m := gamePlatformRe.FindString(dirname); m != "" {
		result.Platform = m
	}

	// Remove platform in parentheses or brackets for cleaner title parsing
	cleaned := dirname
	platformParenRe := regexp.MustCompile(`(?i)\s*[\(\[](?:PC|Windows|Linux|Mac|macOS|PS[2-5]|PlayStation[\s._-]*[2-5]?|Xbox(?:[\s._-]*(?:One|360|Series[\s._-]*[XS]))?|Switch|Nintendo|GOG|Steam)[\)\]]\s*`)
	cleaned = platformParenRe.ReplaceAllString(cleaned, " ")
	cleaned = strings.TrimSpace(cleaned)

	// Try year extraction like movies
	if m := movieYearParenRe.FindStringSubmatch(cleaned); m != nil {
		result.Title = CleanTitle(m[1])
		if y, err := strconv.Atoi(m[2]); err == nil && y >= 1900 && y <= 2099 {
			result.Year = &y
		}
	} else if m := movieYearBracketRe.FindStringSubmatch(cleaned); m != nil {
		result.Title = CleanTitle(m[1])
		if y, err := strconv.Atoi(m[2]); err == nil && y >= 1900 && y <= 2099 {
			result.Year = &y
		}
	} else if m := movieYearDotRe.FindStringSubmatch(cleaned); m != nil {
		result.Title = CleanTitle(m[1])
		if y, err := strconv.Atoi(m[2]); err == nil && y >= 1900 && y <= 2099 {
			result.Year = &y
		}
	} else {
		result.Title = CleanTitle(cleaned)
		result.Year = ExtractYear(cleaned)
	}

	return result
}

// ParseSoftwareTitle extracts name and version from software directory names.
// Handles patterns like "Ubuntu 24.04", "Microsoft Office 2021", "VLC 3.0.20".
func ParseSoftwareTitle(dirname string) ParsedTitle {
	var result ParsedTitle

	// Extract version
	if m := softwareVersionRe.FindStringSubmatch(dirname); m != nil {
		result.Version = m[1]
	}

	// Extract platform
	if m := gamePlatformRe.FindString(dirname); m != "" {
		result.Platform = m
	}

	// Try year extraction
	if m := movieYearParenRe.FindStringSubmatch(dirname); m != nil {
		result.Title = CleanTitle(m[1])
		if y, err := strconv.Atoi(m[2]); err == nil && y >= 1900 && y <= 2099 {
			result.Year = &y
		}
	} else {
		// Remove version from title
		cleaned := dirname
		if result.Version != "" {
			versionEscaped := regexp.QuoteMeta(result.Version)
			vRe := regexp.MustCompile(`(?:^|[\s._-])v?` + versionEscaped + `(?:[\s._-]|$)`)
			cleaned = vRe.ReplaceAllString(cleaned, " ")
		}
		result.Title = CleanTitle(cleaned)
		result.Year = ExtractYear(dirname)
	}

	return result
}

// CleanTitle replaces dots and underscores with spaces, trims whitespace,
// and collapses multiple consecutive spaces into one.
func CleanTitle(raw string) string {
	s := dotUnderscoreRe.ReplaceAllString(raw, " ")
	s = strings.TrimSpace(s)
	// Remove trailing quality/source tags
	for _, qi := range qualityIndicators {
		s = qi.re.ReplaceAllString(s, "")
	}
	s = strings.TrimRight(s, " -._[](){}|")
	s = multiSpaceRe.ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}

// ExtractYear finds a 4-digit year (1900-2099) in the given string.
// Returns nil if no valid year is found.
func ExtractYear(s string) *int {
	// Try parenthesized year first: (2023)
	if m := yearParenRe.FindStringSubmatch(s); m != nil {
		if y, err := strconv.Atoi(m[1]); err == nil && y >= 1900 && y <= 2099 {
			return &y
		}
	}
	// Try bracketed year: [2023]
	if m := yearBracketRe.FindStringSubmatch(s); m != nil {
		if y, err := strconv.Atoi(m[1]); err == nil && y >= 1900 && y <= 2099 {
			return &y
		}
	}
	// Try inline year: separated by dots, spaces, or dashes
	if m := yearInlineRe.FindStringSubmatch(s); m != nil {
		if y, err := strconv.Atoi(m[1]); err == nil && y >= 1900 && y <= 2099 {
			return &y
		}
	}
	return nil
}

// extractQualityHints finds quality indicator strings in a name.
func extractQualityHints(name string) []string {
	var hints []string
	seen := make(map[string]bool)
	for _, qi := range qualityIndicators {
		if qi.re.MatchString(name) && !seen[qi.label] {
			hints = append(hints, qi.label)
			seen[qi.label] = true
		}
	}
	return hints
}
