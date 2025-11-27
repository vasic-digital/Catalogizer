package filesystem

import (
	"testing"
)

func TestDefaultClientFactory_SupportedProtocols(t *testing.T) {
	factory := NewDefaultClientFactory()

	protocols := factory.SupportedProtocols()

	expected := []string{"smb", "ftp", "nfs", "webdav", "local"}
	if len(protocols) != len(expected) {
		t.Errorf("Expected %d protocols, got %d", len(expected), len(protocols))
	}

	for i, protocol := range expected {
		if i >= len(protocols) || protocols[i] != protocol {
			t.Errorf("Expected protocol %s at index %d, got %s", protocol, i, protocols[i])
		}
	}
}

func TestDefaultClientFactory_CreateClient(t *testing.T) {
	factory := NewDefaultClientFactory()

	tests := []struct {
		name     string
		config   *StorageConfig
		wantErr  bool
		protocol string
	}{
		{
			name: "SMB client",
			config: &StorageConfig{
				Protocol: "smb",
				Settings: map[string]interface{}{
					"host":     "localhost",
					"port":     445,
					"share":    "test",
					"username": "user",
					"password": "pass",
					"domain":   "WORKGROUP",
				},
			},
			wantErr:  false,
			protocol: "smb",
		},
		{
			name: "FTP client",
			config: &StorageConfig{
				Protocol: "ftp",
				Settings: map[string]interface{}{
					"host":     "localhost",
					"port":     21,
					"username": "user",
					"password": "pass",
					"path":     "/",
				},
			},
			wantErr:  false,
			protocol: "ftp",
		},
		{
			name: "NFS client",
			config: &StorageConfig{
				Protocol: "nfs",
				Settings: map[string]interface{}{
					"host":        "localhost",
					"path":        "/export",
					"mount_point": "/mnt/nfs",
					"options":     "vers=3",
				},
			},
			wantErr:  false,
			protocol: "nfs",
		},
		{
			name: "WebDAV client",
			config: &StorageConfig{
				Protocol: "webdav",
				Settings: map[string]interface{}{
					"url":      "http://localhost/webdav",
					"username": "user",
					"password": "pass",
					"path":     "/",
				},
			},
			wantErr:  false,
			protocol: "webdav",
		},
		{
			name: "Local client",
			config: &StorageConfig{
				Protocol: "local",
				Settings: map[string]interface{}{
					"base_path": "/tmp",
				},
			},
			wantErr:  false,
			protocol: "local",
		},
		{
			name: "Unsupported protocol",
			config: &StorageConfig{
				Protocol: "unsupported",
			},
			wantErr:  true,
			protocol: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := factory.CreateClient(tt.config)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if client == nil {
				t.Error("Expected client, got nil")
				return
			}

			if client.GetProtocol() != tt.protocol {
				t.Errorf("Expected protocol %s, got %s", tt.protocol, client.GetProtocol())
			}
		})
	}
}
