package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// =============================================================================
// Helper: create provider for tests
// =============================================================================

func newTestBookProvider(t *testing.T) *BookRecognitionProvider {
	t.Helper()
	logger, _ := zap.NewDevelopment()
	return NewBookRecognitionProvider(logger)
}

// =============================================================================
// extractTitleFromOCR Tests
// =============================================================================

func TestBookRecognitionProvider_ExtractTitleFromOCR_LargestBlock(t *testing.T) {
	p := newTestBookProvider(t)

	blocks := []OCRTextBlock{
		{Text: "small text", Confidence: 0.9, BoundingBox: OCRBoundingBox{Width: 50, Height: 20}},
		{Text: "The Great Gatsby", Confidence: 0.95, BoundingBox: OCRBoundingBox{Width: 300, Height: 100}},
		{Text: "F. Scott Fitzgerald", Confidence: 0.9, BoundingBox: OCRBoundingBox{Width: 100, Height: 30}},
	}

	title := p.extractTitleFromOCR("Some text", blocks)
	assert.Equal(t, "The Great Gatsby", title)
}

func TestBookRecognitionProvider_ExtractTitleFromOCR_LowConfidence(t *testing.T) {
	p := newTestBookProvider(t)

	blocks := []OCRTextBlock{
		{Text: "Blurry Title", Confidence: 0.5, BoundingBox: OCRBoundingBox{Width: 300, Height: 100}},
	}

	// Low confidence blocks should be skipped, falls back to text parsing
	title := p.extractTitleFromOCR("This is a proper title line\nISBN 12345", blocks)
	assert.Equal(t, "This is a proper title line", title)
}

func TestBookRecognitionProvider_ExtractTitleFromOCR_FallbackToText(t *testing.T) {
	p := newTestBookProvider(t)

	title := p.extractTitleFromOCR("Short\nThis Is A Long Enough Title\nISBN 978-3-16-148410-0", nil)
	assert.Equal(t, "This Is A Long Enough Title", title)
}

func TestBookRecognitionProvider_ExtractTitleFromOCR_NoMatch(t *testing.T) {
	p := newTestBookProvider(t)

	title := p.extractTitleFromOCR("ab\ncd", nil)
	assert.Equal(t, "", title)
}

// =============================================================================
// extractAuthorFromOCR Tests
// =============================================================================

func TestBookRecognitionProvider_ExtractAuthorFromOCR_ByPattern(t *testing.T) {
	p := newTestBookProvider(t)

	author := p.extractAuthorFromOCR("The Great Gatsby by Francis Scott Fitzgerald", nil)
	assert.Contains(t, author, "Scott Fitzgerald")
}

func TestBookRecognitionProvider_ExtractAuthorFromOCR_FromBlocks(t *testing.T) {
	p := newTestBookProvider(t)

	blocks := []OCRTextBlock{
		{Text: "the complete guide to modern cooking techniques", Confidence: 0.9, BoundingBox: OCRBoundingBox{Width: 200, Height: 50}},
		{Text: "John Smith", Confidence: 0.85, BoundingBox: OCRBoundingBox{Width: 100, Height: 30}},
	}

	author := p.extractAuthorFromOCR("Some text without by pattern", blocks)
	assert.Equal(t, "John Smith", author)
}

func TestBookRecognitionProvider_ExtractAuthorFromOCR_NoMatch(t *testing.T) {
	p := newTestBookProvider(t)

	author := p.extractAuthorFromOCR("no author info", nil)
	assert.Equal(t, "", author)
}

// =============================================================================
// extractISBNFromText Tests
// =============================================================================

func TestBookRecognitionProvider_ExtractISBNFromText_ISBN13(t *testing.T) {
	p := newTestBookProvider(t)

	isbn := p.extractISBNFromText("ISBN 978-3-16-148410-0 is the identifier")
	assert.NotEmpty(t, isbn)
	assert.Contains(t, isbn, "978316148410")
}

