package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"go.uber.org/zap"
)

// newChatCompletionsServer stands up an in-process OpenAI-compatible
// /chat/completions endpoint. On a non-200 status it returns that status
// with an error body (simulating DeepSeek/OpenRouter HTTP 402, rate-limit
// 429, or a 5xx). On 200 it returns a valid chat-completions envelope
// whose message content is the metadata JSON carrying coverURL. hits, when
// non-nil, counts every request so tests can assert a dead candidate is
// not re-probed and a healthy first candidate short-circuits the rest.
func newChatCompletionsServer(t *testing.T, status int, coverURL string, hits *int32) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if hits != nil {
			atomic.AddInt32(hits, 1)
		}
		if status != http.StatusOK {
			w.WriteHeader(status)
			_, _ = io.WriteString(w, `{"error":{"message":"backend unavailable"}}`)
			return
		}
		content := fmt.Sprintf(
			`{"title":"Inception","year":2010,"description":"A thief who steals secrets.","cover_url":%q,"rating":8.8}`,
			coverURL,
		)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]string{"content": content}},
			},
		})
	}))
	t.Cleanup(srv.Close)
	return srv
}

// testCandidate builds an llmCandidate pointed at an in-process server.
func testCandidate(name, baseURL string) llmCandidate {
	return llmCandidate{name: name, apiKey: "test-key-" + name, baseURL: baseURL, model: "test-model"}
}

// TestLLMProviderSearchFailsOverToWorkingCandidate is the RED test: with
// the no-failover code that binds to the first candidate, candidate A
// returning HTTP 402 makes Search give up and error. With failover, Search
// advances to candidate B (HTTP 200) and returns B's result.
func TestLLMProviderSearchFailsOverToWorkingCandidate(t *testing.T) {
	const coverB = "https://images.example.com/inception_b.jpg"
	var hitsA, hitsB int32

	serverA := newChatCompletionsServer(t, http.StatusPaymentRequired, "", &hitsA) // 402, out of credit
	serverB := newChatCompletionsServer(t, http.StatusOK, coverB, &hitsB)          // 200, funded & working

	provider := newLLMProviderWithCandidates(http.DefaultClient, zap.NewNop(), []llmCandidate{
		testCandidate("deepseek", serverA.URL),
		testCandidate("groq", serverB.URL),
	})

	results, err := provider.Search(context.Background(), "Inception", "movie", nil)
	if err != nil {
		t.Fatalf("expected failover to the working candidate, got error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result from the working candidate, got %d", len(results))
	}
	if results[0].CoverURL == nil || *results[0].CoverURL != coverB {
		t.Fatalf("expected cover from candidate B (%q), got %v", coverB, results[0].CoverURL)
	}
	if got := atomic.LoadInt32(&hitsA); got != 1 {
		t.Fatalf("expected candidate A probed exactly once, got %d", got)
	}
	if got := atomic.LoadInt32(&hitsB); got != 1 {
		t.Fatalf("expected candidate B served exactly once, got %d", got)
	}
	if name := provider.GetName(); name != "llm-groq" {
		t.Fatalf("expected GetName to reflect the active working candidate llm-groq, got %q", name)
	}
}

// TestLLMProviderSearchAllCandidatesFail proves there is no false success:
// when every candidate returns a payment/quota error, Search returns an
// error and no result (never a bluff PASS).
func TestLLMProviderSearchAllCandidatesFail(t *testing.T) {
	serverA := newChatCompletionsServer(t, http.StatusPaymentRequired, "", nil) // 402
	serverB := newChatCompletionsServer(t, http.StatusTooManyRequests, "", nil) // 429

	provider := newLLMProviderWithCandidates(http.DefaultClient, zap.NewNop(), []llmCandidate{
		testCandidate("deepseek", serverA.URL),
		testCandidate("openrouter", serverB.URL),
	})

	results, err := provider.Search(context.Background(), "Inception", "movie", nil)
	if err == nil {
		t.Fatalf("expected an error when all candidates fail, got results: %v", results)
	}
	if results != nil {
		t.Fatalf("expected no results when all candidates fail, got: %v", results)
	}
}

