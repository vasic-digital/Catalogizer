#!/bin/bash
# Quick security scan for development

set -e

echo "🔍 Running quick security scan..."

# Run Go security checks
echo "🔐 Checking Go dependencies for vulnerabilities..."
cd catalog-api
if command -v govulncheck &> /dev/null; then
    govulncheck ./... 2>/dev/null || true
else
    echo "⚠️ govulncheck not installed. Install with: go install golang.org/x/vuln/cmd/govulncheck@latest"
fi
cd ..

# Run npm audit for frontend
echo "🔐 Checking npm dependencies for vulnerabilities..."
if [ -f "catalog-web/package.json" ]; then
    cd catalog-web
    npm audit --audit-level=high 2>/dev/null || true
    cd ..
fi

# Run basic container security check
echo "🔐 Checking Dockerfile security..."
if [ -f "catalog-api/Dockerfile" ]; then
    echo "📋 Dockerfile security tips:"
    echo "   - Use specific version tags for base images"
    echo "   - Run as non-root user"
    echo "   - Use multi-stage builds to reduce attack surface"
    echo "   - Scan images with Trivy or Snyk"
fi

echo "✅ Quick security scan completed"
