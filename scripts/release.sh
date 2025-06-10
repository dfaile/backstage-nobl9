#!/bin/bash
set -e

# Configuration
VERSION_FILE="VERSION"
CHANGELOG_FILE="CHANGELOG.md"
RELEASE_BRANCH="main"
DEVELOPMENT_BRANCH="develop"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Helper functions
log() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
    exit 1
}

# Check if we're on the release branch
current_branch=$(git rev-parse --abbrev-ref HEAD)
if [ "$current_branch" != "$RELEASE_BRANCH" ]; then
    error "Must be on $RELEASE_BRANCH branch to create a release"
fi

# Check for uncommitted changes
if ! git diff-index --quiet HEAD --; then
    error "Working directory has uncommitted changes"
fi

# Read current version
if [ ! -f "$VERSION_FILE" ]; then
    error "Version file not found"
fi
current_version=$(cat "$VERSION_FILE")
log "Current version: $current_version"

# Prompt for new version
read -p "Enter new version (e.g., 1.0.0): " new_version
if [ -z "$new_version" ]; then
    error "Version cannot be empty"
fi

# Validate version format (simple check)
if ! [[ $new_version =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    error "Invalid version format. Use semantic versioning (e.g., 1.0.0)"
fi

# Update version file
echo "$new_version" > "$VERSION_FILE"
log "Updated version to $new_version"

# Update changelog
if [ ! -f "$CHANGELOG_FILE" ]; then
    echo "# Changelog" > "$CHANGELOG_FILE"
fi

# Get commit messages since last release
last_tag=$(git describe --tags --abbrev=0 2>/dev/null || echo "initial")
commits=$(git log "$last_tag..HEAD" --pretty=format:"- %s")

# Add new version to changelog
{
    echo -e "\n## [$new_version] - $(date +%Y-%m-%d)\n"
    echo "### Changes"
    echo "$commits"
    echo
} | cat - "$CHANGELOG_FILE" > temp && mv temp "$CHANGELOG_FILE"

# Commit changes
git add "$VERSION_FILE" "$CHANGELOG_FILE"
git commit -m "Release version $new_version"
log "Committed version update"

# Create tag
git tag -a "v$new_version" -m "Release version $new_version"
log "Created tag v$new_version"

# Build release
log "Building release..."
go build -v -o nobl9-bot ./cmd/bot

# Create release archive
archive_name="nobl9-bot-v$new_version.tar.gz"
tar -czf "$archive_name" nobl9-bot
log "Created release archive: $archive_name"

# Push changes
read -p "Push changes to remote? (y/n): " push_confirm
if [ "$push_confirm" = "y" ]; then
    git push origin "$RELEASE_BRANCH"
    git push origin "v$new_version"
    log "Pushed changes to remote"
fi

# Merge back to development branch
read -p "Merge back to $DEVELOPMENT_BRANCH? (y/n): " merge_confirm
if [ "$merge_confirm" = "y" ]; then
    git checkout "$DEVELOPMENT_BRANCH"
    git merge "$RELEASE_BRANCH"
    git push origin "$DEVELOPMENT_BRANCH"
    git checkout "$RELEASE_BRANCH"
    log "Merged changes back to $DEVELOPMENT_BRANCH"
fi

log "Release $new_version completed successfully" 