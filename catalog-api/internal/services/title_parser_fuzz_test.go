package services

import (
	"testing"
)

func FuzzParseMovieTitle(f *testing.F) {
	seeds := []string{
		"",
		"The Matrix (1999)",
		"Inception.2010.1080p.BluRay.x264",
		"Movie.Title.2023.720p.WEB-DL",
		"Some.Movie.REMASTERED.2020.2160p.UHD",
		"A Movie",
		"2001 A Space Odyssey",
		"(2020)",
		"....",
		"The_Dark_Knight_2008_BDRip",
		"Back to the Future Part II 1989",
		"Файл Фильма (2023)",
		"\x00\x01\x02",
		"movie (1899)",
		"movie (2100)",
		"    spaces    ",
	}

	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, input string) {
		result := ParseMovieTitle(input)

		// Invariant 1: must not panic (implicit)

		// Invariant 2: if year is set, it must be in valid range
		if result.Year != nil {
			y := *result.Year
			if y < 1900 || y > 2099 {
				t.Errorf("ParseMovieTitle(%q).Year = %d, outside 1900-2099", input, y)
			}
		}
	})
}

func FuzzParseTVShow(f *testing.F) {
	seeds := []string{
		"",
		"Breaking Bad S01E01",
		"The.Office.S03E05.720p",
		"Game of Thrones - S08E06 - The Iron Throne",
		"show.name.1x01.pilot",
		"Show S1E1",
		"Show S99E99",
		"Show Season 1 Episode 1",
		"S01E01",
		"show_name_s02e03_720p",
		"\x00\x01\x02",
		"   ",
	}

	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, input string) {
		result := ParseTVShow(input)

		// Invariant 1: must not panic (implicit)

		// Invariant 2: season must be non-negative if set
		if result.Season != nil && *result.Season < 0 {
			t.Errorf("ParseTVShow(%q).Season = %d, negative", input, *result.Season)
		}

		// Invariant 3: episode must be non-negative if set
		if result.Episode != nil && *result.Episode < 0 {
			t.Errorf("ParseTVShow(%q).Episode = %d, negative", input, *result.Episode)
		}

		// Invariant 4: if year is set, it must be in valid range
		if result.Year != nil {
			y := *result.Year
			if y < 1900 || y > 2099 {
				t.Errorf("ParseTVShow(%q).Year = %d, outside 1900-2099", input, y)
			}
		}
	})
}

func FuzzParseMusicAlbum(f *testing.F) {
	seeds := []string{
		"",
		"Pink Floyd - The Dark Side of the Moon (1973)",
		"Artist - Album (2020)",
		"Artist_-_Album_2019",
		"Unknown Album",
		"(1999) Album Name",
		"artist - album [FLAC]",
		"\x00\x01\x02",
		"    ",
		" - ",
	}

	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, input string) {
		result := ParseMusicAlbum(input)

		// Invariant 1: must not panic (implicit)

		// Invariant 2: if year is set, it must be in valid range
		if result.Year != nil {
			y := *result.Year
			if y < 1900 || y > 2099 {
				t.Errorf("ParseMusicAlbum(%q).Year = %d, outside 1900-2099", input, y)
			}
		}
	})
}

func FuzzParseGameTitle(f *testing.F) {
	seeds := []string{
		"",
		"The Witcher 3 Wild Hunt (PC)",
		"Game Name [PS5]",
		"game.name.xbox.2023",
		"Game (Nintendo Switch)",
		"Game",
		"\x00\x01\x02",
	}

	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, input string) {
		result := ParseGameTitle(input)
		// Must not panic
		_ = result
	})
}

func FuzzParseSoftwareTitle(f *testing.F) {
	seeds := []string{
		"",
		"Adobe Photoshop v23.0",
		"software-name-1.2.3",
		"App_v2.0_x64",
		"Program (x86)",
		"\x00\x01\x02",
	}

	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, input string) {
		result := ParseSoftwareTitle(input)
		// Must not panic
		_ = result
	})
}

func FuzzCleanTitle(f *testing.F) {
	seeds := []string{
		"",
		"hello.world",
		"hello_world",
		"hello.world.2023.1080p",
		"already clean title",
		"   lots   of   spaces   ",
		"...",
		"___",
		"mixed.dots_and_underscores",
		"\x00\x01\x02",
		"\t\n\r",
	}

	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, input string) {
		result := CleanTitle(input)
		// Must not panic
		_ = result
	})
}

func FuzzExtractYear(f *testing.F) {
	seeds := []string{
		"",
		"2023",
		"movie (2023)",
		"1900",
		"2099",
		"1899",
		"2100",
		"no year here",
		"1234",
		"9999",
		"20232023",
	}

	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, input string) {
		result := ExtractYear(input)

		// Invariant: if a year is returned, it must be in 1900-2099
		if result != nil {
			y := *result
			if y < 1900 || y > 2099 {
				t.Errorf("ExtractYear(%q) = %d, outside 1900-2099", input, y)
			}
		}
	})
}
