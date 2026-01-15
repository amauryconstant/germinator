#!/usr/bin/env bash

set -e

errors=0

echo "Validating release prerequisites..."

# Check 1: Git state must be clean
echo ""
echo "Checking git state..."
if [ -n "$(git status --porcelain)" ]; then
  echo "ERROR: Git working directory has uncommitted changes:"
  git status --short
  errors=$((errors + 1))
else
  echo "✓ Git working directory is clean"
fi

# Check 2: Must be on main branch
echo ""
echo "Checking branch..."
CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
if [ "$CURRENT_BRANCH" != "main" ]; then
  echo "ERROR: Not on main branch (current: $CURRENT_BRANCH)"
  errors=$((errors + 1))
else
  echo "✓ On main branch"
fi

# Check 3: Git tag validation
echo ""
echo "Checking Git tag..."
if [ -z "$CI_COMMIT_TAG" ]; then
  GIT_TAG=$(git describe --tags --exact-match 2>/dev/null || echo "")
else
  GIT_TAG="$CI_COMMIT_TAG"
fi

if [ -z "$GIT_TAG" ]; then
  echo "ERROR: No Git tag found"
  echo "  Create and push a tag to trigger release:"
  echo "  git tag vX.Y.Z && git push origin vX.Y.Z"
  errors=$((errors + 1))
else
  echo "✓ Found Git tag: $GIT_TAG"
  
  # Extract version from tag (strip 'v' prefix)
  TAG_VERSION="${GIT_TAG#v}"
  
  if [ "$TAG_VERSION" = "$GIT_TAG" ]; then
    echo "ERROR: Git tag must start with 'v' (format: vX.Y.Z)"
    errors=$((errors + 1))
  else
    echo "  Tag version: $TAG_VERSION"
  fi
fi

# Check 4: Compare tag with version.go
echo ""
echo "Checking version.go..."
if [ ! -f "internal/version/version.go" ]; then
  echo "ERROR: internal/version/version.go not found"
  errors=$((errors + 1))
else
  CODE_VERSION=$(grep 'Version = "' internal/version/version.go | sed 's/.*Version = "\(.*\)".*/\1/' | xargs)
  
  if [ -z "$CODE_VERSION" ]; then
    echo "ERROR: Could not extract version from internal/version/version.go"
    errors=$((errors + 1))
  else
    echo "  Code version: $CODE_VERSION"
    
    if [ "$TAG_VERSION" != "$CODE_VERSION" ]; then
      echo "ERROR: Tag version does not match code version"
      echo "  Tag version: $TAG_VERSION"
      echo "  Code version: $CODE_VERSION"
      echo "  To fix: create matching tag: git tag v${CODE_VERSION}"
      errors=$((errors + 1))
    else
      echo "✓ Tag version matches code version"
    fi
  fi
fi

# Check 5: Validate GoReleaser config
echo ""
echo "Checking GoReleaser configuration..."
if [ ! -f ".goreleaser.yml" ]; then
  echo "ERROR: .goreleaser.yml not found"
  errors=$((errors + 1))
else
  if command -v goreleaser &> /dev/null; then
    if goreleaser check &> /dev/null; then
      echo "✓ GoReleaser configuration is valid"
    else
      echo "ERROR: GoReleaser configuration is invalid"
      goreleaser check
      errors=$((errors + 1))
    fi
  else
    echo "WARNING: goreleaser not found, skipping configuration check"
  fi
fi

echo ""
if [ $errors -gt 0 ]; then
  echo "❌ Validation failed with $errors error(s)"
  echo ""
  echo "Fix all errors before proceeding with release."
  exit 1
else
  echo "✅ All validation checks passed"
  echo ""
  echo "Release is ready to proceed."
  exit 0
fi
