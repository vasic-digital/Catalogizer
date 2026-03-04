# Fuzz Testing Guide

This project uses Go's built-in fuzz testing framework (`testing.F`) to find crashes, panics, and invariant violations by generating random inputs.

## Overview

Fuzz tests are located alongside their target code in `*_fuzz_test.go` files. The project has 14 fuzz functions across 4 areas:

| Area | File | Functions | Purpose |
|------|------|-----------|---------|
| SQL Dialect | `database/dialect_fuzz_test.go` | 4 | Verify SQL rewriting never produces invalid output |
| Config Parsing | `filesystem/factory_fuzz_test.go` | 2 | Verify setting extraction handles all input types |
| Title Parsing | `internal/services/title_parser_fuzz_test.go` | 7 | Verify media title parsers never crash and produce valid years |
| Security | `internal/handlers/download_fuzz_test.go` | 2 | Verify path traversal and header injection prevention |

## Running Fuzz Tests

Go only runs one fuzz target per `go test` invocation. Specify the target with `-fuzz`:

```bash
cd catalog-api

# Run a specific fuzz target for 30 seconds
go test -fuzz=FuzzParseMovieTitle -fuzztime=30s ./internal/services/

# Run for a specific number of iterations
go test -fuzz=FuzzSanitizeArchivePath -fuzztime=10000x ./internal/handlers/

# Run indefinitely until a failure is found (Ctrl+C to stop)
go test -fuzz=FuzzRewritePlaceholders ./database/

# Run all fuzz targets in a package sequentially (10 seconds each)
for f in FuzzRewritePlaceholders FuzzRewriteInsertOrIgnore FuzzRewriteBooleanLiterals FuzzRewriteInsertOrReplace; do
  go test -fuzz=$f -fuzztime=10s ./database/
done
```

Resource-limited execution (recommended for this project):

```bash
GOMAXPROCS=3 go test -fuzz=FuzzParseMovieTitle -fuzztime=30s ./internal/services/
```

## Seed Corpus

Each fuzz function provides seed inputs via `f.Add()`. These seeds cover known edge cases:

```go
func FuzzSanitizeArchivePath(f *testing.F) {
    seeds := []string{
        "",
        "../../etc/passwd",       // path traversal
        "foo/bar/\x00baz",        // null byte injection
        "normal_file.txt",        // valid input
        strings.Repeat("../", 100) + "etc/passwd",  // deep traversal
    }
    for _, s := range seeds {
        f.Add(s)
    }
    // ...
}
```

Go stores crash-triggering inputs in `testdata/fuzz/<FunctionName>/` directories automatically. These corpus entries are version-controlled and replayed on every `go test` run (even without `-fuzz`).

## Writing Invariants

Fuzz tests verify **invariants** -- properties that must hold for all inputs:

```go
f.Fuzz(func(t *testing.T, input string) {
    result := SomeFunction(input)

    // Invariant 1: must not panic (implicit -- reaching here means no panic)

    // Invariant 2: output must satisfy a property
    if strings.HasPrefix(result, "/") {
        t.Errorf("SomeFunction(%q) = %q starts with /", input, result)
    }

    // Invariant 3: SQLite dialect must be identity
    sq := &Dialect{Type: DialectSQLite}
    if sq.Rewrite(input) != input {
        t.Errorf("SQLite should return input unchanged")
    }
})
```

### Common Invariant Patterns Used in This Project

**No-panic invariant** (implicit): Every fuzz target verifies the function does not panic on any input.

**Range invariant** (title parsers): If a year is extracted, it must be in [1900, 2099]:

```go
if result.Year != nil {
    y := *result.Year
    if y < 1900 || y > 2099 {
        t.Errorf("year %d outside valid range", y)
    }
}
```

**Security invariant** (download handler): Output must never contain path traversal sequences:

```go
if strings.HasPrefix(result, "../") {
    t.Errorf("result starts with ../")
}
if strings.Contains(result, "/../") {
    t.Errorf("result contains /../")
}
```

**Identity invariant** (dialect rewriting): SQLite dialect must return the query unchanged:

```go
sq := &Dialect{Type: DialectSQLite}
if sq.RewritePlaceholders(query) != query {
    t.Errorf("SQLite should not rewrite")
}
```

## Writing a New Fuzz Test

1. Create a `*_fuzz_test.go` file next to the code under test.
2. Name the function `Fuzz<TargetFunction>`.
3. Add seed corpus entries covering edge cases (empty strings, unicode, null bytes, boundary values).
4. Define invariants that must hold for all inputs.

Example template:

```go
package mypackage

import "testing"

func FuzzMyFunction(f *testing.F) {
    f.Add("")
    f.Add("normal input")
    f.Add("\x00\x01\x02")
    f.Add("edge case value")

    f.Fuzz(func(t *testing.T, input string) {
        result := MyFunction(input)

        // Define invariants here
        if !IsValid(result) {
            t.Errorf("MyFunction(%q) produced invalid result: %v", input, result)
        }
    })
}
```

## Viewing the Corpus

```bash
# List saved corpus entries for a fuzz target
ls catalog-api/testdata/fuzz/FuzzParseMovieTitle/

# Each file contains a single input that triggered a failure
cat catalog-api/testdata/fuzz/FuzzParseMovieTitle/corpus_entry_name
```

## Integration with Regular Tests

Fuzz seed corpus entries are always run during normal `go test` (without the `-fuzz` flag). This means any crash-triggering input found by fuzzing becomes a permanent regression test.
