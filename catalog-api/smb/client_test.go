package smb

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"testing"
)

func TestSmbConfig_Initialization(t *testing.T) {
	tests := []struct {
		name   string
		config SmbConfig
		want   SmbConfig
	}{
		{
			name: "standard config",
			config: SmbConfig{
				Host:     "192.168.1.100",
				Port:     445,
				Share:    "media",
				Username: "admin",
				Password: "secret",
				Domain:   "WORKGROUP",
			},
			want: SmbConfig{
				Host:     "192.168.1.100",
				Port:     445,
				Share:    "media",
				Username: "admin",
				Password: "secret",
				Domain:   "WORKGROUP",
			},
		},
		{
			name: "hostname config",
			config: SmbConfig{
				Host:     "fileserver.local",
				Port:     445,
				Share:    "shared",
				Username: "user",
				Password: "pass",
				Domain:   "",
			},
			want: SmbConfig{
				Host:     "fileserver.local",
				Port:     445,
				Share:    "shared",
				Username: "user",
				Password: "pass",
				Domain:   "",
			},
		},
		{
			name: "custom port",
			config: SmbConfig{
				Host: "10.0.0.1",
				Port: 4455,
			},
			want: SmbConfig{
				Host: "10.0.0.1",
				Port: 4455,
			},
		},
		{
			name:   "zero value config",
			config: SmbConfig{},
			want:   SmbConfig{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.config.Host != tt.want.Host {
				t.Errorf("Host = %q, want %q", tt.config.Host, tt.want.Host)
			}
			if tt.config.Port != tt.want.Port {
				t.Errorf("Port = %d, want %d", tt.config.Port, tt.want.Port)
			}
			if tt.config.Share != tt.want.Share {
				t.Errorf("Share = %q, want %q", tt.config.Share, tt.want.Share)
			}
			if tt.config.Username != tt.want.Username {
				t.Errorf("Username = %q, want %q", tt.config.Username, tt.want.Username)
			}
			if tt.config.Password != tt.want.Password {
				t.Errorf("Password = %q, want %q", tt.config.Password, tt.want.Password)
			}
			if tt.config.Domain != tt.want.Domain {
				t.Errorf("Domain = %q, want %q", tt.config.Domain, tt.want.Domain)
			}
		})
	}
}

func TestSmbConfig_AddressFormatting(t *testing.T) {
	tests := []struct {
		name     string
		config   SmbConfig
		wantAddr string
	}{
		{
			name:     "standard SMB port",
			config:   SmbConfig{Host: "192.168.1.100", Port: 445},
			wantAddr: "192.168.1.100:445",
		},
		{
			name:     "custom port",
			config:   SmbConfig{Host: "10.0.0.50", Port: 4455},
			wantAddr: "10.0.0.50:4455",
		},
		{
			name:     "hostname with port",
			config:   SmbConfig{Host: "fileserver.example.com", Port: 445},
			wantAddr: "fileserver.example.com:445",
		},
		{
			name:     "zero port",
			config:   SmbConfig{Host: "host", Port: 0},
			wantAddr: "host:0",
		},
		{
			name:     "IPv6 host",
			config:   SmbConfig{Host: "::1", Port: 445},
			wantAddr: "[::1]:445",
		},
		{
			name:     "empty host",
			config:   SmbConfig{Host: "", Port: 445},
			wantAddr: ":445",
		},
		{
			name:     "high port",
			config:   SmbConfig{Host: "server", Port: 65535},
			wantAddr: "server:65535",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Replicate the address formatting logic from NewSmbClient
			// using net.JoinHostPort to correctly handle IPv6 addresses
			addr := net.JoinHostPort(tt.config.Host, fmt.Sprintf("%d", tt.config.Port))
			if addr != tt.wantAddr {
				t.Errorf("address = %q, want %q", addr, tt.wantAddr)
			}
		})
	}
}

