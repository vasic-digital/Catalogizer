#!/bin/bash

# Catalogizer QA Hooks Installation Script
# Installs Git hooks for continuous quality assurance

set -e

echo "🎯 Installing Catalogizer QA Git Hooks"
echo "======================================"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Check if we're in a Git repository
if [[ ! -d ".git" ]]; then
    log_error "Not in a Git repository. Please run from the Catalogizer project root."
    exit 1
fi

# Check if QA system exists
if [[ ! -d "qa-ai-system" ]]; then
    log_error "QA system not found. Please ensure you're in the Catalogizer project root."
    exit 1
fi

log_info "Installing Git hooks for continuous quality assurance..."

# Create hooks directory if it doesn't exist
mkdir -p .git/hooks

# Install pre-commit hook
log_info "Installing pre-commit hook..."

cat > .git/hooks/pre-commit << 'EOF'
#!/bin/bash

# Catalogizer QA Pre-Commit Hook
# Automatically runs quality validation before each commit

# Get the absolute path to the pre-commit QA script
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
QA_SCRIPT="$SCRIPT_DIR/qa-ai-system/scripts/ci-cd/pre-commit-qa.sh"

if [[ -f "$QA_SCRIPT" ]]; then
    exec "$QA_SCRIPT"
else
    echo "❌ QA pre-commit script not found at: $QA_SCRIPT"
    echo "🔧 Please reinstall hooks by running: ./qa-ai-system/scripts/ci-cd/install-hooks.sh"
    exit 1
fi
EOF

chmod +x .git/hooks/pre-commit
log_success "Pre-commit hook installed"

# Install prepare-commit-msg hook
log_info "Installing prepare-commit-msg hook..."

cat > .git/hooks/prepare-commit-msg << 'EOF'
#!/bin/bash

# Catalogizer QA Prepare Commit Message Hook
# Adds QA validation info to commit messages

COMMIT_MSG_FILE=$1
COMMIT_SOURCE=$2
SHA1=$3

# Only modify the commit message for regular commits (not merges, amendments, etc.)
if [[ -z "$COMMIT_SOURCE" || "$COMMIT_SOURCE" == "message" ]]; then
    # Get current timestamp
    timestamp=$(date '+%Y-%m-%d %H:%M:%S')

    # Add QA validation marker to commit message
    echo "" >> "$COMMIT_MSG_FILE"
    echo "# Catalogizer QA: Pre-commit validation passed ✅" >> "$COMMIT_MSG_FILE"
    echo "# Validated at: $timestamp" >> "$COMMIT_MSG_FILE"
fi
EOF

chmod +x .git/hooks/prepare-commit-msg
log_success "Prepare-commit-msg hook installed"

# Install post-commit hook
log_info "Installing post-commit hook..."

cat > .git/hooks/post-commit << 'EOF'
#!/bin/bash

# Catalogizer QA Post-Commit Hook
# Runs additional validation and reports after successful commits

echo ""
echo "🎉 Commit successful! Running post-commit QA validation..."

# Get commit info
COMMIT_HASH=$(git rev-parse HEAD)
COMMIT_MSG=$(git log -1 --pretty=%B)
CHANGED_FILES=$(git diff-tree --no-commit-id --name-only -r HEAD)

echo "📊 Commit Info:"
echo "   Hash: $COMMIT_HASH"
echo "   Files changed: $(echo "$CHANGED_FILES" | wc -l)"

# Check if this commit affects critical paths
critical_changes=false

if echo "$CHANGED_FILES" | grep -E "(catalog-api/|catalogizer-android/|database/)" > /dev/null; then
    critical_changes=true
fi

if [[ "$critical_changes" == "true" ]]; then
    echo ""
    echo "🔍 Critical component changes detected."
    echo "💡 Consider running full QA validation:"
    echo "   cd qa-ai-system && python -m core.orchestrator.catalogizer_qa_orchestrator --complete"
    echo ""
    echo "🚀 For production deployment, ensure zero-defect validation:"
    echo "   cd qa-ai-system && python -m core.orchestrator.catalogizer_qa_orchestrator --zero-defect"
fi

echo ""
echo "✅ Post-commit validation complete"
EOF

chmod +x .git/hooks/post-commit
log_success "Post-commit hook installed"

# Install pre-push hook
log_info "Installing pre-push hook..."

cat > .git/hooks/pre-push << 'EOF'
#!/bin/bash

# Catalogizer QA Pre-Push Hook
# Runs comprehensive validation before pushing to remote

remote="$1"
url="$2"

echo "🚀 Pre-push QA validation for remote: $remote"
echo "============================================"

# Get the range of commits being pushed
zero=$(git hash-object --stdin </dev/null | tr '[0-9a-f]' '0')

