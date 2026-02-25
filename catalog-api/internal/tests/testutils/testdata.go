package testutils

import (
	"math/rand"
	"time"
)

// TestDataGenerator provides methods to generate test data
type TestDataGenerator struct {
	rng *rand.Rand
}

// NewTestDataGenerator creates a new test data generator
func NewTestDataGenerator(seed int64) *TestDataGenerator {
	return &TestDataGenerator{
		rng: rand.New(rand.NewSource(seed)),
	}
}

// RandomString generates a random string of specified length
func (g *TestDataGenerator) RandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[g.rng.Intn(len(charset))]
	}
	return string(b)
}

// RandomInt generates a random integer between min and max
func (g *TestDataGenerator) RandomInt(min, max int) int {
	return g.rng.Intn(max-min+1) + min
}

// RandomTime generates a random time within the last year
func (g *TestDataGenerator) RandomTime() time.Time {
	now := time.Now()
	yearAgo := now.AddDate(-1, 0, 0)
	diff := now.Sub(yearAgo)
	randomDuration := time.Duration(g.rng.Int63n(int64(diff)))
	return yearAgo.Add(randomDuration)
}

// RandomEmail generates a random email address
func (g *TestDataGenerator) RandomEmail() string {
	return g.RandomString(8) + "@example.com"
}

// RandomFilePath generates a random file path
func (g *TestDataGenerator) RandomFilePath() string {
	extensions := []string{".mp4", ".mkv", ".avi", ".mp3", ".flac", ".jpg", ".png"}
	dirs := []string{"movies", "tv_shows", "music", "photos", "documents"}
	dir := dirs[g.rng.Intn(len(dirs))]
	filename := g.RandomString(10) + extensions[g.rng.Intn(len(extensions))]
	return "/" + dir + "/" + filename
}

// RandomMediaType generates a random media type
func (g *TestDataGenerator) RandomMediaType() string {
	types := []string{"movie", "tv_show", "music", "photo", "document"}
	return types[g.rng.Intn(len(types))]
}

// GenerateUsers generates n test users
func (g *TestDataGenerator) GenerateUsers(n int) []MockUser {
	users := make([]MockUser, n)
	for i := 0; i < n; i++ {
		users[i] = MockUser{
			ID:        int64(i + 1),
			Username:  "user_" + g.RandomString(6),
			Email:     g.RandomEmail(),
			Password:  "hash_" + g.RandomString(32),
			RoleID:    g.RandomInt(1, 3),
			IsActive:  g.rng.Float32() > 0.1, // 90% active
			CreatedAt: g.RandomTime(),
		}
	}
	return users
}

// GenerateMediaItems generates n test media items
func (g *TestDataGenerator) GenerateMediaItems(n int, userID int64) []MockMediaItem {
	items := make([]MockMediaItem, n)
	for i := 0; i < n; i++ {
		items[i] = MockMediaItem{
			ID:          int64(i + 1),
			Title:       "Media " + g.RandomString(8),
			Type:        g.RandomMediaType(),
			Path:        g.RandomFilePath(),
			Duration:    g.RandomInt(60, 7200),
			CreatedAt:   g.RandomTime(),
			UserID:      userID,
			StorageRoot: int64(g.RandomInt(1, 5)),
		}
	}
	return items
}