func TestIsNotExistError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "file does not exist",
			err:  errors.New("file does not exist"),
			want: true,
		},
		{
			name: "no such file or directory",
			err:  errors.New("no such file or directory"),
			want: true,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "permission denied",
			err:  errors.New("permission denied"),
			want: false,
		},
		{
			name: "connection reset",
			err:  errors.New("connection reset by peer"),
			want: false,
		},
		{
			name: "empty error string",
			err:  errors.New(""),
			want: false,
		},
		{
			name: "similar but different message uppercase",
			err:  errors.New("File does not exist"),
			want: false,
		},
		{
			name: "similar but different message with prefix",
			err:  errors.New("error: file does not exist"),
			want: false,
		},
		{
			name: "wrapped error with different message",
			err:  fmt.Errorf("wrapped: %w", errors.New("some error")),
			want: false,
		},
		{
			name: "access denied",
			err:  errors.New("access denied"),
			want: false,
		},
		{
			name: "network error",
			err:  errors.New("network unreachable"),
			want: false,
		},
		{
			name: "partial match - file does not",
			err:  errors.New("file does not"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isNotExistError(tt.err)
			if got != tt.want {
				t.Errorf("isNotExistError(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

func TestNewSmbClient_NilConfig(t *testing.T) {
	// NewSmbClient with nil config should panic (nil pointer dereference
	// when accessing config.Host) or return an error.
	defer func() {
		if r := recover(); r != nil {
			// Panicking on nil config is acceptable behavior
			return
		}
	}()

	client, err := NewSmbClient(nil)
	if err == nil && client != nil {
		t.Error("expected error or panic with nil config")
	}
}

func TestSmbClient_Close_NilFields(t *testing.T) {
	// A client with all nil internal fields should not panic on Close
	client := &SmbClient{
		conn:    nil,
		session: nil,
		share:   nil,
		config:  nil,
	}

	err := client.Close()
	if err != nil {
		t.Errorf("Close on nil-field client should return nil, got %v", err)
	}
}

func TestSmbClient_Close_PartialNilFields(t *testing.T) {
	tests := []struct {
		name   string
		client *SmbClient
	}{
		{
			name: "only config set",
			client: &SmbClient{
				conn:    nil,
				session: nil,
				share:   nil,
				config:  &SmbConfig{Host: "test", Port: 445},
			},
		},
		{
			name: "all nil",
			client: &SmbClient{
				conn:    nil,
				session: nil,
				share:   nil,
				config:  nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.client.Close()
			if err != nil {
				t.Errorf("Close should return nil for nil conn/session/share, got %v", err)
			}
		})
	}
}

func TestSmbClient_GetConfig(t *testing.T) {
	config := &SmbConfig{
		Host:     "server.local",
		Port:     445,
		Share:    "data",
		Username: "admin",
		Password: "password123",
		Domain:   "CORP",
	}

	client := &SmbClient{
		config: config,
	}

	got := client.GetConfig()
	if got != config {
		t.Error("GetConfig should return the same config pointer")
	}
	if got.Host != "server.local" {
		t.Errorf("GetConfig().Host = %q, want %q", got.Host, "server.local")
	}
	if got.Port != 445 {
		t.Errorf("GetConfig().Port = %d, want %d", got.Port, 445)
	}
	if got.Share != "data" {
		t.Errorf("GetConfig().Share = %q, want %q", got.Share, "data")
	}
	if got.Username != "admin" {
		t.Errorf("GetConfig().Username = %q, want %q", got.Username, "admin")
	}
	if got.Domain != "CORP" {
		t.Errorf("GetConfig().Domain = %q, want %q", got.Domain, "CORP")
	}
}

func TestSmbClient_GetConfig_Nil(t *testing.T) {
	client := &SmbClient{
		config: nil,
	}

	got := client.GetConfig()
	if got != nil {
		t.Errorf("GetConfig should return nil when config is nil, got %v", got)
	}
}

func TestSmbClient_GetConfig_Mutability(t *testing.T) {
	// Verify that GetConfig returns the original pointer, meaning mutations
	// to the returned config affect the client's config.
	config := &SmbConfig{
		Host: "original.local",
		Port: 445,
	}

	client := &SmbClient{config: config}

	retrieved := client.GetConfig()
	retrieved.Host = "modified.local"

	if client.GetConfig().Host != "modified.local" {
		t.Error("GetConfig should return the same pointer; mutation should be visible")
	}
}

func TestSmbConfig_FieldDefaults(t *testing.T) {
	// Verify zero-value behavior of SmbConfig
	var config SmbConfig

	if config.Host != "" {
		t.Errorf("default Host should be empty, got %q", config.Host)
	}
	if config.Port != 0 {
		t.Errorf("default Port should be 0, got %d", config.Port)
	}
	if config.Share != "" {
		t.Errorf("default Share should be empty, got %q", config.Share)
	}
	if config.Username != "" {
		t.Errorf("default Username should be empty, got %q", config.Username)
	}
	if config.Password != "" {
		t.Errorf("default Password should be empty, got %q", config.Password)
	}
	if config.Domain != "" {
		t.Errorf("default Domain should be empty, got %q", config.Domain)
	}
}

func TestSmbConfig_SpecialCharacters(t *testing.T) {
	tests := []struct {
		name   string
		config SmbConfig
	}{
		{
			name: "password with special chars",
			config: SmbConfig{
				Host:     "server",
				Port:     445,
				Share:    "share$",
				Username: "domain\\user",
				Password: "p@ss!w0rd#$%",
				Domain:   "MY.DOMAIN",
			},
		},
		{
			name: "unicode in paths",
			config: SmbConfig{
				Host:  "server",
				Port:  445,
				Share: "shared-folder",
			},
		},
		{
			name: "share with dollar sign (admin share)",
			config: SmbConfig{
				Host:  "server",
				Port:  445,
				Share: "C$",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify the config stores special characters correctly
			if tt.name == "password with special chars" {
				if tt.config.Password != "p@ss!w0rd#$%" {
					t.Errorf("Password not stored correctly: %q", tt.config.Password)
				}
				if tt.config.Username != "domain\\user" {
					t.Errorf("Username not stored correctly: %q", tt.config.Username)
				}
				if tt.config.Share != "share$" {
					t.Errorf("Share not stored correctly: %q", tt.config.Share)
				}
			}
			if tt.name == "share with dollar sign (admin share)" {
				if tt.config.Share != "C$" {
					t.Errorf("Share not stored correctly: %q", tt.config.Share)
				}
			}

			// Verify address formatting works with any config
			// using net.JoinHostPort to correctly handle IPv6 addresses
			addr := net.JoinHostPort(tt.config.Host, fmt.Sprintf("%d", tt.config.Port))
			if !strings.Contains(addr, tt.config.Host) {
				t.Errorf("address %q should contain host %q", addr, tt.config.Host)
			}
		})
	}
}
