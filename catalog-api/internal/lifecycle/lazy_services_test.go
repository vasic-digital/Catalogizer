package lifecycle

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
)

func TestLazyServiceRegistry_Get_InitializesOnFirstAccess(t *testing.T) {
	registry := NewLazyServiceRegistry()

	var callCount int32
	registry.Register("db", func() (interface{}, error) {
		atomic.AddInt32(&callCount, 1)
		return "db-connection", nil
	})

	// First access triggers initialization
	val, err := registry.Get("db")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "db-connection" {
		t.Fatalf("expected %q, got %q", "db-connection", val)
	}
	if atomic.LoadInt32(&callCount) != 1 {
		t.Fatalf("expected init called once, got %d", callCount)
	}

	// Second access returns cached value without re-initializing
	val2, err := registry.Get("db")
	if err != nil {
		t.Fatalf("unexpected error on second get: %v", err)
	}
	if val2 != "db-connection" {
		t.Fatalf("expected cached %q, got %q", "db-connection", val2)
	}
	if atomic.LoadInt32(&callCount) != 1 {
		t.Fatalf("expected init still called once, got %d", callCount)
	}
}

func TestLazyServiceRegistry_Get_UnknownService(t *testing.T) {
	registry := NewLazyServiceRegistry()

	val, err := registry.Get("nonexistent")
	if err == nil {
		t.Fatalf("expected error for unknown service, got val=%v", val)
	}
	if val != nil {
		t.Fatalf("expected nil value, got %v", val)
	}
}

func TestLazyServiceRegistry_Get_InitError(t *testing.T) {
	registry := NewLazyServiceRegistry()

	expectedErr := fmt.Errorf("connection refused")
	registry.Register("broken", func() (interface{}, error) {
		return nil, expectedErr
	})

	val, err := registry.Get("broken")
	if err == nil {
		t.Fatalf("expected error, got val=%v", val)
	}
	if val != nil {
		t.Fatalf("expected nil value on error, got %v", val)
	}
}

func TestLazyServiceRegistry_ConcurrentAccess(t *testing.T) {
	registry := NewLazyServiceRegistry()

	var callCount int32
	registry.Register("shared", func() (interface{}, error) {
		atomic.AddInt32(&callCount, 1)
		return "shared-resource", nil
	})

	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines)

	errors := make(chan error, goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			val, err := registry.Get("shared")
			if err != nil {
				errors <- fmt.Errorf("unexpected error: %w", err)
				return
			}
			if val != "shared-resource" {
				errors <- fmt.Errorf("expected %q, got %v", "shared-resource", val)
			}
		}()
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Error(err)
	}

	if count := atomic.LoadInt32(&callCount); count != 1 {
		t.Fatalf("expected init called exactly once, got %d", count)
	}
}

func TestLazyServiceRegistry_Names(t *testing.T) {
	registry := NewLazyServiceRegistry()

	registry.Register("alpha", func() (interface{}, error) { return nil, nil })
	registry.Register("beta", func() (interface{}, error) { return nil, nil })
	registry.Register("gamma", func() (interface{}, error) { return nil, nil })

	names := registry.Names()
	if len(names) != 3 {
		t.Fatalf("expected 3 names, got %d", len(names))
	}

	nameSet := make(map[string]bool)
	for _, n := range names {
		nameSet[n] = true
	}
	for _, expected := range []string{"alpha", "beta", "gamma"} {
		if !nameSet[expected] {
			t.Errorf("expected name %q in registry, not found", expected)
		}
	}
}

func TestLazyServiceRegistry_ReRegister(t *testing.T) {
	registry := NewLazyServiceRegistry()

	registry.Register("svc", func() (interface{}, error) {
		return "v1", nil
	})

	val, err := registry.Get("svc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "v1" {
		t.Fatalf("expected %q, got %v", "v1", val)
	}

	// Re-register with new init function
	registry.Register("svc", func() (interface{}, error) {
		return "v2", nil
	})

	val2, err := registry.Get("svc")
	if err != nil {
		t.Fatalf("unexpected error after re-register: %v", err)
	}
	if val2 != "v2" {
		t.Fatalf("expected %q after re-register, got %v", "v2", val2)
	}
}
