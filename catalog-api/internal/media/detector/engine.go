package detector

import (
	"catalogizer/internal/media/models"
	"encoding/json"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"go.uber.org/zap"
)

// DetectionEngine handles automatic content type detection
type DetectionEngine struct {
	logger    *zap.Logger
	rules     []models.DetectionRule
	mediaTypes map[int64]*models.MediaType
}

// NewDetectionEngine creates a new detection engine
func NewDetectionEngine(logger *zap.Logger) *DetectionEngine {
	return &DetectionEngine{
		logger:     logger,
		rules:      []models.DetectionRule{},
		mediaTypes: make(map[int64]*models.MediaType),
	}
}

// LoadRules loads detection rules from database or configuration
func (e *DetectionEngine) LoadRules(rules []models.DetectionRule, mediaTypes []models.MediaType) {
	e.rules = rules

	// Sort rules by priority (higher first)
	sort.Slice(e.rules, func(i, j int) bool {
		return e.rules[i].Priority > e.rules[j].Priority
	})

	// Index media types
	for _, mt := range mediaTypes {
		e.mediaTypes[mt.ID] = &mt
	}

	e.logger.Info("Detection rules loaded",
		zap.Int("rules_count", len(e.rules)),
		zap.Int("media_types", len(e.mediaTypes)))
}

// DetectionResult represents the result of content detection
type DetectionResult struct {
	MediaTypeID     int64                    `json:"media_type_id"`
	MediaType       *models.MediaType        `json:"media_type"`
	Confidence      float64                  `json:"confidence"`
	Method          string                   `json:"method"`
	MatchedPatterns []string                 `json:"matched_patterns"`
	AnalysisData    *models.AnalysisData     `json:"analysis_data"`
	SuggestedTitle  string                   `json:"suggested_title"`
	SuggestedYear   *int                     `json:"suggested_year"`
	QualityHints    []string                 `json:"quality_hints"`
}

// AnalyzeDirectory analyzes a directory to determine its content type
func (e *DetectionEngine) AnalyzeDirectory(directoryPath string, files []FileInfo) (*DetectionResult, error) {
	results := make([]*DetectionResult, 0)

	// Run all enabled rules
	for _, rule := range e.rules {
		if !rule.Enabled {
			continue
		}

		mediaType := e.mediaTypes[rule.MediaTypeID]
		if mediaType == nil {
			continue
		}

		var confidence float64
		var matchedPatterns []string
		var method string

		switch rule.RuleType {
		case "filename_pattern":
			confidence, matchedPatterns = e.analyzeFilenamePatterns(rule.Pattern, files)
			method = "filename_pattern"

		case "directory_structure":
			confidence, matchedPatterns = e.analyzeDirectoryStructure(rule.Pattern, directoryPath, files)
			method = "directory_structure"

		case "file_analysis":
			confidence, matchedPatterns = e.analyzeFileContent(rule.Pattern, files)
			method = "file_analysis"

		case "hybrid":
			confidence, matchedPatterns = e.analyzeHybrid(rule.Pattern, directoryPath, files)
			method = "hybrid"
		}

		if confidence > 0 {
			// Apply rule weight
			confidence *= rule.ConfidenceWeight

			// Extract additional metadata
			title, year := e.extractTitleAndYear(directoryPath, files, mediaType.Name)
			qualityHints := e.extractQualityHints(directoryPath, files)

			result := &DetectionResult{
				MediaTypeID:     rule.MediaTypeID,
				MediaType:       mediaType,
				Confidence:      confidence,
				Method:          method,
				MatchedPatterns: matchedPatterns,
				SuggestedTitle:  title,
				SuggestedYear:   year,
				QualityHints:    qualityHints,
				AnalysisData:    e.buildAnalysisData(directoryPath, files, matchedPatterns, confidence),
			}

			results = append(results, result)
		}
	}

	// Return the highest confidence result
	if len(results) == 0 {
		return nil, nil
	}

	// Sort by confidence
	sort.Slice(results, func(i, j int) bool {
		return results[i].Confidence > results[j].Confidence
	})

	return results[0], nil
}

// FileInfo represents basic file information for analysis
type FileInfo struct {
	Name      string
	Path      string
	Size      int64
	Extension string
	IsDir     bool
}

