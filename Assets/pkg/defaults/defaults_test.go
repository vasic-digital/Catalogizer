package defaults

import (
	"io"
	"testing"

	"digital.vasic.assets/pkg/asset"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmbeddedProvider_GetDefault(t *testing.T) {
	p := NewEmbeddedProvider()

	types := []asset.Type{
		asset.TypeImage,
		asset.TypeVideoThumbnail,
		asset.TypeAudioCover,
		asset.TypeDocumentThumbnail,
	}

	for _, assetType := range types {
		t.Run(string(assetType), func(t *testing.T) {
			reader, info, err := p.GetDefault(assetType)
			require.NoError(t, err)
			defer reader.Close()

			data, err := io.ReadAll(reader)
			require.NoError(t, err)
			assert.True(t, len(data) > 0, "default content should not be empty")
			assert.Equal(t, "image/png", info.ContentType)
			assert.True(t, info.Size > 0)
		})
	}
}

func TestEmbeddedProvider_CustomRegistration(t *testing.T) {
	p := NewEmbeddedProvider()

	customContent := []byte("custom-image-data")
	p.Register(asset.TypeImage, customContent, "image/jpeg")

	reader, info, err := p.GetDefault(asset.TypeImage)
	require.NoError(t, err)
	defer reader.Close()

	data, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, customContent, data)
	assert.Equal(t, "image/jpeg", info.ContentType)
}

func TestEmbeddedProvider_UnknownType_FallsBackToImage(t *testing.T) {
	p := NewEmbeddedProvider()

	reader, info, err := p.GetDefault(asset.Type("unknown_type"))
	require.NoError(t, err)
	defer reader.Close()

	data, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.True(t, len(data) > 0)
	assert.Equal(t, "image/png", info.ContentType)
}
