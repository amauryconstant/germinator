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
CURRENT_VERSION=$(sed -n 's/.*const Version = "\(.*\)".*/\1/p' internal/version/version.go)

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

# Update internal/version/version.go
sed -i "s/const Version = \".*\"/const Version = \"${NEW_VERSION}\"/" internal/version/version.go
echo "Updated internal/version/version.go to ${NEW_VERSION}"

echo ""
echo "Next steps:"
echo "  1. Review changes: git status"
echo "  2. Commit changes: git commit -am 'Bump version to ${NEW_VERSION}'"
echo "  3. Push changes: git push origin"