// analyzeFilenamePatterns checks files against filename patterns
func (e *DetectionEngine) analyzeFilenamePatterns(pattern string, files []FileInfo) (float64, []string) {
	var patterns []string
	if err := json.Unmarshal([]byte(pattern), &patterns); err != nil {
		// Single pattern
		patterns = []string{pattern}
	}

	totalFiles := len(files)
	if totalFiles == 0 {
		return 0, nil
	}

	matchedFiles := 0
	matchedPatterns := make([]string, 0)

	for _, p := range patterns {
		regex, err := e.globToRegex(p)
		if err != nil {
			continue
		}

		for _, file := range files {
			if regex.MatchString(strings.ToLower(file.Name)) {
				matchedFiles++
				matchedPatterns = append(matchedPatterns, p)
				break // Don't count same pattern multiple times
			}
		}
	}

	confidence := float64(matchedFiles) / float64(len(patterns))
	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence, matchedPatterns
}

// analyzeDirectoryStructure analyzes directory structure patterns
func (e *DetectionEngine) analyzeDirectoryStructure(pattern string, dirPath string, files []FileInfo) (float64, []string) {
	var structureRules map[string]interface{}
	if err := json.Unmarshal([]byte(pattern), &structureRules); err != nil {
		return 0, nil
	}

	confidence := 0.0
	matchedPatterns := make([]string, 0)

	// Check for required directories
	if requiredDirs, ok := structureRules["required_dirs"].([]interface{}); ok {
		for _, reqDir := range requiredDirs {
			dirName := reqDir.(string)
			found := false
			for _, file := range files {
				if file.IsDir && strings.Contains(strings.ToLower(file.Name), strings.ToLower(dirName)) {
					found = true
					break
				}
			}
			if found {
				confidence += 0.3
				matchedPatterns = append(matchedPatterns, "dir:"+dirName)
			}
		}
	}

	// Check for file type distribution
	if fileTypes, ok := structureRules["file_types"].(map[string]interface{}); ok {
		for ext, minCount := range fileTypes {
			count := 0
			for _, file := range files {
				if strings.EqualFold(file.Extension, ext) {
					count++
				}
			}
			if count >= int(minCount.(float64)) {
				confidence += 0.2
				matchedPatterns = append(matchedPatterns, "filetype:"+ext)
			}
		}
	}

	// Normalize confidence
	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence, matchedPatterns
}

// analyzeFileContent analyzes actual file content (basic implementation)
func (e *DetectionEngine) analyzeFileContent(pattern string, files []FileInfo) (float64, []string) {
	// This would implement more sophisticated content analysis
	// For now, it's a placeholder that analyzes file extensions and sizes

	var contentRules map[string]interface{}
	if err := json.Unmarshal([]byte(pattern), &contentRules); err != nil {
		return 0, nil
	}

	confidence := 0.0
	matchedPatterns := make([]string, 0)

	// Analyze file size patterns
	if sizeRules, ok := contentRules["size_patterns"].(map[string]interface{}); ok {
		for sizeType, conditions := range sizeRules {
			condMap := conditions.(map[string]interface{})
			minSize := int64(condMap["min_size"].(float64))
			maxSize := int64(condMap["max_size"].(float64))

			matchCount := 0
			for _, file := range files {
				if file.Size >= minSize && file.Size <= maxSize {
					matchCount++
				}
			}

			if matchCount > 0 {
				confidence += 0.2
				matchedPatterns = append(matchedPatterns, "size:"+sizeType)
			}
		}
	}

	return confidence, matchedPatterns
}

// analyzeHybrid combines multiple analysis methods
func (e *DetectionEngine) analyzeHybrid(pattern string, dirPath string, files []FileInfo) (float64, []string) {
	var hybridRules map[string]interface{}
	if err := json.Unmarshal([]byte(pattern), &hybridRules); err != nil {
		return 0, nil
	}

	totalConfidence := 0.0
	allPatterns := make([]string, 0)

	// Run filename analysis
	if filenamePattern, ok := hybridRules["filename"].(string); ok {
		conf, patterns := e.analyzeFilenamePatterns(filenamePattern, files)
		totalConfidence += conf * 0.4
		allPatterns = append(allPatterns, patterns...)
	}

	// Run structure analysis
	if structurePattern, ok := hybridRules["structure"].(string); ok {
		conf, patterns := e.analyzeDirectoryStructure(structurePattern, dirPath, files)
		totalConfidence += conf * 0.3
		allPatterns = append(allPatterns, patterns...)
	}

	// Run content analysis
	if contentPattern, ok := hybridRules["content"].(string); ok {
		conf, patterns := e.analyzeFileContent(contentPattern, files)
		totalConfidence += conf * 0.3
		allPatterns = append(allPatterns, patterns...)
	}

	return totalConfidence, allPatterns
}

