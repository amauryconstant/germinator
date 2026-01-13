#!/usr/bin/env bash

set -e

PART=$1

if [ -z "$PART" ]; then
  echo "Usage: $0 <patch|minor|major>"
  exit 1
fi

if [ ! -f "go.mod" ]; then
  echo "Error: go.mod not found in current directory"
  exit 1
fi

# Get the latest git tag as current version
CURRENT_VERSION=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")

# Remove 'v' prefix if present
CURRENT_VERSION=${CURRENT_VERSION#v}

# Parse semver
IFS='.' read -r MAJOR MINOR PATCH <<< "$CURRENT_VERSION"

# Validate semver format
if ! [[ "$MAJOR" =~ ^[0-9]+$ ]] || ! [[ "$MINOR" =~ ^[0-9]+$ ]] || ! [[ "$PATCH" =~ ^[0-9]+$ ]]; then
  echo "Error: Invalid version format: $CURRENT_VERSION"
  exit 1
fi

# Bump version based on part
case $PART in
  patch)
    NEW_PATCH=$((PATCH + 1))
    NEW_VERSION="${MAJOR}.${MINOR}.${NEW_PATCH}"
    ;;
  minor)
    NEW_MINOR=$((MINOR + 1))
    NEW_VERSION="${MAJOR}.${NEW_MINOR}.0"
    ;;
  major)
    NEW_MAJOR=$((MAJOR + 1))
    NEW_VERSION="${NEW_MAJOR}.0.0"
    ;;
  *)
    echo "Error: Invalid part '$PART'. Use patch, minor, or major"
    exit 1
    ;;
esac

# Create git tag
NEW_TAG="v${NEW_VERSION}"

echo "Bumping version: ${CURRENT_VERSION} -> ${NEW_VERSION}"
git tag "$NEW_TAG"
echo "Created git tag: $NEW_TAG"

echo ""
echo "Next steps:"
echo "  1. Review changes: git status"
echo "  2. Commit changes: git commit -am 'Bump version to ${NEW_VERSION}'"
echo "  3. Push tag: git push origin $NEW_TAG"
