package handlers

import (
	"strings"
	"testing"
)

func FuzzSanitizeArchivePath(f *testing.F) {
	// Seed corpus: known edge cases for path traversal
	seeds := []string{
		"",
		".",
		"..",
		"../",
		"../../etc/passwd",
		"../../../etc/shadow",
		"/etc/passwd",
		"foo/bar/baz.txt",
		"foo/../bar",
		"foo/../../bar",
		"./foo/bar",
		"foo/bar/../../../etc/passwd",
		"foo/bar/\x00baz",
		"foo\x00bar",
		"normal_file.txt",
		"dir/sub/file.mp4",
		"....",
		"..../",
		"....//",
		".../test",
		"foo/bar/",
		"/",
		"//",
		"///",
		"a/../b/../c/../d",
		strings.Repeat("../", 100) + "etc/passwd",
		"foo\\..\\bar",
		"foo\\bar\\baz",
		"\t\n\r file.txt",
		"файл.txt",
		"文件.txt",
		"file name with spaces.txt",
		"file%2F..%2F..%2Fetc%2Fpasswd",
	}

	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, input string) {
		result := sanitizeArchivePath(input)

		// Invariant 1: result must never be empty
		if result == "" {
			t.Errorf("sanitizeArchivePath(%q) returned empty string", input)
		}

		// Invariant 2: result must not start with /
		if strings.HasPrefix(result, "/") {
			t.Errorf("sanitizeArchivePath(%q) = %q starts with /", input, result)
		}

		// Invariant 3: result must not start with ../
		if strings.HasPrefix(result, "../") {
			t.Errorf("sanitizeArchivePath(%q) = %q starts with ../", input, result)
		}

		// Invariant 4: result must not be ".."
		if result == ".." {
			t.Errorf("sanitizeArchivePath(%q) = %q is '..'", input, result)
		}

		// Invariant 5: result must not contain /../ (path traversal within path)
		if strings.Contains(result, "/../") {
			t.Errorf("sanitizeArchivePath(%q) = %q contains /../", input, result)
		}
	})
}

func FuzzSanitizeContentDisposition(f *testing.F) {
	// Seed corpus: known edge cases for header injection
	seeds := []string{
		"",
		"normal_file.txt",
		"file with spaces.mp4",
		`file"with"quotes.txt`,
		"file\r\nInjected-Header: value",
		"file\rpartial",
		"file\nnewline",
		"\r\n\r\n",
		`"; filename="malicious.exe`,
		"file\x00null.txt",
		"файл.txt",
		"文件名.mp4",
		strings.Repeat("a", 10000),
		"file\t\ttabs.txt",
		"file\r\nContent-Type: text/html\r\n\r\n<script>alert(1)</script>",
	}

	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, input string) {
		result := sanitizeContentDisposition(input)

		// Invariant 1: result must not contain double quotes
		if strings.Contains(result, "\"") {
			t.Errorf("sanitizeContentDisposition(%q) = %q contains double quote", input, result)
		}

		// Invariant 2: result must not contain carriage return
		if strings.Contains(result, "\r") {
			t.Errorf("sanitizeContentDisposition(%q) = %q contains \\r", input, result)
		}

		// Invariant 3: result must not contain newline
		if strings.Contains(result, "\n") {
			t.Errorf("sanitizeContentDisposition(%q) = %q contains \\n", input, result)
		}
	})
}
