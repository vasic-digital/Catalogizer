package services

import (
	"testing"
)

// Movie title benchmark inputs representing common real-world naming
// conventions: parenthesized year, dotted format with quality tags,
// bracket year, bare title, and web-dl format.
var movieInputs = []string{
	"The Matrix (1999)",
	"The.Matrix.1999.1080p.BluRay",
	"Inception.2010.2160p.UHD.HDR.REMUX",
	"Some Movie Without Year",
	"Movie.Name.2023.WEB-DL.720p",
	"The Shawshank Redemption [1994]",
	"Interstellar 2014 IMAX BluRay 1080p",
	"Blade.Runner.2049.2017.2160p.UHD.BluRay.x265",
}

// TV show benchmark inputs covering S01E01, season-only, and complete
// series naming conventions.
var tvInputs = []string{
	"Breaking.Bad.S01E01.720p",
	"Game of Thrones S08E06",
	"The.Office.US.S03E15.1080p.WEB-DL",
	"Stranger Things Season 4 Episode 9",
	"Band.of.Brothers.Complete.Series",
	"Dark.S02E04.German.720p",
	"Better.Call.Saul.S06E13.PROPER.720p",
	"The.Last.of.Us.S01E03.2160p.AMZN.WEB-DL",
}

// Music album benchmark inputs with artist-album, parenthesized year,
// and slash-separated formats.
var musicInputs = []string{
	"Pink Floyd - The Dark Side of the Moon (1973)",
	"Radiohead - OK Computer (1997)",
	"Miles Davis - Kind of Blue",
	"Led Zeppelin - Led Zeppelin IV (1971)",
	"Nirvana - Nevermind (1991)",
	"The Beatles - Abbey Road (1969)",
	"Metallica - Master of Puppets (1986)",
	"Daft Punk - Random Access Memories (2013)",
}

// Game title benchmark inputs with platform and edition info.
var gameInputs = []string{
	"The Legend of Zelda - Tears of the Kingdom [Switch]",
	"Cyberpunk 2077 (PC)",
	"Red.Dead.Redemption.2.PS5",
	"Elden Ring [PS4]",
	"Baldurs.Gate.3.v4.1.1.GOG",
	"Grand Theft Auto V (Xbox One)",
	"The Witcher 3 Wild Hunt GOTY Edition",
	"Hollow Knight [Switch]",
}

// Software title benchmark inputs with version strings and platform tags.
var softwareInputs = []string{
	"Adobe Photoshop 2024 v25.3.1",
	"Visual.Studio.Code.1.87.0",
	"Firefox 123.0",
	"OBS.Studio.30.1.2",
	"VLC.Media.Player.3.0.20",
	"JetBrains IntelliJ IDEA 2024.1",
	"Blender 4.0.2",
	"GIMP 2.10.36",
}

// BenchmarkParseMovieTitle measures the cost of parsing movie directory
// names into structured title + year + quality hints. This exercises the
// regex matching pipeline for movie-format names.
func BenchmarkParseMovieTitle(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ParseMovieTitle(movieInputs[i%len(movieInputs)])
	}
}

// BenchmarkParseMovieTitle_Single measures parsing a single dotted movie
// title with quality indicators, isolating the regex cost for the most
// common file naming convention.
func BenchmarkParseMovieTitle_Single(b *testing.B) {
	input := "The.Matrix.1999.1080p.BluRay.x264-GROUP"

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ParseMovieTitle(input)
	}
}

// BenchmarkParseTVTitle measures TV show title parsing, which matches
// season/episode patterns (S01E01, 1x01, "Season N Episode N").
func BenchmarkParseTVTitle(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ParseTVShow(tvInputs[i%len(tvInputs)])
	}
}

// BenchmarkParseTVTitle_Single measures parsing a single TV episode name
// with the most common S01E01 format.
func BenchmarkParseTVTitle_Single(b *testing.B) {
	input := "Breaking.Bad.S05E16.Felina.720p.BluRay.x264"

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ParseTVShow(input)
	}
}

// BenchmarkParseMusicTitle measures music album name parsing, which
// extracts artist, album, and year from dash-separated or
// slash-separated directory names.
func BenchmarkParseMusicTitle(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ParseMusicAlbum(musicInputs[i%len(musicInputs)])
	}
}

// BenchmarkParseMusicTitle_Single measures parsing a single music album
// name in the "Artist - Album (Year)" format.
func BenchmarkParseMusicTitle_Single(b *testing.B) {
	input := "Pink Floyd - The Dark Side of the Moon (1973)"

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ParseMusicAlbum(input)
	}
}

// BenchmarkParseGameTitle measures game directory name parsing, which
// extracts title and platform from bracket or parenthesized suffixes.
func BenchmarkParseGameTitle(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ParseGameTitle(gameInputs[i%len(gameInputs)])
	}
}

// BenchmarkParseSoftwareTitle measures software name parsing, which
// extracts title and version from dotted or spaced formats.
func BenchmarkParseSoftwareTitle(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ParseSoftwareTitle(softwareInputs[i%len(softwareInputs)])
	}
}

// BenchmarkCleanTitle measures the cost of cleaning a raw title string
// by replacing dots/underscores with spaces and collapsing whitespace.
func BenchmarkCleanTitle(b *testing.B) {
	inputs := []string{
		"The.Matrix.1999.1080p.BluRay",
		"some_movie_with_underscores",
		"Already Clean Title",
		"Mixed.Dots_And.Underscores_Here",
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CleanTitle(inputs[i%len(inputs)])
	}
}

// BenchmarkExtractYear measures the cost of finding a 4-digit year in a
// string, which uses regex matching.
func BenchmarkExtractYear(b *testing.B) {
	inputs := []string{
		"The Matrix (1999)",
		"No year in this string",
		"2023 at the beginning",
		"Something.2010.Something",
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ExtractYear(inputs[i%len(inputs)])
	}
}
