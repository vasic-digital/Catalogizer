# TEST COVERAGE EXPANSION PLAN
## Target: 95% Coverage Across All Components

---

## 1. CURRENT COVERAGE BASELINE

### Go Backend (catalog-api)

| Package | Current | Target | Gap | Priority |
|---------|---------|--------|-----|----------|
| utils | 100.0% | 95% | ✅ Met | N/A |
| internal/media/models | 100.0% | 95% | ✅ Met | N/A |
| internal/recovery | 99.4% | 95% | ✅ Met | N/A |
| internal/media/detector | 94.6% | 95% | 0.4% | Low |
| internal/metrics | 93.9% | 95% | 1.1% | Low |
| internal/middleware | 93.0% | 95% | 2.0% | Low |
| internal/smb | 91.3% | 95% | 3.7% | Medium |
| internal/config | 90.6% | 95% | 4.4% | Medium |
| middleware | 85.4% | 95% | 9.6% | Medium |
| internal/media/providers | 83.7% | 95% | 11.3% | Medium |
| internal/media/database | 81.1% | 95% | 13.9% | High |
| internal/auth | 75.7% | 95% | 19.3% | High |
| config | 73.8% | 95% | 21.2% | High |
| models | 63.4% | 95% | 31.6% | High |
| repository | 48.6% | 95% | 46.4% | Critical |
| internal/media/analyzer | 40.7% | 95% | 54.3% | Critical |
| database | 40.4% | 95% | 54.6% | Critical |
| internal/handlers | 31.0% | 95% | 64.0% | Critical |
| internal/media/realtime | 31.0% | 95% | 64.0% | Critical |
| handlers | 30.4% | 95% | 64.6% | Critical |
| internal/services | 30.3% | 95% | 64.7% | Critical |
| filesystem | 29.4% | 95% | 65.6% | Critical |
| services | 24.6% | 95% | 70.4% | Critical |
| smb | 16.7% | 95% | 78.3% | Critical |
| challenges | 4.2% | 95% | 90.8% | Critical |

### Frontend (catalog-web)

| Component | Current | Target | Gap | Priority |
|-----------|---------|--------|-----|----------|
| src/components | 92.6% | 95% | 2.4% | Low |
| src/contexts | 98.82% | 95% | ✅ Met | N/A |
| src/hooks | 83.66% | 95% | 11.3% | Medium |
| src/lib | 75.1% | 95% | 19.9% | High |
| src/pages | 44.57% | 95% | 50.4% | Critical |
| src/types | 69.56% | 95% | 25.4% | High |

---

## 2. TEST IMPLEMENTATION PLAN

### 2.1 Go Backend - Critical Priority Packages

#### Package: challenges (4.2% → 95%)

**Current State:**
- Only `config_test.go` exists
- 15,656 lines of challenge code mostly untested

**Tests to Create:**
```go
// challenges/challenges_test.go
package challenges

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

// Test all challenge constructors
func TestNewSMBConnectivityChallenge(t *testing.T) {
    ep := &EndpointConfig{Host: "localhost", Port: 445}
    ch := NewSMBConnectivityChallenge(ep)
    assert.NotNil(t, ch)
    assert.NotEmpty(t, ch.ID())
    assert.NotEmpty(t, ch.Name())
}

func TestNewDirectoryDiscoveryChallenge(t *testing.T) {
    ep := &EndpointConfig{Host: "localhost", Directories: []DirectoryConfig{
        {Path: "/media", ContentType: "movie"},
    }}
    ch := NewDirectoryDiscoveryChallenge(ep)
    assert.NotNil(t, ch)
}

// Test challenge validation
func TestChallengeValidation(t *testing.T) {
    tests := []struct {
        name    string
        config  EndpointConfig
        wantErr bool
    }{
        {"valid config", EndpointConfig{Host: "localhost", Port: 445}, false},
        {"missing host", EndpointConfig{Port: 445}, true},
        {"invalid port", EndpointConfig{Host: "localhost", Port: -1}, true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.config.Validate()
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}

// Test user flow challenges
func TestUserFlowAPIChallenges(t *testing.T) {
    // Mock HTTP server for testing
    srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
    }))
    defer srv.Close()
    
    // Test each user flow challenge
    ch := NewUserFlowAPIHealthChallenge(srv.URL)
    result := ch.Execute(context.Background())
    assert.True(t, result.Success)
}
```

**Estimated Tests:** 150+ tests needed

#### Package: smb (16.7% → 95%)

**Current State:**
- Integration tests skipped
- Mock tests minimal