// TestLLMProviderSearchFirstCandidateUsedSecondNeverCalled proves a healthy
// first candidate short-circuits failover: candidate B must never be hit.
func TestLLMProviderSearchFirstCandidateUsedSecondNeverCalled(t *testing.T) {
	const coverA = "https://images.example.com/inception_a.jpg"
	var hitsA, hitsB int32

	serverA := newChatCompletionsServer(t, http.StatusOK, coverA, &hitsA) // 200
	serverB := newChatCompletionsServer(t, http.StatusOK, "unused", &hitsB)

	provider := newLLMProviderWithCandidates(http.DefaultClient, zap.NewNop(), []llmCandidate{
		testCandidate("deepseek", serverA.URL),
		testCandidate("groq", serverB.URL),
	})

	results, err := provider.Search(context.Background(), "Inception", "movie", nil)
	if err != nil {
		t.Fatalf("expected success from the first candidate, got error: %v", err)
	}
	if len(results) != 1 || results[0].CoverURL == nil || *results[0].CoverURL != coverA {
		t.Fatalf("expected cover from candidate A (%q), got %+v", coverA, results)
	}
	if got := atomic.LoadInt32(&hitsB); got != 0 {
		t.Fatalf("expected candidate B never called when A succeeds, got %d hits", got)
	}
}

// TestLLMProviderSearchRemembersWorkingCandidate proves the dead first
// candidate is not re-probed first on every call: after the initial
// failover from A (402) to B (200), the next Search starts at B directly,
// so A is probed only on the first call.
func TestLLMProviderSearchRemembersWorkingCandidate(t *testing.T) {
	const coverB = "https://images.example.com/inception_b.jpg"
	var hitsA, hitsB int32

	serverA := newChatCompletionsServer(t, http.StatusPaymentRequired, "", &hitsA) // 402 always
	serverB := newChatCompletionsServer(t, http.StatusOK, coverB, &hitsB)          // 200 always

	provider := newLLMProviderWithCandidates(http.DefaultClient, zap.NewNop(), []llmCandidate{
		testCandidate("deepseek", serverA.URL),
		testCandidate("groq", serverB.URL),
	})

	for i := 0; i < 2; i++ {
		results, err := provider.Search(context.Background(), "Inception", "movie", nil)
		if err != nil {
			t.Fatalf("call %d: expected failover success, got error: %v", i+1, err)
		}
		if len(results) != 1 || results[0].CoverURL == nil || *results[0].CoverURL != coverB {
			t.Fatalf("call %d: expected cover from candidate B, got %+v", i+1, results)
		}
	}

	if got := atomic.LoadInt32(&hitsA); got != 1 {
		t.Fatalf("expected dead candidate A probed only on the first call, got %d hits", got)
	}
	if got := atomic.LoadInt32(&hitsB); got != 2 {
		t.Fatalf("expected candidate B served both calls, got %d hits", got)
	}
}

// TestLLMProviderDisabledWhenNoCandidates verifies the disabled fallback
// path: no candidates -> not enabled, Search errors, name is llm-fallback.
func TestLLMProviderDisabledWhenNoCandidates(t *testing.T) {
	provider := newLLMProviderWithCandidates(http.DefaultClient, zap.NewNop(), nil)
	if provider.IsEnabled() {
		t.Fatalf("expected provider disabled with no candidates")
	}
	if name := provider.GetName(); name != "llm-fallback" {
		t.Fatalf("expected llm-fallback name, got %q", name)
	}
	if _, err := provider.Search(context.Background(), "Inception", "movie", nil); err == nil {
		t.Fatalf("expected error from a disabled provider")
	}
}
