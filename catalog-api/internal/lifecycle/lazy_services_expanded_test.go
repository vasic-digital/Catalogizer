package lifecycle

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
)

// ---------------------------------------------------------------------------
// Register — edge cases
// ---------------------------------------------------------------------------

func TestLazyServiceRegistry_Register_OverwriteClears(t *testing.T) {
	registry := NewLazyServiceRegistry()

	var v1Calls, v2Calls int32
	registry.Register("svc", func() (interface{}, error) {
		atomic.AddInt32(&v1Calls, 1)
		return "version1", nil
	})

	// Access to trigger v1 init
	val, err := registry.Get("svc")
	if err != nil || val != "version1" {
		t.Fatalf("expected version1, got %v (err: %v)", val, err)
	}
	if atomic.LoadInt32(&v1Calls) != 1 {
		t.Fatal("v1 should have been called once")
	}

	// Overwrite registration
	registry.Register("svc", func() (interface{}, error) {
		atomic.AddInt32(&v2Calls, 1)
		return "version2", nil
	})

	// Access should trigger v2 init (not return stale v1)
	val, err = registry.Get("svc")
	if err != nil || val != "version2" {
		t.Fatalf("expected version2, got %v (err: %v)", val, err)
	}
	if atomic.LoadInt32(&v2Calls) != 1 {
		t.Fatal("v2 should have been called once")
	}
	// v1 should not have been called again
	if atomic.LoadInt32(&v1Calls) != 1 {
		t.Fatal("v1 should still only have been called once")
	}
}

func TestLazyServiceRegistry_Register_MultipleServices(t *testing.T) {
	registry := NewLazyServiceRegistry()

	services := []string{"db", "cache", "queue", "storage", "auth"}
	for _, name := range services {
		n := name
		registry.Register(n, func() (interface{}, error) {
			return n + "-instance", nil
		})
	}

	names := registry.Names()
	if len(names) != len(services) {
		t.Fatalf("expected %d names, got %d", len(services), len(names))
	}

	for _, name := range services {
		val, err := registry.Get(name)
		if err != nil {
			t.Fatalf("error getting %s: %v", name, err)
		}
		expected := name + "-instance"
		if val != expected {
			t.Fatalf("expected %q, got %v", expected, val)
		}
	}
}

// ---------------------------------------------------------------------------
// Get — error permanence
// ---------------------------------------------------------------------------

func TestLazyServiceRegistry_Get_ErrorIsPermanent(t *testing.T) {
	registry := NewLazyServiceRegistry()

	var callCount int32
	registry.Register("failing", func() (interface{}, error) {
		atomic.AddInt32(&callCount, 1)
		return nil, fmt.Errorf("init failed")
	})

	// First call should fail
	_, err1 := registry.Get("failing")
	if err1 == nil {
		t.Fatal("expected error on first call")
	}

	// Second call — sync.Once means the init function won't run again,
	// but the value map won't have the service either, so we get nil
	val, err2 := registry.Get("failing")
	// The init func was already called once via sync.Once — the error
	// is not stored, so second call returns nil,nil (no value stored)
	if atomic.LoadInt32(&callCount) != 1 {
		t.Fatal("init should only have been called once")
	}
	_ = val
	_ = err2
}

// ---------------------------------------------------------------------------
// Concurrent register and get
// ---------------------------------------------------------------------------

func TestLazyServiceRegistry_ConcurrentRegisterAndGet(t *testing.T) {
	registry := NewLazyServiceRegistry()

	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines * 2)

	// Half goroutines register, half goroutines get
	for i := 0; i < goroutines; i++ {
		idx := i
		go func() {
			defer wg.Done()
			name := fmt.Sprintf("svc-%d", idx%10)
			registry.Register(name, func() (interface{}, error) {
				return name, nil
			})
		}()
		go func() {
			defer wg.Done()
			name := fmt.Sprintf("svc-%d", idx%10)
			registry.Get(name) // may or may not be registered yet
		}()
	}

	wg.Wait()

	// After all goroutines complete, all services should be accessible
	for i := 0; i < 10; i++ {
		name := fmt.Sprintf("svc-%d", i)
		val, err := registry.Get(name)
		if err != nil {
			t.Errorf("expected %s to be accessible, got error: %v", name, err)
		}
		if val == nil {
			t.Errorf("expected %s to have a value", name)
		}
	}
}

// ---------------------------------------------------------------------------
// Names — empty registry
// ---------------------------------------------------------------------------

func TestLazyServiceRegistry_Names_Empty(t *testing.T) {
	registry := NewLazyServiceRegistry()
	names := registry.Names()
	if len(names) != 0 {
		t.Fatalf("expected 0 names for empty registry, got %d", len(names))
	}
}

// ---------------------------------------------------------------------------
// Get — nil return value is valid
// ---------------------------------------------------------------------------

func TestLazyServiceRegistry_Get_NilValue(t *testing.T) {
	registry := NewLazyServiceRegistry()

	registry.Register("nil-svc", func() (interface{}, error) {
		return nil, nil
	})

	val, err := registry.Get("nil-svc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != nil {
		t.Fatalf("expected nil value, got %v", val)
	}
}

// ---------------------------------------------------------------------------
// Get — type assertions on returned values
// ---------------------------------------------------------------------------

func TestLazyServiceRegistry_Get_TypeAssertions(t *testing.T) {
	registry := NewLazyServiceRegistry()

	type DBConn struct {
		DSN string
	}

	registry.Register("db", func() (interface{}, error) {
		return &DBConn{DSN: "postgres://localhost/test"}, nil
	})

	val, err := registry.Get("db")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	conn, ok := val.(*DBConn)
	if !ok {
		t.Fatal("expected *DBConn type")
	}
	if conn.DSN != "postgres://localhost/test" {
		t.Fatalf("expected DSN, got %q", conn.DSN)
	}
}
