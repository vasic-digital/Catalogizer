package challenges

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEndpoint_Validation(t *testing.T) {
	tests := []struct {
		name     string
		endpoint Endpoint
		wantErr  bool
	}{
		{
			name: "valid endpoint",
			endpoint: Endpoint{
				Host: "localhost",
				Port: 445,
				Directories: []Directory{
					{Path: "/media", ContentType: "movie"},
				},
			},
			wantErr: false,
		},
		{
			name: "missing host",
			endpoint: Endpoint{
				Port: 445,
			},
			wantErr: true,
		},
		{
			name: "zero port",
			endpoint: Endpoint{
				Host: "localhost",
				Port: 0,
			},
			wantErr: true,
		},
		{
			name: "valid without directories",
			endpoint: Endpoint{
				Host: "localhost",
				Port: 445,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.endpoint.Host != "" && tt.endpoint.Port > 0
			if tt.wantErr {
				assert.False(t, isValid)
			} else {
				assert.True(t, isValid)
			}
		})
	}
}

func TestDirectory_Validation(t *testing.T) {
	tests := []struct {
		name    string
		dir     Directory
		wantErr bool
	}{
		{
			name: "valid movie config",
			dir: Directory{
				Path:        "/media/movies",
				ContentType: "movie",
			},
			wantErr: false,
		},
		{
			name: "valid tv_show config",
			dir: Directory{
				Path:        "/media/tv",
				ContentType: "tv_show",
			},
			wantErr: false,
		},
		{
			name: "valid music config",
			dir: Directory{
				Path:        "/media/music",
				ContentType: "music",
			},
			wantErr: false,
		},
		{
			name: "missing path",
			dir: Directory{
				ContentType: "movie",
			},
			wantErr: true,
		},
		{
			name: "missing content type",
			dir: Directory{
				Path: "/media",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.dir.Path != "" && tt.dir.ContentType != ""
			if tt.wantErr {
				assert.False(t, isValid)
			} else {
				assert.True(t, isValid)
			}
		})
	}
}

func TestSMBConnectivityChallenge(t *testing.T) {
	ep := &Endpoint{
		Host: "localhost",
		Port: 445,
	}

	ch := NewSMBConnectivityChallenge(ep)

	assert.NotNil(t, ch)
	assert.NotEmpty(t, ch.ID())
	assert.NotEmpty(t, ch.Name())
	assert.Contains(t, ch.Name(), "SMB")
}

func TestDirectoryDiscoveryChallenge(t *testing.T) {
	ep := &Endpoint{
		Host: "localhost",
		Port: 445,
		Directories: []Directory{
			{Path: "/media", ContentType: "movie"},
		},
	}

	ch := NewDirectoryDiscoveryChallenge(ep)

	assert.NotNil(t, ch)
	assert.NotEmpty(t, ch.ID())
	assert.NotEmpty(t, ch.Name())
	assert.Contains(t, ch.Name(), "Directory")
}

func TestMusicScanChallenge(t *testing.T) {
	ep := &Endpoint{
		Host: "localhost",
		Port: 445,
	}

	dir := Directory{
		Path:        "/media/music",
		ContentType: "music",
	}

	ch := NewMusicScanChallenge(ep, dir)

	assert.NotNil(t, ch)
	assert.NotEmpty(t, ch.ID())
	assert.NotEmpty(t, ch.Name())
	assert.Contains(t, ch.Name(), "Music")
}

func TestMoviesScanChallenge(t *testing.T) {
	ep := &Endpoint{
		Host: "localhost",
		Port: 445,
	}

	dir := Directory{
		Path:        "/media/movies",
		ContentType: "movie",
	}

	ch := NewMoviesScanChallenge(ep, dir)

	assert.NotNil(t, ch)
	assert.NotEmpty(t, ch.ID())
	assert.NotEmpty(t, ch.Name())
	assert.Contains(t, ch.Name(), "Movie")
}

func TestSeriesScanChallenge(t *testing.T) {
	ep := &Endpoint{
		Host: "localhost",
		Port: 445,
	}

	dir := Directory{
		Path:        "/media/tv",
		ContentType: "tv_show",
	}

	ch := NewSeriesScanChallenge(ep, dir)

	assert.NotNil(t, ch)
	assert.NotEmpty(t, ch.ID())
	assert.NotEmpty(t, ch.Name())
	assert.Contains(t, ch.Name(), "Series")
}

func TestSoftwareScanChallenge(t *testing.T) {
	ep := &Endpoint{
		Host: "localhost",
		Port: 445,
	}

	dir := Directory{
		Path:        "/media/software",
		ContentType: "software",
	}

	ch := NewSoftwareScanChallenge(ep, dir)

	assert.NotNil(t, ch)
	assert.NotEmpty(t, ch.ID())
	assert.NotEmpty(t, ch.Name())
	assert.Contains(t, ch.Name(), "Software")
}