func TestBookRecognitionProvider_ExtractISBNFromText_NoISBN(t *testing.T) {
	p := newTestBookProvider(t)

	isbn := p.extractISBNFromText("This text has no ISBN")
	assert.Equal(t, "", isbn)
}

// =============================================================================
// extractMetadataFromContent Tests
// =============================================================================

func TestBookRecognitionProvider_ExtractMetadataFromContent(t *testing.T) {
	p := newTestBookProvider(t)

	text := "Short line\nThis Is The Book Title That Is Long Enough\nChapter 1: The Beginning\nSome content here.\nChapter 2: The Middle\nMore content."

	metadata := p.extractMetadataFromContent(text)
	assert.Equal(t, "This Is The Book Title That Is Long Enough", metadata.Title)
	assert.Len(t, metadata.ChapterTitles, 2)
	assert.Equal(t, "The Beginning", metadata.ChapterTitles[0])
	assert.Equal(t, "The Middle", metadata.ChapterTitles[1])
}

func TestBookRecognitionProvider_ExtractMetadataFromContent_NoChapters(t *testing.T) {
	p := newTestBookProvider(t)

	text := "Some paragraph that is long enough to qualify as a title\nAnother paragraph."

	metadata := p.extractMetadataFromContent(text)
	assert.NotEmpty(t, metadata.Title)
	assert.Len(t, metadata.ChapterTitles, 0)
}

// =============================================================================
// extractTopics Tests
// =============================================================================

func TestBookRecognitionProvider_ExtractTopics(t *testing.T) {
	p := newTestBookProvider(t)

	text := "Machine Learning is advancing. Machine Learning uses Neural Networks. Neural Networks are powerful."

	topics := p.extractTopics(text)
	// "Machine Learning" and "Neural Networks" should appear multiple times
	assert.NotEmpty(t, topics)
}

func TestBookRecognitionProvider_ExtractTopics_NoRepeats(t *testing.T) {
	p := newTestBookProvider(t)

	text := "Alpha Beta Gamma Delta Epsilon"
	topics := p.extractTopics(text)
	// No topic appears more than once
	assert.Empty(t, topics)
}

// =============================================================================
// extractKeywords Tests
// =============================================================================

func TestBookRecognitionProvider_ExtractKeywords(t *testing.T) {
	p := newTestBookProvider(t)

	text := "programming language concepts programming design patterns programming best practices design architecture"

	keywords := p.extractKeywords(text)
	assert.NotEmpty(t, keywords)
}

func TestBookRecognitionProvider_ExtractKeywords_FilterStopWords(t *testing.T) {
	p := newTestBookProvider(t)

	text := "the the the and and and is is is"
	keywords := p.extractKeywords(text)
	assert.Empty(t, keywords, "stop words should be filtered out")
}

// =============================================================================
// calculateReadabilityScore Tests
// =============================================================================

func TestBookRecognitionProvider_CalculateReadabilityScore(t *testing.T) {
	p := newTestBookProvider(t)

	// Simple text should be more readable (higher score)
	simpleText := "The cat sat on the mat. The dog ran fast. It was a good day."
	score := p.calculateReadabilityScore(simpleText)
	assert.Greater(t, score, 0.0)
	assert.LessOrEqual(t, score, 1.0)
}

func TestBookRecognitionProvider_CalculateReadabilityScore_EmptyText(t *testing.T) {
	p := newTestBookProvider(t)

	score := p.calculateReadabilityScore("")
	assert.Equal(t, 0.0, score)
}

func TestBookRecognitionProvider_CalculateReadabilityScore_NoSentences(t *testing.T) {
	p := newTestBookProvider(t)

	score := p.calculateReadabilityScore("words without any sentence ending punctuation")
	assert.Equal(t, 0.0, score)
}

// =============================================================================
// determineBookType Tests
// =============================================================================

