#!/bin/bash
###############################################################################
# SECURITY TOOLS INSTALLATION GUIDE
# Catalogizer Project - Security Tool Setup
###############################################################################

# This script installs all required security scanning tools
# Run with: ./install-security-tools.sh

set -e

INSTALL_DIR="${HOME}/bin"
mkdir -p "$INSTALL_DIR"

echo "=== Installing Security Tools ==="
echo "Install directory: $INSTALL_DIR"
echo ""

# Add to PATH if not already present
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo "Adding $INSTALL_DIR to PATH"
    echo 'export PATH="$HOME/bin:$PATH"' >> ~/.bashrc
    export PATH="$INSTALL_DIR:$PATH"
fi

# 1. Trivy - Container and filesystem vulnerability scanner
echo "[1/5] Installing Trivy..."
if ! command -v trivy &> /dev/null; then
    curl -sfL https://raw.githubusercontent.com/aquasecurity/trivy/main/contrib/install.sh | sh -s -- -b "$INSTALL_DIR"
    echo "✓ Trivy installed"
else
    echo "✓ Trivy already installed ($(trivy version))"
fi

# 2. Gosec - Go security checker
echo "[2/5] Installing Gosec..."
if ! command -v gosec &> /dev/null; then
    go install github.com/securego/gosec/v2/cmd/gosec@latest
    echo "✓ Gosec installed"
else
    echo "✓ Gosec already installed ($(gosec -version 2>&1 | head -1))"
fi

# 3. Nancy - Go dependency vulnerability scanner
echo "[3/5] Installing Nancy..."
if ! command -v nancy &> /dev/null; then
    go install github.com/sonatypecommunity/nancy@latest
    echo "✓ Nancy installed"
else
    echo "✓ Nancy already installed"
fi

# 4. Semgrep - Static analysis security testing
echo "[4/5] Installing Semgrep..."
if ! command -v semgrep &> /dev/null; then
    pip3 install semgrep
    echo "✓ Semgrep installed"
else
    echo "✓ Semgrep already installed ($(semgrep --version))"
fi

# 5. GitLeaks - Secret detection
echo "[5/5] Installing GitLeaks..."
if ! command -v gitleaks &> /dev/null; then
    go install github.com/zricethezav/gitleaks/v8@latest
    echo "✓ GitLeaks installed"
else
    echo "✓ GitLeaks already installed ($(gitleaks version))"
fi

echo ""
echo "=== Installation Complete ==="
echo ""
echo "Installed tools:"
command -v trivy &> /dev/null && echo "  ✓ Trivy: $(trivy version 2>/dev/null | head -1)"
command -v gosec &> /dev/null && echo "  ✓ Gosec: $(gosec -version 2>&1 | head -1)"
command -v nancy &> /dev/null && echo "  ✓ Nancy: installed"
command -v semgrep &> /dev/null && echo "  ✓ Semgrep: $(semgrep --version 2>/dev/null)"
command -v gitleaks &> /dev/null && echo "  ✓ GitLeaks: $(gitleaks version 2>/dev/null)"
echo ""
echo "Make sure $INSTALL_DIR is in your PATH"
echo "Run: export PATH=\"$HOME/bin:\$PATH\""
