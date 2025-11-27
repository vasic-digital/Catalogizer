package factories

import (
	"fmt"
	"math/rand"
	"time"

	"catalogizer/internal/media/models"
	"github.com/google/uuid"
)

// MediaFactory creates test media instances
type MediaFactory struct {
	random *rand.Rand
}

// NewMediaFactory creates a new MediaFactory
func NewMediaFactory() *MediaFactory {
	return &MediaFactory{
		random: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// CreateMedia creates a single media item with default values
func (f *MediaFactory) CreateMedia() *models.Media {
	id := uuid.New().String()
	now := time.Now()
	
	return &models.Media{
		ID:          id,
		Name:        fmt.Sprintf("media_%d.jpg", f.random.Intn(1000)),
		Path:        fmt.Sprintf("/test/path/to/media_%d.jpg", f.random.Intn(1000)),
		Size:        f.random.Int63n(10000000), // 0-10MB
		MimeType:    "image/jpeg",
		CreatedAt:   now.Add(-time.Duration(f.random.Intn(86400)) * time.Second),
		UpdatedAt:   now,
		Checksum:    fmt.Sprintf("checksum_%s", id),
		Thumbnail:   fmt.Sprintf("/thumbnails/%s.jpg", id),
	}
}

// CreateMediaWithOverrides creates a media item with custom overrides
func (f *MediaFactory) CreateMediaWithOverrides(overrides map[string]interface{}) *models.Media {
	media := f.CreateMedia()
	
	for key, value := range overrides {
		switch key {
		case "ID":
			media.ID = value.(string)
		case "Name":
			media.Name = value.(string)
		case "Path":
			media.Path = value.(string)
		case "Size":
			media.Size = value.(int64)
		case "MimeType":
			media.MimeType = value.(string)
		case "CreatedAt":
			media.CreatedAt = value.(time.Time)
		case "UpdatedAt":
			media.UpdatedAt = value.(time.Time)
		case "Checksum":
			media.Checksum = value.(string)
		case "Thumbnail":
			media.Thumbnail = value.(string)
		}
	}
	
	return media
}

// CreateMediaList creates a list of media items
func (f *MediaFactory) CreateMediaList(count int) []*models.Media {
	media := make([]*models.Media, count)
	for i := 0; i < count; i++ {
		media[i] = f.CreateMedia()
	}
	return media
}

// CreateMediaListWithMediaType creates a list of media items with specific type
func (f *MediaFactory) CreateMediaListWithMediaType(count int, mediaType string) []*models.Media {
	media := make([]*models.Media, count)
	for i := 0; i < count; i++ {
		item := f.CreateMedia()
		item.MimeType = mediaType
		media[i] = item
	}
	return media
}

// UserFactory creates test user instances
type UserFactory struct {
	random *rand.Rand
}

// NewUserFactory creates a new UserFactory
func NewUserFactory() *UserFactory {
	return &UserFactory{
		random: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// CreateUser creates a single user with default values
func (f *UserFactory) CreateUser() *models.User {
	id := uuid.New().String()
	
	return &models.User{
		ID:        id,
		Username:  fmt.Sprintf("user_%d", f.random.Intn(1000)),
		Email:     fmt.Sprintf("user_%d@example.com", f.random.Intn(1000)),
		Password:  fmt.Sprintf("password_%d", f.random.Intn(1000)),
		Role:      "user",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// CreateUserWithRole creates a user with a specific role
func (f *UserFactory) CreateUserWithRole(role string) *models.User {
	user := f.CreateUser()
	user.Role = role
	return user
}

// CreateAdminUser creates an admin user
func (f *UserFactory) CreateAdminUser() *models.User {
	return f.CreateUserWithRole("admin")
}

// CollectionFactory creates test collection instances
type CollectionFactory struct {
	random *rand.Rand
}

// NewCollectionFactory creates a new CollectionFactory
func NewCollectionFactory() *CollectionFactory {
	return &CollectionFactory{
		random: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// CreateCollection creates a single collection with default values
func (f *CollectionFactory) CreateCollection() *models.Collection {
	id := uuid.New().String()
	now := time.Now()
	
	return &models.Collection{
		ID:          id,
		Name:        fmt.Sprintf("Collection %d", f.random.Intn(1000)),
		Description: fmt.Sprintf("Test collection description %d", f.random.Intn(1000)),
		CreatedAt:   now,
		UpdatedAt:   now,
		UserID:      uuid.New().String(),
	}
}

// StorageSourceFactory creates test storage source instances
type StorageSourceFactory struct {
	random *rand.Rand
}

// NewStorageSourceFactory creates a new StorageSourceFactory
func NewStorageSourceFactory() *StorageSourceFactory {
	return &StorageSourceFactory{
		random: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// CreateStorageSource creates a single storage source
func (f *StorageSourceFactory) CreateStorageSource() *models.StorageSource {
	protocols := []string{"smb", "ftp", "nfs", "webdav", "local"}
	protocol := protocols[f.random.Intn(len(protocols))]
	
	return &models.StorageSource{
		ID:       uuid.New().String(),
		Name:     fmt.Sprintf("Source %d", f.random.Intn(1000)),
		Protocol: protocol,
		Host:     fmt.Sprintf("host%d.example.com", f.random.Intn(100)),
		Path:     fmt.Sprintf("/share%d", f.random.Intn(10)),
		Username: fmt.Sprintf("user%d", f.random.Intn(100)),
		Password: fmt.Sprintf("pass%d", f.random.Intn(1000)),
		Enabled:  true,
	}
}