**Tests to Create:**
```go
// smb/mock_test.go
package smb

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestClient_ParseHostPath(t *testing.T) {
    tests := []struct {
        input     string
        wantHost  string
        wantPath  string
        wantErr   bool
    }{
        {"//server/share/path", "server", "/share/path", false},
        {"\\\\server\\share\\path", "server", "/share/path", false},
        {"invalid", "", "", true},
        {"//server", "server", "", false},
    }
    for _, tt := range tests {
        t.Run(tt.input, func(t *testing.T) {
            host, path, err := ParseHostPath(tt.input)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.wantHost, host)
                assert.Equal(t, tt.wantPath, path)
            }
        })
    }
}

func TestClient_ConnectionPool(t *testing.T) {
    pool := NewConnectionPool(5)
    
    // Test acquire/release
    conn, err := pool.Acquire("server:445")
    assert.NoError(t, err)
    assert.NotNil(t, conn)
    
    pool.Release("server:445", conn)
    assert.Equal(t, 1, pool.Available("server:445"))
}
```

**Estimated Tests:** 80+ tests needed

#### Package: services (24.6% → 95%)

**Tests to Create:**
```go
// services/webdav_client_test.go (expand existing)

func TestWebDAVClient_UploadLargeFile(t *testing.T) {
    // Test chunked upload
    client := NewWebDAVClient(testConfig)
    
    largeFile := generateTestFile(100 * 1024 * 1024) // 100MB
    defer os.Remove(largeFile.Name())
    
    err := client.Upload(context.Background(), largeFile, "/remote/large.bin")
    assert.NoError(t, err)
}

func TestWebDAVClient_ResumeUpload(t *testing.T) {
    // Test upload resume after interruption
    client := NewWebDAVClient(testConfig)
    
    // Simulate partial upload
    partial := createPartialUpload()
    err := client.ResumeUpload(context.Background(), partial)
    assert.NoError(t, err)
}
```

**Estimated Tests:** 200+ tests needed

### 2.2 Frontend - Critical Priority Components

#### Component: src/pages (44.57% → 95%)

**Files Needing Tests:**

##### Playlists.tsx (25.16%)
```typescript
// src/pages/__tests__/Playlists.comprehensive.test.tsx
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import Playlists from '../Playlists';

describe('Playlists', () => {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } }
  });

  const wrapper = ({ children }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );

  describe('Playlist Creation', () => {
    it('creates a new playlist with valid name', async () => {
      render(<Playlists />, { wrapper });
      
      fireEvent.click(screen.getByText('Create Playlist'));
      fireEvent.change(screen.getByPlaceholderText('Playlist name'), {
        target: { value: 'My New Playlist' }
      });
      fireEvent.click(screen.getByText('Save'));
      
      await waitFor(() => {
        expect(screen.getByText('My New Playlist')).toBeInTheDocument();
      });
    });

    it('validates playlist name length', async () => {
      render(<Playlists />, { wrapper });
      
      fireEvent.click(screen.getByText('Create Playlist'));
      fireEvent.change(screen.getByPlaceholderText('Playlist name'), {
        target: { value: 'ab' } // Too short
      });
      fireEvent.click(screen.getByText('Save'));
      
      await waitFor(() => {
        expect(screen.getByText(/at least 3 characters/i)).toBeInTheDocument();
      });
    });
  });

  describe('Playlist Operations', () => {
    it('reorders items via drag and drop', async () => {
      // Test drag and drop functionality
    });

    it('removes items from playlist', async () => {
      // Test removal
    });

    it('exports playlist to M3U', async () => {
      // Test export functionality
    });
  });

  describe('Bulk Operations', () => {
    it('selects multiple items', async () => {
      // Test multi-select
    });

    it('deletes selected items', async () => {
      // Test bulk delete
    });
  });
});
```

**Estimated Tests:** 50+ tests per page component

##### Favorites.tsx (27.27%)
```typescript
// src/pages/__tests__/Favorites.comprehensive.test.tsx
describe('Favorites', () => {
  describe('Display', () => {
    it('shows favorites grouped by type', async () => {
      // Test grouping
    });

    it('shows empty state when no favorites', async () => {
      // Test empty state
    });
  });

  describe('Operations', () => {
    it('removes item from favorites', async () => {
      // Test removal
    });

    it('clears all favorites', async () => {
      // Test clear all
    });
  });
});
```

##### Collections.tsx (41.17%)
```typescript
// src/pages/__tests__/Collections.comprehensive.test.tsx
describe('Collections', () => {
  describe('Collection Management', () => {
    it('creates collection with custom thumbnail', async () => {});
    it('adds items to collection', async () => {});
    it('removes items from collection', async () => {});
    it('deletes collection', async () => {});
    it('shares collection', async () => {});
  });

  describe('Filtering & Sorting', () => {
    it('filters by media type', async () => {});
    it('sorts by date added', async () => {});
    it('sorts by name', async () => {});
    it('searches within collection', async () => {});
  });
});
```

---

## 3. TEST INFRASTRUCTURE

### 3.1 Mock Server Setup

