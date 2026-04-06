package challenges

import (
	"encoding/json"
	"fmt"
	"os"
)

// ChallengeBank represents the JSON challenge bank structure.
type ChallengeBank struct {
	Version     string               `json:"version"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Generated   string               `json:"generated"`
	Categories  map[string]string    `json:"categories"`
	Challenges  []ChallengeBankEntry `json:"challenges"`
}

// ChallengeBankEntry represents a single challenge definition
// from the bank.
type ChallengeBankEntry struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
}

// LoadChallengeBank reads and parses the challenge bank JSON
// file at the given path.
func LoadChallengeBank(path string) (*ChallengeBank, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read challenge bank: %w", err)
	}

	var bank ChallengeBank
	if err := json.Unmarshal(data, &bank); err != nil {
		return nil, fmt.Errorf("failed to parse challenge bank: %w", err)
	}

	return &bank, nil
}

// GetChallengesByCategory returns all challenge entries for
// a given category.
func (b *ChallengeBank) GetChallengesByCategory(
	category string,
) []ChallengeBankEntry {
	var result []ChallengeBankEntry
	for _, c := range b.Challenges {
		if c.Category == category {
			result = append(result, c)
		}
	}
	return result
}

// GetChallengesBySeverity returns all challenges at a given
// severity level.
func (b *ChallengeBank) GetChallengesBySeverity(
	severity string,
) []ChallengeBankEntry {
	var result []ChallengeBankEntry
	for _, c := range b.Challenges {
		if c.Severity == severity {
			result = append(result, c)
		}
	}
	return result
}

// GetChallengeByID returns the challenge entry with the given
// ID, or nil if not found.
func (b *ChallengeBank) GetChallengeByID(
	id string,
) *ChallengeBankEntry {
	for i := range b.Challenges {
		if b.Challenges[i].ID == id {
			return &b.Challenges[i]
		}
	}
	return nil
}
