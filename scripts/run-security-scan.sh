#!/bin/bash
# Run comprehensive security scanning for Catalogizer

set -e

echo "🔐 Starting Catalogizer Security Scanning Suite"
echo "=============================================="

# Load environment variables
if [ -f .env.security ]; then
    source .env.security
fi

# Start SonarQube services
echo "🚀 Starting SonarQube..."
podman-compose -f docker-compose.security.yml up -d sonarqube sonarqube-db

echo "⏳ Waiting for SonarQube to be ready..."
sleep 30

# Check if SonarQube is running
if curl -s http://localhost:9000/api/system/status | grep -q "UP"; then
    echo "✅ SonarQube is running at http://localhost:9000"
    echo "   Default credentials: admin/admin"
    echo "   Please log in and generate a project token, then update SONAR_TOKEN in .env.security"
else
    echo "⚠️ SonarQube may still be starting up. Check logs with:"
    echo "   podman-compose -f docker-compose.security.yml logs sonarqube"
fi

# Run Snyk scanning (if token is set)
if [ "$SNYK_TOKEN" != "dummy-token-for-freemium-mode" ] && [ -n "$SNYK_TOKEN" ]; then
    echo "🔍 Running Snyk scanning..."
    podman-compose -f docker-compose.security.yml --profile snyk-scan run --rm snyk-cli
    echo "✅ Snyk scanning completed. Reports saved to reports/"
else
    echo "⚠️ Snyk token not configured. Running in freemium mode..."
    echo "   Get a free token at: https://app.snyk.io/account"
    echo "   Then update SNYK_TOKEN in .env.security"
fi

# Run OWASP Dependency Check
echo "🔍 Running OWASP Dependency Check..."
podman-compose -f docker-compose.security.yml --profile dependency-check run --rm dependency-check
echo "✅ OWASP Dependency Check completed. Reports saved to reports/dependency-check/"

# Run Trivy scanning
echo "🔍 Running Trivy vulnerability scanning..."
podman-compose -f docker-compose.security.yml --profile trivy-scan run --rm trivy-scanner
echo "✅ Trivy scanning completed. Reports saved to reports/"

echo ""
echo "📊 Security Scanning Summary"
echo "==========================="
echo "1. SonarQube: http://localhost:9000 (admin/admin)"
echo "2. Snyk Reports: reports/snyk-*.json"
echo "3. OWASP Dependency Check: reports/dependency-check/"
echo "4. Trivy Reports: reports/trivy-results.json"
echo ""
echo "🔧 Next Steps:"
echo "   - Log into SonarQube and generate a project token"
echo "   - Update SONAR_TOKEN in .env.security"
echo "   - Get a free Snyk token from https://app.snyk.io/account"
echo "   - Update SNYK_TOKEN in .env.security"
echo "   - Review security reports and fix critical issues"
