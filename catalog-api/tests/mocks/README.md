# Mock Servers

**DEPRECATED FOR INTEGRATION TESTS**

These mock servers should ONLY be used for unit tests.

For integration tests, use the real containerized services defined in `docker-compose.test.yml`:

```bash
# Start real test services
podman-compose -f docker-compose.test.yml up -d

# Set environment variables for integration tests
export SMB_TEST_SERVER=localhost
export SMB_TEST_PORT=445
export FTP_TEST_SERVER=localhost
export FTP_TEST_PORT=21
export WEBDAV_TEST_URL=http://localhost:8081
export NFS_TEST_SERVER=localhost
```

The integration tests in `tests/integration/protocol_connectivity_test.go` use these environment variables to connect to real services.

## Mock Files (Unit Tests Only)

- `ftp_mock_server.go` - Mock FTP server for unit testing
- `smb_mock_server.go` - Mock SMB server for unit testing  
- `webdav_mock_server.go` - Mock WebDAV server for unit testing
- `nfs_mock_server.go` - Mock NFS server for unit testing

**Do not use these in integration tests or challenges.**