func TestBookRecognitionProvider_DetermineBookType_Magazine(t *testing.T) {
	p := newTestBookProvider(t)

	volumeInfo := GoogleBookVolumeInfo{PrintType: "MAGAZINE"}
	result := p.determineBookType(volumeInfo)
	assert.Equal(t, MediaTypeMagazine, result)
}

func TestBookRecognitionProvider_DetermineBookType_Comic(t *testing.T) {
	p := newTestBookProvider(t)

	volumeInfo := GoogleBookVolumeInfo{
		PrintType:  "BOOK",
		Categories: []string{"Comics & Graphic Novels"},
	}
	result := p.determineBookType(volumeInfo)
	assert.Equal(t, MediaTypeComicBook, result)
}

func TestBookRecognitionProvider_DetermineBookType_Manual(t *testing.T) {
	p := newTestBookProvider(t)

	volumeInfo := GoogleBookVolumeInfo{
		PrintType:  "BOOK",
		Categories: []string{"Reference & Manual"},
	}
	result := p.determineBookType(volumeInfo)
	assert.Equal(t, MediaTypeManual, result)
}

func TestBookRecognitionProvider_DetermineBookType_Default(t *testing.T) {
	p := newTestBookProvider(t)

	volumeInfo := GoogleBookVolumeInfo{
		PrintType:  "BOOK",
		Categories: []string{"Fiction"},
	}
	result := p.determineBookType(volumeInfo)
	assert.Equal(t, MediaTypeBook, result)
}

// =============================================================================
// mapCrossrefType Tests
// =============================================================================

func TestBookRecognitionProvider_MapCrossrefType(t *testing.T) {
	p := newTestBookProvider(t)

	tests := []struct {
		input    string
		expected MediaType
	}{
		{"journal-article", MediaTypeJournal},
		{"book-chapter", MediaTypeBook},
		{"book", MediaTypeBook},
		{"proceedings-article", MediaTypeJournal},
		{"reference-entry", MediaTypeManual},
		{"unknown-type", MediaTypeBook},
		{"JOURNAL-ARTICLE", MediaTypeJournal},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := p.mapCrossrefType(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// =============================================================================
// getFileExtension Tests
// =============================================================================

func TestBookRecognitionProvider_GetFileExtension(t *testing.T) {
	p := newTestBookProvider(t)

	tests := []struct {
		filename string
		expected string
	}{
		{"book.pdf", "pdf"},
		{"document.epub", "epub"},
		{"archive.tar.gz", "gz"},
		{"noextension", ""},
		{"file.with.many.dots.txt", "txt"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			assert.Equal(t, tt.expected, p.getFileExtension(tt.filename))
		})
	}
}

// =============================================================================
// parseDate Tests
// =============================================================================

func TestBookRecognitionProvider_ParseDate(t *testing.T) {
	p := newTestBookProvider(t)

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"ISO date", "2024-01-15", false},
		{"Year-month", "2024-01", false},
		{"Year only", "2024", false},
		{"Long format", "January 15, 2024", false},
		{"Short format", "Jan 15, 2024", false},
		{"ISO datetime", "2024-01-15T10:30:00Z", false},
		{"Invalid", "not-a-date", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			date, err := p.parseDate(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.False(t, date.IsZero())
			}
		})
	}
}

// =============================================================================
// calculateGoogleBooksConfidence Tests
// =============================================================================

