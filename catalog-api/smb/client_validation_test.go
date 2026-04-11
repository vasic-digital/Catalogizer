package smb

import (
	"strings"
	"testing"
	"time"
)

// TestNewSmbClient_MissingFields exercises each validation branch added
// when NewSmbClient stopped trusting its config blindly.
func TestNewSmbClient_MissingFields(t *testing.T) {
	cases := []struct {
		name   string
		cfg    *SmbConfig
		wantIn string
	}{
		{
			name:   "missing host",
			cfg:    &SmbConfig{Port: 445, Share: "share"},
			wantIn: "Host",
		},
		{
			name:   "port zero",
			cfg:    &SmbConfig{Host: "127.0.0.1", Port: 0, Share: "share"},
			wantIn: "Port",
		},
		{
			name:   "port overflow",
			cfg:    &SmbConfig{Host: "127.0.0.1", Port: 99999, Share: "share"},
			wantIn: "Port",
		},
		{
			name:   "port negative",
			cfg:    &SmbConfig{Host: "127.0.0.1", Port: -1, Share: "share"},
			wantIn: "Port",
		},
		{
			name:   "missing share",
			cfg:    &SmbConfig{Host: "127.0.0.1", Port: 445},
			wantIn: "Share",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewSmbClient(tc.cfg)
			if err == nil {
				t.Fatalf("expected error for %s, got nil", tc.name)
			}
			if !strings.Contains(err.Error(), tc.wantIn) {
				t.Errorf("error %q should mention %q", err.Error(), tc.wantIn)
			}
		})
	}
}

// TestNewSmbClient_NilConfigReturnsError verifies NewSmbClient now
// returns a typed error on nil config instead of panicking. The older
// test TestNewSmbClient_NilConfig is tolerant of either behavior; this
// one is strict.
func TestNewSmbClient_NilConfigReturnsError(t *testing.T) {
	client, err := NewSmbClient(nil)
	if err == nil {
		t.Fatal("expected error for nil config, got nil")
	}
	if client != nil {
		t.Error("expected nil client when config is nil")
	}
	if !strings.Contains(err.Error(), "nil") {
		t.Errorf("error should mention nil, got %q", err.Error())
	}
}

// TestNewSmbClient_UnreachableHostFastFail verifies the bounded dial
// timeout causes a quick failure on an unreachable host. The pre-fix
// code used net.Dial which inherited the OS TCP timeout (~2 min).
func TestNewSmbClient_UnreachableHostFastFail(t *testing.T) {
	cfg := &SmbConfig{
		Host:     "192.0.2.1", // TEST-NET-1, guaranteed unroutable
		Port:     445,
		Share:    "share",
		Username: "user",
		Password: "pass",
	}
	start := time.Now()
	_, err := NewSmbClient(cfg)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected error for unreachable host, got nil")
	}
	// The dial timeout is 5s; give generous CI slack but fail loudly
	// if the OS default is back.
	if elapsed > 15*time.Second {
		t.Errorf("dial should fail fast with bounded timeout; took %v", elapsed)
	}
}