// extractTitleAndYear extracts title and year from directory/file names
func (e *DetectionEngine) extractTitleAndYear(dirPath string, files []FileInfo, mediaType string) (string, *int) {
	dirName := filepath.Base(dirPath)

	// Extract year using regex
	yearRegex := regexp.MustCompile(`\((\d{4})\)|\[(\d{4})\]|(\d{4})`)
	yearMatches := yearRegex.FindStringSubmatch(dirName)

	var year *int
	if len(yearMatches) > 0 {
		for i := 1; i < len(yearMatches); i++ {
			if yearMatches[i] != "" {
				if y := parseInt(yearMatches[i]); y > 1900 && y <= time.Now().Year()+5 {
					year = &y
				}
				break
			}
		}
	}

	// Clean title
	title := dirName

	// Remove year from title
	if year != nil {
		title = yearRegex.ReplaceAllString(title, "")
	}

	// Remove common release info
	cleanupRegex := regexp.MustCompile(`(?i)\b(bluray|brrip|dvdrip|webrip|hdtv|720p|1080p|4k|x264|x265|h264|h265|aac|dts|ac3|complete|season|series)\b`)
	title = cleanupRegex.ReplaceAllString(title, "")

	// Clean up extra spaces and dots
	title = regexp.MustCompile(`[._\-\s]+`).ReplaceAllString(title, " ")
	title = strings.TrimSpace(title)

	return title, year
}

// extractQualityHints extracts quality indicators from filenames
func (e *DetectionEngine) extractQualityHints(dirPath string, files []FileInfo) []string {
	hints := make([]string, 0)
	text := strings.ToLower(dirPath)

	// Add file names to analysis
	for _, file := range files {
		text += " " + strings.ToLower(file.Name)
	}

	qualityPatterns := map[string][]string{
		"4K":      {"4k", "uhd", "2160p"},
		"1080p":   {"1080p", "fullhd", "fhd"},
		"720p":    {"720p", "hd"},
		"BluRay":  {"bluray", "brrip", "bd"},
		"WEB-DL":  {"webdl", "web-dl", "webrip"},
		"HDR":     {"hdr", "hdr10", "dolby.vision"},
		"Lossless": {"flac", "lossless", "dts-hd"},
		"Remux":   {"remux"},
	}

	for quality, patterns := range qualityPatterns {
		for _, pattern := range patterns {
			if strings.Contains(text, pattern) {
				hints = append(hints, quality)
				break
			}
		}
	}

	return hints
}

// buildAnalysisData creates detailed analysis information
func (e *DetectionEngine) buildAnalysisData(dirPath string, files []FileInfo, patterns []string, confidence float64) *models.AnalysisData {
	fileTypes := make(map[string]int)
	sizeDistribution := make(map[string]int64)

	for _, file := range files {
		if file.Extension != "" {
			fileTypes[file.Extension]++
		}

		// Categorize by size
		var sizeCategory string
		switch {
		case file.Size > 10*1024*1024*1024: // > 10GB
			sizeCategory = "very_large"
		case file.Size > 1024*1024*1024: // > 1GB
			sizeCategory = "large"
		case file.Size > 100*1024*1024: // > 100MB
			sizeCategory = "medium"
		case file.Size > 10*1024*1024: // > 10MB
			sizeCategory = "small"
		default:
			sizeCategory = "tiny"
		}
		sizeDistribution[sizeCategory] += file.Size
	}

	return &models.AnalysisData{
		MatchedPatterns:   patterns,
		FileTypes:         fileTypes,
		SizeDistribution:  sizeDistribution,
		QualityIndicators: e.extractQualityHints(dirPath, files),
		FilenameScore:     confidence * 0.4,
		StructureScore:    confidence * 0.3,
		MetadataScore:     confidence * 0.3,
	}
}

// Helper functions
func (e *DetectionEngine) globToRegex(glob string) (*regexp.Regexp, error) {
	// Convert glob pattern to regex
	pattern := strings.ReplaceAll(glob, "*", ".*")
	pattern = strings.ReplaceAll(pattern, "?", ".")
	pattern = "^" + pattern + "$"
	return regexp.Compile("(?i)" + pattern)
}

func parseInt(s string) int {
	var result int
	for _, char := range s {
		if char >= '0' && char <= '9' {
			result = result*10 + int(char-'0')
		}
	}
	return result
}

// GetSupportedMediaTypes returns all supported media types
func (e *DetectionEngine) GetSupportedMediaTypes() []models.MediaType {
	types := make([]models.MediaType, 0, len(e.mediaTypes))
	for _, mt := range e.mediaTypes {
		types = append(types, *mt)
	}
	return types
}

// ValidateDetection validates a detection result
func (e *DetectionEngine) ValidateDetection(result *DetectionResult) bool {
	if result == nil {
		return false
	}

	// Minimum confidence threshold
	if result.Confidence < 0.1 {
		return false
	}

	// Must have at least one matched pattern
	if len(result.MatchedPatterns) == 0 {
		return false
	}

	return true
}