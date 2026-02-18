package defaults

import (
	"bytes"
	"embed"
	"fmt"
	"io"
	"sync"

	"digital.vasic.assets/pkg/asset"
	"digital.vasic.assets/pkg/store"
)

//go:embed images/*.png
var embeddedImages embed.FS

// Provider returns default/fallback content for asset types.
type Provider interface {
	GetDefault(assetType asset.Type) (io.ReadCloser, *store.Info, error)
	Register(assetType asset.Type, content []byte, contentType string)
}

// EmbeddedProvider serves default images from embedded files.
type EmbeddedProvider struct {
	mu       sync.RWMutex
	custom   map[asset.Type]customDefault
	fallback map[asset.Type]string
}

type customDefault struct {
	content     []byte
	contentType string
}

// NewEmbeddedProvider creates a provider backed by embedded placeholder PNGs.
func NewEmbeddedProvider() *EmbeddedProvider {
	return &EmbeddedProvider{
		custom: make(map[asset.Type]customDefault),
		fallback: map[asset.Type]string{
			asset.TypeImage:             "images/image.png",
			asset.TypeVideoThumbnail:    "images/video.png",
			asset.TypeAudioCover:        "images/audio.png",
			asset.TypeDocumentThumbnail: "images/document.png",
		},
	}
}

// GetDefault returns placeholder content for the given asset type.
func (p *EmbeddedProvider) GetDefault(assetType asset.Type) (io.ReadCloser, *store.Info, error) {
	p.mu.RLock()
	if custom, ok := p.custom[assetType]; ok {
		p.mu.RUnlock()
		return io.NopCloser(bytes.NewReader(custom.content)), &store.Info{
			ContentType: custom.contentType,
			Size:        int64(len(custom.content)),
		}, nil
	}
	p.mu.RUnlock()

	path, ok := p.fallback[assetType]
	if !ok {
		path = "images/image.png"
	}

	data, err := embeddedImages.ReadFile(path)
	if err != nil {
		return nil, nil, fmt.Errorf("read embedded default for %s: %w", assetType, err)
	}

	return io.NopCloser(bytes.NewReader(data)), &store.Info{
		ContentType: "image/png",
		Size:        int64(len(data)),
	}, nil
}

// Register adds a custom default for the given asset type.
func (p *EmbeddedProvider) Register(assetType asset.Type, content []byte, contentType string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.custom[assetType] = customDefault{
		content:     content,
		contentType: contentType,
	}
}
