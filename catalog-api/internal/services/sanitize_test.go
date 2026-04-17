package services

import (
	"strings"
	"testing"
)

func TestSanitizeMetadataString_StripsScriptTags(t *testing.T) {
	in := `Hello <script>alert("x")</script>world`
	got := SanitizeMetadataString(in)
	if strings.Contains(got, "<script") || strings.Contains(got, "alert") {
		t.Errorf("sanitiser leaked script: %q", got)
	}
	if !strings.Contains(got, "Hello") || !strings.Contains(got, "world") {
		t.Errorf("sanitiser lost legitimate content: %q", got)
	}
}

func TestSanitizeMetadataString_RemovesAllMarkup(t *testing.T) {
	in := `<b>bold</b> <i>italic</i>`
	got := SanitizeMetadataString(in)
	if strings.Contains(got, "<") || strings.Contains(got, ">") {
		t.Errorf("strict policy should remove all tags, got %q", got)
	}
}

func TestSanitizeMetadataString_TrimsWhitespace(t *testing.T) {
	if got := SanitizeMetadataString("  hello  "); got != "hello" {
		t.Errorf("whitespace not trimmed: %q", got)
	}
}

func TestSanitizeMetadataString_NilMapNilResult(t *testing.T) {
	if got := SanitizeCoverFields(nil); got != nil {
		t.Errorf("nil map must return nil, got %+v", got)
	}
}

func TestSanitizeCoverFields_StripsEveryValue(t *testing.T) {
	in := map[string]string{
		"title":       "Matrix <script>alert(1)</script>",
		"description": "<p>Pretty</p>",
		"url":         "https://example.com/cover.jpg",
	}
	got := SanitizeCoverFields(in)
	if strings.Contains(got["title"], "<") || strings.Contains(got["title"], "alert") {
		t.Errorf("title not sanitised: %q", got["title"])
	}
	if strings.Contains(got["description"], "<p>") {
		t.Errorf("description not sanitised: %q", got["description"])
	}
	if got["url"] != "https://example.com/cover.jpg" {
		t.Errorf("innocuous url mangled: %q", got["url"])
	}
}

func TestSanitizeCoverFields_PreservesKeys(t *testing.T) {
	in := map[string]string{"a": "x", "b": "y"}
	got := SanitizeCoverFields(in)
	if len(got) != 2 {
		t.Errorf("len = %d, want 2", len(got))
	}
}