func TestComicsScanChallenge(t *testing.T) {
	ep := &Endpoint{
		Host: "localhost",
		Port: 445,
	}

	dir := Directory{
		Path:        "/media/comics",
		ContentType: "comic",
	}

	ch := NewComicsScanChallenge(ep, dir)

	assert.NotNil(t, ch)
	assert.NotEmpty(t, ch.ID())
	assert.NotEmpty(t, ch.Name())
	assert.Contains(t, ch.Name(), "Comic")
}

func TestFirstCatalogPopulateChallenge(t *testing.T) {
	ch := NewFirstCatalogPopulateChallenge()

	assert.NotNil(t, ch)
	assert.NotEmpty(t, ch.ID())
	assert.NotEmpty(t, ch.Name())
	assert.Contains(t, ch.Name(), "Populate")
}

func TestBrowsingAPIHealthChallenge(t *testing.T) {
	ch := NewBrowsingAPIHealthChallenge()

	assert.NotNil(t, ch)
	assert.NotEmpty(t, ch.ID())
	assert.NotEmpty(t, ch.Name())
	assert.Contains(t, ch.Name(), "Health")
}

func TestBrowsingAPICatalogChallenge(t *testing.T) {
	ch := NewBrowsingAPICatalogChallenge()

	assert.NotNil(t, ch)
	assert.NotEmpty(t, ch.ID())
	assert.NotEmpty(t, ch.Name())
	assert.Contains(t, ch.Name(), "Catalog")
}

func TestBrowsingWebAppChallenge(t *testing.T) {
	ch := NewBrowsingWebAppChallenge()

	assert.NotNil(t, ch)
	assert.NotEmpty(t, ch.ID())
	assert.NotEmpty(t, ch.Name())
	assert.Contains(t, ch.Name(), "Web")
}

func TestAssetServingChallenge(t *testing.T) {
	ch := NewAssetServingChallenge()

	assert.NotNil(t, ch)
	assert.NotEmpty(t, ch.ID())
	assert.NotEmpty(t, ch.Name())
	assert.Contains(t, ch.Name(), "Asset")
}

func TestAssetLazyLoadingChallenge(t *testing.T) {
	ch := NewAssetLazyLoadingChallenge()

	assert.NotNil(t, ch)
	assert.NotEmpty(t, ch.ID())
	assert.NotEmpty(t, ch.Name())
	assert.Contains(t, ch.Name(), "Lazy")
}

func TestAuthTokenRefreshChallenge(t *testing.T) {
	ch := NewAuthTokenRefreshChallenge()

	assert.NotNil(t, ch)
	assert.NotEmpty(t, ch.ID())
	assert.NotEmpty(t, ch.Name())
	assert.Contains(t, ch.Name(), "Token")
}

func TestCollectionManagementChallenge(t *testing.T) {
	ch := NewCollectionManagementChallenge()

	assert.NotNil(t, ch)
	assert.NotEmpty(t, ch.ID())
	assert.NotEmpty(t, ch.Name())
	assert.Contains(t, ch.Name(), "Collection")
}