```yaml
# docker-compose.test.yml
version: '3.8'
services:
  mock-smb:
    image: dperson/samba:latest
    environment:
      USER: "test;test123"
      SHARE: "media;/mnt/media;yes;no;no;test"
    ports:
      - "445:445"

  mock-ftp:
    image: stilliard/pure-ftpd:latest
    environment:
      PUBLICHOST: localhost
      FTP_USER_NAME: test
      FTP_USER_PASS: test123
      FTP_USER_HOME: /home/ftpuser
    ports:
      - "21:21"
      - "30000-30009:30000-30009"

  mock-webdav:
    image: bytemark/webdav:latest
    environment:
      AUTH_TYPE: Basic
      USERNAME: test
      PASSWORD: test123
    ports:
      - "8081:80"

  mock-nfs:
    image: erichough/nfs-server:latest
    environment:
      NFS_EXPORT_0: "/mnt/media *(rw,no_subtree_check)"
    cap_add:
      - SYS_ADMIN
    ports:
      - "2049:2049"
```

### 3.2 Test Helper Functions

```go
// internal/tests/helpers.go
package tests

import (
    "context"
    "testing"
    "time"
)

// WaitForCondition polls until condition is true or timeout
func WaitForCondition(t *testing.T, condition func() bool, timeout time.Duration) bool {
    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()
    
    ticker := time.NewTicker(100 * time.Millisecond)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return false
        case <-ticker.C:
            if condition() {
                return true
            }
        }
    }
}

// AssertEventually asserts that condition becomes true within timeout
func AssertEventually(t *testing.T, condition func() bool, timeout time.Duration, msg string) {
    if !WaitForCondition(t, condition, timeout) {
        t.Errorf("Condition not met within %v: %s", timeout, msg)
    }
}

// MockDatabase creates an in-memory SQLite database for testing
func MockDatabase(t *testing.T) *database.DB {
    t.Helper()
    sqlDB, err := sql.Open("sqlite3", ":memory:")
    if err != nil {
        t.Fatalf("Failed to create mock database: %v", err)
    }
    return database.WrapDB(sqlDB, database.DialectSQLite)
}
```

### 3.3 Test Data Fixtures

```go
// internal/tests/fixtures/media.go
package fixtures

import "catalogizer/models"

func SampleMovie() *models.MediaItem {
    return &models.MediaItem{
        ID:          "movie-001",
        Title:       "Test Movie",
        MediaType:   "movie",
        Year:        2024,
        Description: "A test movie for unit testing",
        FilePath:    "/media/movies/test-movie.mkv",
    }
}

func SampleTVShow() *models.MediaItem {
    return &models.MediaItem{
        ID:          "tvshow-001",
        Title:       "Test TV Show",
        MediaType:   "tv_show",
        Year:        2024,
        Description: "A test TV show for unit testing",
    }
}

func SampleMusicAlbum() *models.MediaItem {
    return &models.MediaItem{
        ID:          "album-001",
        Title:       "Test Album",
        MediaType:   "music_album",
        Year:        2024,
        Artist:      "Test Artist",
    }
}
```

---

## 4. COVERAGE TRACKING

### 4.1 Daily Coverage Report Script

```bash
#!/bin/bash
# scripts/coverage-report.sh

echo "=== Coverage Report $(date) ==="

# Go Backend
cd catalog-api
go test ./... -coverprofile=coverage.out -covermode=atomic
go tool cover -func=coverage.out > ../reports/coverage/go-coverage.txt
echo "Go coverage: $(go tool cover -func=coverage.out | tail -1)"

# Frontend
cd ../catalog-web
npm run test:coverage -- --reporter=json --reporter=default
cp coverage/coverage-final.json ../reports/coverage/frontend-coverage.json
echo "Frontend coverage: $(cat coverage/coverage-summary.json | jq '.total.lines.pct')%"

# Generate HTML reports
cd ..
go tool cover -html=catalog-api/coverage.out -o reports/coverage/go-coverage.html
```

### 4.2 Coverage Badge Generation

```bash
# Generate coverage badges for README
go tool cover -func=catalog-api/coverage.out | tail -1 | awk '{print $3}' | sed 's/%//' | \
  xargs -I {} curl -s "https://img.shields.io/badge/coverage-{}%25-brightgreen" > docs/badges/go-coverage.svg
```

---

## 5. MILESTONE CHECKPOINTS

| Week | Target | Verification |
|------|--------|--------------|
| Week 5 | 50% overall | `go test -cover ./... | grep total` |
| Week 6 | 65% overall | `go test -cover ./... | grep total` |
| Week 7 | 80% overall | `go test -cover ./... | grep total` |
| Week 8 | 90% overall | `go test -cover ./... | grep total` |
| Week 9 | 95% overall | Final verification |

---

*Document Generated: 2026-02-27*
*Status: Implementation Ready*