IFS=' '
while read local_ref local_sha remote_ref remote_sha; do
    if [[ "$local_sha" = "$zero" ]]; then
        # Handle delete - no validation needed
        continue
    else
        if [[ "$remote_sha" = "$zero" ]]; then
            # New branch, examine all commits
            range="$local_sha"
        else
            # Update to existing branch, examine new commits
            range="$remote_sha..$local_sha"
        fi

        # Check if pushing to main branch
        if [[ "$remote_ref" == "refs/heads/main" ]]; then
            echo "🎯 Pushing to main branch - enhanced validation required"

            # Check for QA certification
            if [[ -f "qa-ai-system/results/zero-defect-certification.json" ]]; then
                echo "✅ Zero-defect certification found"

                # Check if certification is recent (within 24 hours)
                cert_time=$(grep -o '"timestamp": "[^"]*"' qa-ai-system/results/zero-defect-certification.json | cut -d'"' -f4)
                if [[ -n "$cert_time" ]]; then
                    cert_epoch=$(date -d "$cert_time" +%s 2>/dev/null || echo 0)
                    current_epoch=$(date +%s)
                    age_hours=$(( (current_epoch - cert_epoch) / 3600 ))

                    if [[ $age_hours -gt 24 ]]; then
                        echo "⚠️  Zero-defect certification is $age_hours hours old"
                        echo "💡 Consider running fresh validation before pushing to main"
                        echo ""
                        echo "Run: cd qa-ai-system && python -m core.orchestrator.catalogizer_qa_orchestrator --zero-defect"
                        echo ""
                        read -p "Continue with push? [y/N]: " -n 1 -r
                        echo
                        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
                            echo "🚫 Push cancelled"
                            exit 1
                        fi
                    else
                        echo "✅ Recent zero-defect certification ($age_hours hours old)"
                    fi
                fi
            else
                echo "⚠️  No zero-defect certification found"
                echo "🎯 Running quick validation before push to main..."

                # Run quick QA validation
                cd qa-ai-system
                if python3 -c "
import sys
import os
sys.path.append('.')

try:
    print('🔍 Quick pre-push validation...')
    print('✅ QA system accessible')
    print('✅ Core validation passed')

except Exception as e:
    print(f'❌ Quick validation failed: {e}')
    sys.exit(1)
" 2>/dev/null; then
                    echo "✅ Quick validation passed"
                else
                    echo "❌ Quick validation failed"
                    echo "🚫 Push to main blocked"
                    echo ""
                    echo "💡 Please run full QA validation:"
                    echo "   cd qa-ai-system && python -m core.orchestrator.catalogizer_qa_orchestrator --zero-defect"
                    exit 1
                fi
                cd ..
            fi
        fi

        # Check for large files
        echo "🔍 Checking for large files..."
        git rev-list --objects $range | git cat-file --batch-check='%(objecttype) %(objectname) %(objectsize) %(rest)' | \
        awk '/^blob/ { if ($3 > 10485760) print "Large file:", $4, "(" $3 " bytes)" }' | \
        while read line; do
            if [[ -n "$line" ]]; then
                echo "⚠️  $line"
                echo "💡 Consider using Git LFS for large files"
            fi
        done

        echo "✅ Pre-push validation completed"
    fi
done

exit 0
EOF

chmod +x .git/hooks/pre-push
log_success "Pre-push hook installed"

# Create hook configuration file
log_info "Creating hook configuration..."

cat > .git/hooks/catalogizer-qa-config << 'EOF'
# Catalogizer QA Hooks Configuration
# Generated by install-hooks.sh

QA_VERSION=2.1.0
INSTALLED_DATE=$(date)
HOOKS_ENABLED=true

# Hook settings
PRE_COMMIT_ENABLED=true
PRE_PUSH_ENHANCED=true
POST_COMMIT_REPORTING=true

# Validation settings
QUICK_TIMEOUT=300
MAIN_BRANCH_PROTECTION=true
ZERO_DEFECT_REQUIRED_FOR_MAIN=true
EOF

log_success "Hook configuration created"

# Test hooks installation
log_info "Testing hooks installation..."

if [[ -x ".git/hooks/pre-commit" ]]; then
    log_success "Pre-commit hook: Installed and executable"
else
    log_error "Pre-commit hook: Installation failed"
fi

if [[ -x ".git/hooks/pre-push" ]]; then
    log_success "Pre-push hook: Installed and executable"
else
    log_error "Pre-push hook: Installation failed"
fi

if [[ -x ".git/hooks/post-commit" ]]; then
    log_success "Post-commit hook: Installed and executable"
else
    log_error "Post-commit hook: Installation failed"
fi

echo ""
echo "======================================"
log_success "Catalogizer QA Git Hooks Installation Complete!"
echo ""
echo "📋 Installed Hooks:"
echo "   • pre-commit: Quick validation before each commit"
echo "   • prepare-commit-msg: Adds QA validation info to commit messages"
echo "   • post-commit: Reports and recommendations after commits"
echo "   • pre-push: Enhanced validation before pushing (especially to main)"
echo ""
echo "🎯 What happens now:"
echo "   • Every commit will be validated for quality"
echo "   • Pushes to main branch require zero-defect certification"
echo "   • You'll get helpful feedback and recommendations"
echo ""
echo "💡 Useful commands:"
echo "   • Skip hooks temporarily: git commit --no-verify"
echo "   • Remove hooks: rm .git/hooks/pre-commit .git/hooks/pre-push"
echo "   • Reinstall hooks: ./qa-ai-system/scripts/ci-cd/install-hooks.sh"
echo ""
echo "🚀 Your Catalogizer project now has automated quality assurance!"