func TestBookRecognitionProvider_CalculateGoogleBooksConfidence(t *testing.T) {
	p := newTestBookProvider(t)

	tests := []struct {
		name        string
		rating      float64
		ratingCount int
		expected    float64
	}{
		{"High rating high count", 4.5, 200, 0.8},
		{"Medium rating medium count", 3.7, 60, 0.7},
		{"Low rating low count", 2.0, 5, 0.5},
		{"Zero rating", 0.0, 0, 0.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := p.calculateGoogleBooksConfidence(tt.rating, tt.ratingCount)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// =============================================================================
// calculateOpenLibraryConfidence Tests
// =============================================================================

func TestBookRecognitionProvider_CalculateOpenLibraryConfidence(t *testing.T) {
	p := newTestBookProvider(t)

	tests := []struct {
		name         string
		editionCount int
		ebookCount   int
		expected     float64
	}{
		{"Many editions with ebook", 10, 2, 0.8},
		{"Many editions no ebook", 10, 0, 0.7},
		{"Few editions with ebook", 2, 1, 0.6},
		{"Few editions no ebook", 2, 0, 0.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := p.calculateOpenLibraryConfidence(tt.editionCount, tt.ebookCount)
			assert.InDelta(t, tt.expected, result, 0.001)
		})
	}
}

// =============================================================================
// generateID Tests
// =============================================================================

func TestBookRecognitionProvider_GenerateID(t *testing.T) {
	p := newTestBookProvider(t)

	id1 := p.generateID("test-input-1")
	id2 := p.generateID("test-input-2")
	id3 := p.generateID("test-input-1")

	assert.Len(t, id1, 12)
	assert.Len(t, id2, 12)
	assert.NotEqual(t, id1, id2, "different inputs should produce different IDs")
	assert.Equal(t, id1, id3, "same input should produce same ID")
}

// =============================================================================
// basicBookRecognition Tests
// =============================================================================

func TestBookRecognitionProvider_BasicBookRecognition(t *testing.T) {
	p := newTestBookProvider(t)

	req := &MediaRecognitionRequest{
		FilePath:  "/books/test.pdf",
		MediaType: MediaTypeBook,
	}

	result := p.basicBookRecognition(req, "Test Book", "Test Author", "978-3-16-148410-0", MediaTypeBook)

	require.NotNil(t, result)
	assert.Equal(t, "Test Book", result.Title)
	assert.Equal(t, "Test Author", result.Author)
	assert.Equal(t, "978-3-16-148410-0", result.ISBN)
	assert.Equal(t, MediaTypeBook, result.MediaType)
	assert.Equal(t, 0.3, result.Confidence)
	assert.Equal(t, "filename_parsing", result.RecognitionMethod)
	assert.Equal(t, "basic", result.APIProvider)
	assert.NotNil(t, result.ExternalIDs)
}

// =============================================================================
// convertGoogleBook Tests
// =============================================================================

func TestBookRecognitionProvider_ConvertGoogleBook(t *testing.T) {
	p := newTestBookProvider(t)

	book := GoogleBookItem{
		ID: "abc123",
		VolumeInfo: GoogleBookVolumeInfo{
			Title:         "Test Book",
			Subtitle:      "A Subtitle",
			Authors:       []string{"Author One", "Author Two"},
			Publisher:     "Test Publisher",
			PublishedDate: "2024-01-15",
			Description:   "A test book description",
			PageCount:     350,
			PrintType:     "BOOK",
			Categories:    []string{"Fiction", "Science"},
			AverageRating: 4.2,
			RatingsCount:  150,
			Language:      "en",
			ImageLinks: GoogleBookImageLinks{
				Thumbnail: "http://example.com/thumb.jpg",
				Large:     "http://example.com/large.jpg",
			},
			IndustryIdentifiers: []GoogleBookIdentifier{
				{Type: "ISBN_13", Identifier: "9781234567890"},
				{Type: "ISBN_10", Identifier: "1234567890"},
			},
		},
	}

	result := p.convertGoogleBook(book)

	require.NotNil(t, result)
	// Title includes subtitle: "Test Book: A Subtitle"
	assert.Contains(t, result.Title, "Test Book")
	assert.Equal(t, "Author One", result.Author)
	assert.Equal(t, "Test Publisher", result.Publisher)
	assert.Equal(t, 2024, result.Year)
	assert.Equal(t, "A test book description", result.Description)
	assert.Equal(t, 350, result.PageCount)
	assert.Contains(t, result.Genres, "Fiction")
	assert.Contains(t, result.Genres, "Science")
	assert.NotEmpty(t, result.CoverArt)
	assert.Equal(t, "en", result.Language)
	assert.Equal(t, "Google Books", result.APIProvider)
}

func TestBookRecognitionProvider_ConvertGoogleBook_Minimal(t *testing.T) {
	p := newTestBookProvider(t)

	book := GoogleBookItem{
		ID: "min123",
		VolumeInfo: GoogleBookVolumeInfo{
			Title:     "Minimal Book",
			PrintType: "BOOK",
		},
	}

	result := p.convertGoogleBook(book)

	require.NotNil(t, result)
	assert.Equal(t, "Minimal Book", result.Title)
	assert.Equal(t, MediaTypeBook, result.MediaType)
}

// =============================================================================
// convertOpenLibraryBook Tests
// =============================================================================

func TestBookRecognitionProvider_ConvertOpenLibraryBook(t *testing.T) {
	p := newTestBookProvider(t)

	doc := OpenLibraryDocument{
		Key:                 "/works/OL123",
		Title:               "Open Library Book",
		AuthorName:          []string{"Jane Doe"},
		FirstPublishYear:    2020,
		PublisherName:       []string{"OL Publisher"},
		Language:            []string{"eng"},
		Subject:             []string{"Computer Science", "Programming"},
		ISBN:                []string{"9789876543210"},
		NumberOfPagesMedian: 400,
		EditionCount:        8,
		EbookCount:          3,
		CoverID:             12345,
	}

	result := p.convertOpenLibraryBook(doc)

	require.NotNil(t, result)
	assert.Equal(t, "Open Library Book", result.Title)
	assert.Equal(t, "Jane Doe", result.Author)
	assert.Equal(t, 2020, result.Year)
	assert.Equal(t, "OL Publisher", result.Publisher)
	assert.Contains(t, result.Genres, "Computer Science")
	assert.Equal(t, "Open Library", result.APIProvider)
	assert.NotEmpty(t, result.CoverArt)
}

func TestBookRecognitionProvider_ConvertOpenLibraryBook_Minimal(t *testing.T) {
	p := newTestBookProvider(t)

	doc := OpenLibraryDocument{
		Key:   "/works/OL456",
		Title: "Minimal OL Book",
	}

	result := p.convertOpenLibraryBook(doc)

	require.NotNil(t, result)
	assert.Equal(t, "Minimal OL Book", result.Title)
}

// =============================================================================
// convertCrossrefWork Tests
// =============================================================================

func TestBookRecognitionProvider_ConvertCrossrefWork(t *testing.T) {
	p := newTestBookProvider(t)

	work := CrossrefWork{
		DOI:   "10.1234/test",
		Type:  "journal-article",
		Title: []string{"A Research Paper"},
		Author: []CrossrefAuthor{
			{Given: "John", Family: "Doe"},
			{Given: "Jane", Family: "Smith"},
		},
		ContainerTitle: []string{"Journal of Testing"},
		Publisher:      "Academic Press",
		Issued: CrossrefPartialDate{
			DateParts: [][]int{{2023, 6, 15}},
		},
		Abstract: "This is the abstract of the paper.",
		Subject:  []string{"Computer Science", "Testing"},
		ISSN:     []string{"1234-5678"},
		Page:     "100-120",
	}

	result := p.convertCrossrefWork(work)

	require.NotNil(t, result)
	assert.Equal(t, "A Research Paper", result.Title)
	assert.Equal(t, "John Doe", result.Author)
	assert.Equal(t, 2023, result.Year)
	assert.Equal(t, "Academic Press", result.Publisher)
	assert.Equal(t, MediaTypeJournal, result.MediaType)
	assert.Equal(t, "Crossref", result.APIProvider)
}

func TestBookRecognitionProvider_ConvertCrossrefWork_Minimal(t *testing.T) {
	p := newTestBookProvider(t)

	work := CrossrefWork{
		DOI:  "10.5678/minimal",
		Type: "book",
	}

	result := p.convertCrossrefWork(work)

	require.NotNil(t, result)
	assert.Equal(t, MediaTypeBook, result.MediaType)
}
