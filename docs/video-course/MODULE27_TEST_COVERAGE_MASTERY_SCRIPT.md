# Module 27: Test Coverage Mastery

## Video Script — Achieving 95%+ Go and 90%+ Frontend Coverage

### Duration: ~25 minutes

---

### Scene 1: Introduction (2 min)

"High test coverage is not vanity — it's the safety net that lets you refactor fearlessly. In this module, we'll walk through the strategies used to push Catalogizer from ~65% Go / ~46% frontend coverage to 95%+ and 90%+ respectively."

---

### Scene 2: Go Table-Driven Tests (5 min)

**Pattern:** Every handler, service, and repository uses table-driven tests.

```go
func TestSearchFiles(t *testing.T) {
    tests := []struct {
        name    string
        filter  models.SearchFilter
        wantErr bool
        wantMin int
    }{
        {"empty query returns all", models.SearchFilter{}, false, 0},
        {"query by name", models.SearchFilter{Query: "test"}, false, 0},
        {"filter by type", models.SearchFilter{FileType: "video"}, false, 0},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := repo.SearchFiles(ctx, tt.filter, pagination, sort)
            if tt.wantErr { assert.Error(t, err); return }
            assert.NoError(t, err)
            assert.GreaterOrEqual(t, len(result.Files), tt.wantMin)
        })
    }
}
```

**Key:** Test happy path, error path, edge cases (nil input, empty collections, max values).

---

### Scene 3: Test Helper & In-Memory SQLite (3 min)

**File:** `internal/tests/test_helper.go`

```go
db := database.WrapDB(sqlDB, database.DialectSQLite)
```

"Every test gets a fresh in-memory SQLite database. No cleanup needed, no test pollution."

---

### Scene 4: Fuzz Testing (4 min)

```go
func FuzzSearchQuery(f *testing.F) {
    f.Add("normal query")
    f.Add("")
    f.Add("'; DROP TABLE files; --")
    f.Fuzz(func(t *testing.T, query string) {
        _, err := repo.SearchFiles(ctx, models.SearchFilter{Query: query}, pg, sort)
        // Must not panic, must not return unexpected error types
        if err != nil {
            assert.NotContains(t, err.Error(), "syntax error")
        }
    })
}
```

"Fuzz tests find inputs you'd never think to test manually. Target: 20+ fuzz targets covering all API input parsing."

---

### Scene 5: Benchmark Tests (3 min)

```go
func BenchmarkFileSearch(b *testing.B) {
    for i := 0; i < b.N; i++ {
        repo.SearchFiles(ctx, filter, pagination, sort)
    }
}
```

"Run with `go test -bench=. -benchmem` to measure allocations. Target: 50+ benchmarks."

---

### Scene 6: Frontend Testing with Vitest (5 min)

**Pattern:** React Testing Library for behavior-driven tests.

```tsx
import { render, screen, fireEvent } from '@testing-library/react'

describe('MediaCard', () => {
    it('renders media title', () => {
        render(<MediaCard media={mockMedia} />)
        expect(screen.getByText('Test Movie')).toBeInTheDocument()
    })
    it('calls onClick when clicked', async () => {
        const onClick = vi.fn()
        render(<MediaCard media={mockMedia} onClick={onClick} />)
        await userEvent.click(screen.getByRole('article'))
        expect(onClick).toHaveBeenCalledWith(mockMedia)
    })
})
```

**Coverage:** Every component needs: render test, props test, interaction test, error state test.

---

### Scene 7: Hook Testing with renderHook (3 min)

```tsx
import { renderHook, act } from '@testing-library/react'

describe('useFavorites', () => {
    it('toggles favorite state', async () => {
        const { result } = renderHook(() => useFavorites())
        await act(async () => { result.current.toggle(1) })
        expect(result.current.isFavorite(1)).toBe(true)
    })
})
```

---

### Scene 8: Coverage Commands (2 min)

```bash
# Go coverage
cd catalog-api && go test -cover ./... | grep -v "100.0%"

# Frontend coverage
cd catalog-web && npm run test:coverage

# Coverage gates enforced by challenges
# CH-261: Go >= 95%
# CH-266: Frontend >= 90%
```

---

### Summary

- Table-driven tests for all Go code
- In-memory SQLite for isolation
- Fuzz tests for input parsing (20+)
- Benchmarks for performance regression (50+)
- React Testing Library for behavior tests
- renderHook for custom hooks
- Coverage gates enforced by challenge system
