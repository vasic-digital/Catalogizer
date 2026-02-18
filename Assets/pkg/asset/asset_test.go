package asset

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewID(t *testing.T) {
	id1 := NewID()
	id2 := NewID()

	assert.NotEmpty(t, id1.String())
	assert.NotEmpty(t, id2.String())
	assert.NotEqual(t, id1, id2)
}

func TestNew(t *testing.T) {
	a := New(TypeImage, "file", "42")

	assert.NotEmpty(t, a.ID)
	assert.Equal(t, TypeImage, a.Type)
	assert.Equal(t, StatusPending, a.Status)
	assert.Equal(t, "file", a.EntityType)
	assert.Equal(t, "42", a.EntityID)
	assert.NotNil(t, a.Metadata)
	assert.False(t, a.CreatedAt.IsZero())
}

func TestAsset_MarkResolving(t *testing.T) {
	a := New(TypeImage, "file", "1")
	a.MarkResolving()

	assert.Equal(t, StatusResolving, a.Status)
}

func TestAsset_MarkReady(t *testing.T) {
	a := New(TypeImage, "file", "1")
	a.MarkReady("image/jpeg", 1024)

	assert.Equal(t, StatusReady, a.Status)
	assert.Equal(t, "image/jpeg", a.ContentType)
	assert.Equal(t, int64(1024), a.Size)
	require.NotNil(t, a.ResolvedAt)
}

func TestAsset_MarkFailed(t *testing.T) {
	a := New(TypeImage, "file", "1")
	a.MarkFailed()

	assert.Equal(t, StatusFailed, a.Status)
}

func TestAsset_IsTerminal(t *testing.T) {
	tests := []struct {
		status   Status
		terminal bool
	}{
		{StatusPending, false},
		{StatusResolving, false},
		{StatusReady, true},
		{StatusFailed, true},
		{StatusExpired, true},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			a := &Asset{Status: tt.status}
			assert.Equal(t, tt.terminal, a.IsTerminal())
		})
	}
}

func TestAsset_IsExpired(t *testing.T) {
	a := New(TypeImage, "file", "1")
	assert.False(t, a.IsExpired(), "no expiry set")

	past := time.Now().Add(-time.Hour)
	a.ExpiresAt = &past
	assert.True(t, a.IsExpired(), "expired in the past")

	future := time.Now().Add(time.Hour)
	a.ExpiresAt = &future
	assert.False(t, a.IsExpired(), "expires in the future")
}

func TestTypeConstants(t *testing.T) {
	assert.Equal(t, Type("image"), TypeImage)
	assert.Equal(t, Type("video_thumbnail"), TypeVideoThumbnail)
	assert.Equal(t, Type("audio_cover"), TypeAudioCover)
	assert.Equal(t, Type("document_thumbnail"), TypeDocumentThumbnail)
}

func TestStatusConstants(t *testing.T) {
	assert.Equal(t, Status("pending"), StatusPending)
	assert.Equal(t, Status("resolving"), StatusResolving)
	assert.Equal(t, Status("ready"), StatusReady)
	assert.Equal(t, Status("failed"), StatusFailed)
	assert.Equal(t, Status("expired"), StatusExpired)
}
