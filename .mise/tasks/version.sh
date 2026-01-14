#!/usr/bin/env bash

set -e

PART=$1

if [ -z "$PART" ]; then
  echo "Usage: $0 <patch|minor|major>"
  exit 1
fi

if [ ! -f "internal/version/version.go" ]; then
  echo "Error: internal/version/version.go not found in current directory"
  exit 1
fi

# Get current version from internal/version/version.go
CURRENT_VERSION=$(grep 'Version = "' internal/version/version.go | sed 's/.*Version = "\(.*\)".*/\1/')
CURRENT_VERSION=$(echo "$CURRENT_VERSION" | xargs)

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

echo "Bumping version: ${CURRENT_VERSION} -> ${NEW_VERSION}"

# Get git commit SHA and date
COMMIT_SHA=$(git rev-parse HEAD)
COMMIT_DATE=$(date +%Y-%m-%d)

# Update all version variables in internal/version/version.go
sed -i "s/Version = \".*\"/Version = \"${NEW_VERSION}\"/" internal/version/version.go
sed -i "s/Commit  = \".*\"/Commit  = \"${COMMIT_SHA}\"/" internal/version/version.go
sed -i "s/Date    = \".*\"/Date    = \"${COMMIT_DATE}\"/" internal/version/version.go

echo "Updated internal/version/version.go:"
echo "  Version: ${NEW_VERSION}"
echo "  Commit:  ${COMMIT_SHA}"
echo "  Date:    ${COMMIT_DATE}"

echo ""
echo "Next steps:"
echo "  1. Review changes: git status"
echo "  2. Commit changes: git commit -am 'Bump version to ${NEW_VERSION}'"
echo "  3. Create and push tag: git tag v${NEW_VERSION} && git push origin v${NEW_VERSION}"
echo ""
echo "Note: Commit and Date will be automatically updated during GoReleaser release"
