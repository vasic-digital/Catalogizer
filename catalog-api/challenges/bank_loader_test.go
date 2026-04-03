package challenges

import (
	"os"
	"path/filepath"
	"testing"
)

// bankPath returns the absolute path to the challenge bank
// JSON file relative to the repository root.
func bankPath(t *testing.T) string {
	t.Helper()
	// Tests run from catalog-api/challenges/, bank is at
	// ../../challenges/data/challenges_bank.json
	path := filepath.Join("..", "..", "challenges", "data",
		"challenges_bank.json")
	abs, err := filepath.Abs(path)
	if err != nil {
		t.Fatalf("failed to resolve bank path: %v", err)
	}
	return abs
}

func TestLoadChallengeBank(t *testing.T) {
	path := bankPath(t)
	bank, err := LoadChallengeBank(path)
	if err != nil {
		t.Fatalf("LoadChallengeBank(%q): %v", path, err)
	}

	if bank.Version == "" {
		t.Error("expected non-empty Version")
	}
	if bank.Name == "" {
		t.Error("expected non-empty Name")
	}
	if bank.Description == "" {
		t.Error("expected non-empty Description")
	}
	if bank.Generated == "" {
		t.Error("expected non-empty Generated")
	}
	if len(bank.Categories) == 0 {
		t.Error("expected non-empty Categories map")
	}
	if len(bank.Challenges) == 0 {
		t.Fatal("expected non-empty Challenges slice")
	}

	t.Logf("loaded %d challenges, version %s",
		len(bank.Challenges), bank.Version)
}

func TestLoadChallengeBank_MissingFile(t *testing.T) {
	_, err := LoadChallengeBank("/nonexistent/path.json")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestLoadChallengeBank_InvalidJSON(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "bad.json")
	if err := os.WriteFile(tmp, []byte("{invalid"), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	_, err := LoadChallengeBank(tmp)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestGetChallengesByCategory(t *testing.T) {
	path := bankPath(t)
	bank, err := LoadChallengeBank(path)
	if err != nil {
		t.Fatalf("LoadChallengeBank: %v", err)
	}

	// Pick a category that we know exists from the JSON
	storage := bank.GetChallengesByCategory("storage")
	if len(storage) == 0 {
		t.Error("expected at least one storage challenge")
	}
	for _, c := range storage {
		if c.Category != "storage" {
			t.Errorf("expected category=storage, got %q", c.Category)
		}
	}

	// Nonexistent category returns empty
	none := bank.GetChallengesByCategory("nonexistent-category")
	if len(none) != 0 {
		t.Errorf("expected 0 challenges for unknown category, got %d",
			len(none))
	}
}

func TestGetChallengesBySeverity(t *testing.T) {
	path := bankPath(t)
	bank, err := LoadChallengeBank(path)
	if err != nil {
		t.Fatalf("LoadChallengeBank: %v", err)
	}

	critical := bank.GetChallengesBySeverity("critical")
	if len(critical) == 0 {
		t.Error("expected at least one critical challenge")
	}
	for _, c := range critical {
		if c.Severity != "critical" {
			t.Errorf("expected severity=critical, got %q",
				c.Severity)
		}
	}

	// Nonexistent severity returns empty
	none := bank.GetChallengesBySeverity("nonexistent-severity")
	if len(none) != 0 {
		t.Errorf("expected 0 for unknown severity, got %d",
			len(none))
	}
}

func TestGetChallengeByID(t *testing.T) {
	path := bankPath(t)
	bank, err := LoadChallengeBank(path)
	if err != nil {
		t.Fatalf("LoadChallengeBank: %v", err)
	}

	ch := bank.GetChallengeByID("CH-001")
	if ch == nil {
		t.Fatal("expected CH-001 to exist")
	}
	if ch.Name == "" {
		t.Error("expected non-empty name for CH-001")
	}

	missing := bank.GetChallengeByID("NONEXISTENT-999")
	if missing != nil {
		t.Error("expected nil for nonexistent ID")
	}
}

func TestAllEntriesHaveRequiredFields(t *testing.T) {
	path := bankPath(t)
	bank, err := LoadChallengeBank(path)
	if err != nil {
		t.Fatalf("LoadChallengeBank: %v", err)
	}

	for i, c := range bank.Challenges {
		if c.ID == "" {
			t.Errorf("challenge[%d]: empty ID", i)
		}
		if c.Name == "" {
			t.Errorf("challenge[%d] (%s): empty Name", i, c.ID)
		}
		if c.Category == "" {
			t.Errorf("challenge[%d] (%s): empty Category",
				i, c.ID)
		}
	}
}
