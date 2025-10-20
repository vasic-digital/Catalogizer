#!/bin/bash

# Verify Freemium Security Testing Setup
# This script checks if all freemium tools are properly configured

set -e

echo "🔍 Verifying Catalogizer Freemium Security Setup"
echo "================================================"
echo ""

# Function to check SonarQube setup
check_sonarqube() {
    echo "🔍 Checking SonarQube setup..."
    if [ -n "$SONAR_TOKEN" ]; then
        echo "✅ SONAR_TOKEN is set"
        # Test token validity (basic check)
        if curl -s -H "Authorization: Bearer $SONAR_TOKEN" "https://sonarcloud.io/api/user_tokens/search" &> /dev/null; then
            echo "✅ SONAR_TOKEN appears valid"
        else
            echo "⚠️  SONAR_TOKEN may be invalid (or network issue)"
        fi
    else
        echo "❌ SONAR_TOKEN not set"
        echo "   Get your free token at: https://sonarcloud.io/account"
        return 1
    fi
    echo ""
}

# Function to check Snyk setup
check_snyk() {
    echo "🔍 Checking Snyk setup..."
    if [ -n "$SNYK_TOKEN" ]; then
        echo "✅ SNYK_TOKEN is set"
        # Test token validity
        if command -v snyk &> /dev/null; then
            if snyk test --version &> /dev/null; then
                echo "✅ Snyk CLI is installed and working"
                # Try a basic test to verify token
                if snyk whoami &> /dev/null; then
                    echo "✅ SNYK_TOKEN is authenticated"
                else
                    echo "⚠️  SNYK_TOKEN authentication failed"
                    echo "   Run: snyk auth $SNYK_TOKEN"
                fi
            else
                echo "⚠️  Snyk CLI installed but not working"
            fi
        else
            echo "⚠️  Snyk CLI not installed"
            echo "   Run: npm install -g snyk"
        fi
    else
        echo "❌ SNYK_TOKEN not set"
        echo "   Get your free token at: https://snyk.io/account"
        return 1
    fi
    echo ""
}

# Function to check Docker setup
check_docker() {
    echo "🔍 Checking Docker setup..."
    if command -v docker &> /dev/null; then
        echo "✅ Docker is installed"
        if command -v docker-compose &> /dev/null; then
            echo "✅ Docker Compose is installed"
            # Test Docker functionality
            if docker run --rm hello-world &> /dev/null; then
                echo "✅ Docker is working"
            else
                echo "⚠️  Docker installed but not working"
            fi
        else
            echo "❌ Docker Compose not installed"
            echo "   Install Docker Compose for full security testing"
        fi
    else
        echo "❌ Docker not installed"
        echo "   Docker is optional but recommended for full testing"
    fi
    echo ""
}

# Function to check security scripts
check_scripts() {
    echo "🔍 Checking security scripts..."
    SCRIPTS=(
        "scripts/security-test.sh"
        "scripts/sonarqube-scan.sh"
        "scripts/snyk-scan.sh"
        "scripts/setup-freemium-tokens.sh"
        "scripts/verify-freemium-setup.sh"
    )

    for script in "${SCRIPTS[@]}"; do
        if [ -x "$script" ]; then
            echo "✅ $script is executable"
        else
            echo "❌ $script is not executable or missing"
        fi
    done
    echo ""
}

# Function to check Docker Compose configuration
check_docker_compose() {
    echo "🔍 Checking Docker Compose configuration..."
    if [ -f "docker-compose.security.yml" ]; then
        echo "✅ docker-compose.security.yml exists"
        # Basic syntax check
        if docker-compose -f docker-compose.security.yml config &> /dev/null; then
            echo "✅ Docker Compose configuration is valid"
        else
            echo "⚠️  Docker Compose configuration has issues"
        fi
    else
        echo "❌ docker-compose.security.yml not found"
    fi
    echo ""
}

# Function to provide summary
provide_summary() {
    echo "📊 Setup Verification Summary"
    echo "=============================="

    local issues=0

    if [ -z "$SONAR_TOKEN" ]; then ((issues++)); fi
    if [ -z "$SNYK_TOKEN" ]; then ((issues++)); fi
    if ! command -v docker &> /dev/null; then ((issues++)); fi
    if ! command -v docker-compose &> /dev/null; then ((issues++)); fi

    if [ $issues -eq 0 ]; then
        echo "🎉 All freemium tools are properly configured!"
        echo ""
        echo "🚀 Ready to run security tests:"
        echo "   ./scripts/security-test.sh"
    else
        echo "⚠️  $issues configuration issues found"
        echo ""
        echo "🔧 To fix issues:"
        echo "   ./scripts/setup-freemium-tokens.sh"
    fi
    echo ""
    echo "📖 For more information: docs/TESTING_GUIDE.md"
}

# Main execution
main() {
    check_sonarqube
    check_snyk
    check_docker
    check_scripts
    check_docker_compose
    provide_summary
}

# Run main function
main "$@"