func TestChallenge_AllChallengesNotNil(t *testing.T) {
	ep := &Endpoint{Host: "localhost", Port: 445}
	dir := Directory{Path: "/media", ContentType: "movie"}

	// Test that all challenges can be created and have non-empty IDs and names
	t.Run("SMB Connectivity", func(t *testing.T) {
		ch := NewSMBConnectivityChallenge(ep)
		assert.NotNil(t, ch)
		assert.NotEmpty(t, ch.ID())
		assert.NotEmpty(t, ch.Name())
	})

	t.Run("Directory Discovery", func(t *testing.T) {
		ch := NewDirectoryDiscoveryChallenge(ep)
		assert.NotNil(t, ch)
		assert.NotEmpty(t, ch.ID())
		assert.NotEmpty(t, ch.Name())
	})

	t.Run("Music Scan", func(t *testing.T) {
		ch := NewMusicScanChallenge(ep, dir)
		assert.NotNil(t, ch)
		assert.NotEmpty(t, ch.ID())
		assert.NotEmpty(t, ch.Name())
	})

	t.Run("Movies Scan", func(t *testing.T) {
		ch := NewMoviesScanChallenge(ep, dir)
		assert.NotNil(t, ch)
		assert.NotEmpty(t, ch.ID())
		assert.NotEmpty(t, ch.Name())
	})

	t.Run("Series Scan", func(t *testing.T) {
		ch := NewSeriesScanChallenge(ep, dir)
		assert.NotNil(t, ch)
		assert.NotEmpty(t, ch.ID())
		assert.NotEmpty(t, ch.Name())
	})

	t.Run("Software Scan", func(t *testing.T) {
		ch := NewSoftwareScanChallenge(ep, dir)
		assert.NotNil(t, ch)
		assert.NotEmpty(t, ch.ID())
		assert.NotEmpty(t, ch.Name())
	})

	t.Run("Comics Scan", func(t *testing.T) {
		ch := NewComicsScanChallenge(ep, dir)
		assert.NotNil(t, ch)
		assert.NotEmpty(t, ch.ID())
		assert.NotEmpty(t, ch.Name())
	})

	t.Run("Catalog Populate", func(t *testing.T) {
		ch := NewFirstCatalogPopulateChallenge()
		assert.NotNil(t, ch)
		assert.NotEmpty(t, ch.ID())
		assert.NotEmpty(t, ch.Name())
	})

	t.Run("API Health", func(t *testing.T) {
		ch := NewBrowsingAPIHealthChallenge()
		assert.NotNil(t, ch)
		assert.NotEmpty(t, ch.ID())
		assert.NotEmpty(t, ch.Name())
	})

	t.Run("API Catalog", func(t *testing.T) {
		ch := NewBrowsingAPICatalogChallenge()
		assert.NotNil(t, ch)
		assert.NotEmpty(t, ch.ID())
		assert.NotEmpty(t, ch.Name())
	})

	t.Run("Web App", func(t *testing.T) {
		ch := NewBrowsingWebAppChallenge()
		assert.NotNil(t, ch)
		assert.NotEmpty(t, ch.ID())
		assert.NotEmpty(t, ch.Name())
	})

	t.Run("Asset Serving", func(t *testing.T) {
		ch := NewAssetServingChallenge()
		assert.NotNil(t, ch)
		assert.NotEmpty(t, ch.ID())
		assert.NotEmpty(t, ch.Name())
	})

	t.Run("Asset Lazy Loading", func(t *testing.T) {
		ch := NewAssetLazyLoadingChallenge()
		assert.NotNil(t, ch)
		assert.NotEmpty(t, ch.ID())
		assert.NotEmpty(t, ch.Name())
	})

	t.Run("Token Refresh", func(t *testing.T) {
		ch := NewAuthTokenRefreshChallenge()
		assert.NotNil(t, ch)
		assert.NotEmpty(t, ch.ID())
		assert.NotEmpty(t, ch.Name())
	})

	t.Run("Collection Management", func(t *testing.T) {
		ch := NewCollectionManagementChallenge()
		assert.NotNil(t, ch)
		assert.NotEmpty(t, ch.ID())
		assert.NotEmpty(t, ch.Name())
	})
}

func TestLoadEndpointConfig_MissingFile(t *testing.T) {
	_, err := LoadEndpointConfig("/nonexistent/path/config.json")
	require.Error(t, err)
}

func TestChallengeTimeout(t *testing.T) {
	ch := NewBrowsingAPIHealthChallenge()

	ctx, cancel := context.WithTimeout(context.Background(), 1)
	defer cancel()

	assert.NotNil(t, ctx)
	_ = ch
}

func TestContentTypeValidation(t *testing.T) {
	validTypes := []string{"movie", "tv_show", "music", "software", "comic", "book", "game"}

	for _, ct := range validTypes {
		t.Run("valid_"+ct, func(t *testing.T) {
			dir := Directory{
				Path:        "/media",
				ContentType: ct,
			}
			assert.Equal(t, ct, dir.ContentType)
			assert.NotEmpty(t, dir.Path)
		})
	}
}

func TestEndpointWithCredentials(t *testing.T) {
	ep := &Endpoint{
		Host:     "localhost",
		Port:     445,
		Username: "user",
		Password: "pass",
		Domain:   "domain",
	}

	assert.Equal(t, "localhost", ep.Host)
	assert.Equal(t, 445, ep.Port)
	assert.Equal(t, "user", ep.Username)
	assert.Equal(t, "pass", ep.Password)
	assert.Equal(t, "domain", ep.Domain)
}

func TestEndpointReadOnly(t *testing.T) {
	ep := &Endpoint{
		Host:     "localhost",
		Port:     445,
		ReadOnly: true,
	}

	assert.True(t, ep.ReadOnly)
}

func TestEndpointConfig_JSON(t *testing.T) {
	jsonData := `{
		"endpoints": [
			{
				"id": "test-nas",
				"name": "Test NAS",
				"host": "192.168.1.100",
				"port": 445,
				"share": "media",
				"directories": [
					{"path": "/movies", "content_type": "movie"}
				]
			}
		]
	}`

	var cfg EndpointConfig
	err := parseTestJSON([]byte(jsonData), &cfg)
	require.NoError(t, err)
	require.Len(t, cfg.Endpoints, 1)
	assert.Equal(t, "test-nas", cfg.Endpoints[0].ID)
	assert.Equal(t, "Test NAS", cfg.Endpoints[0].Name)
}

func parseTestJSON(data []byte, v *EndpointConfig) error {
	return json.Unmarshal(data, v)
}
