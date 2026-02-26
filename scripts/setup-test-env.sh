#!/bin/bash
#
# Test Environment Setup Script
# Starts test infrastructure for integration tests
#

set -e

echo "=== Setting up Test Environment ==="

# Check if docker-compose.test-infra.yml exists
if [ ! -f "docker-compose.test-infra.yml" ]; then
    echo "Error: docker-compose.test-infra.yml not found"
    echo "Creating minimal test infrastructure..."
    
    # Create a minimal test infra compose file
    cat > docker-compose.test-infra.yml << 'EOF'
version: '3.8'

services:
  # FTP Server for testing
  ftp-server:
    image: docker.io/pureftpd/pure-ftpd
    container_name: catalogizer-ftp-test
    ports:
      - "2121:21"
    environment:
      PUBLICHOST: localhost
    command: -O clf:/var/log/pure-ftpd/transfer.log
    networks:
      - test-network

  # SMB Server for testing  
  smb-server:
    image: docker.io/dperson/samba
    container_name: catalogizer-smb-test
    ports:
      - "1445:445"
    environment:
      USER: testuser;testpass
      SHARE: testshare;/mount;yes;no;no;testuser
    networks:
      - test-network

networks:
  test-network:
    driver: bridge
EOF
    echo "Created docker-compose.test-infra.yml"
fi

# Try to start test infrastructure
echo "Starting test infrastructure..."
if command -v podman-compose > /dev/null 2>&1; then
    podman-compose -f docker-compose.test-infra.yml up -d 2>/dev/null || echo "Note: Could not start test infrastructure (Podman may not be running)"
elif command -v docker-compose > /dev/null 2>&1; then
    docker-compose -f docker-compose.test-infra.yml up -d 2>/dev/null || echo "Note: Could not start test infrastructure (Docker may not be running)"
else
    echo "Warning: Neither podman-compose nor docker-compose found"
    echo "Integration tests requiring external services will be skipped"
fi

# Wait for services
echo "Waiting for test services to be ready..."
sleep 5

# Verify services
echo ""
echo "Checking service availability:"

# Check SMB
if timeout 2 bash -c "</dev/tcp/localhost/1445" 2>/dev/null; then
    echo "  ✓ SMB server: localhost:1445"
else
    echo "  - SMB server: localhost:1445 (not available)"
fi

# Check FTP
if timeout 2 bash -c "</dev/tcp/localhost/2121" 2>/dev/null; then
    echo "  ✓ FTP server: localhost:2121"
else
    echo "  - FTP server: localhost:2121 (not available)"
fi

echo ""
echo "Test environment setup complete!"
echo ""
echo "To stop test environment:"
echo "  podman-compose -f docker-compose.test-infra.yml down"
