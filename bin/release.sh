#!/bin/bash
set -euo pipefail

CHART_FILE="charts/k8s-jwks-proxy/Chart.yaml"
VALUES_FILE="charts/k8s-jwks-proxy/values.yaml"
VERSION_TYPE=${1:-patch}  # default to patch if not provided

if [[ ! -f "$CHART_FILE" ]]; then
  echo "❌ Chart.yaml not found at $CHART_FILE"
  exit 1
fi

if [[ ! -f "$VALUES_FILE" ]]; then
  echo "❌ values.yaml not found at $VALUES_FILE"
  exit 1
fi

# Read current version
current_version=$(grep '^version:' "$CHART_FILE" | awk '{print $2}')
echo "🔢 Current version: $current_version"

# Split version into components
IFS='.' read -r major minor patch <<< "$current_version"

# Bump version
case "$VERSION_TYPE" in
  major)
    major=$((major + 1))
    minor=0
    patch=0
    ;;
  minor)
    minor=$((minor + 1))
    patch=0
    ;;
  patch)
    patch=$((patch + 1))
    ;;
  *)
    echo "❌ Unknown version type: $VERSION_TYPE"
    exit 1
    ;;
esac

new_version="${major}.${minor}.${patch}"
echo "➡️  New version: $new_version"

# Update Chart.yaml
sed -i'' -e "s/^version: .*/version: $new_version/" "$CHART_FILE"
sed -i'' -e "s/^appVersion: .*/appVersion: \"$new_version\"/" "$CHART_FILE"

# Update image.tag in values.yaml
sed -i'' -e "s/^\(\s*tag:\s*\).*/\1\"$new_version\"/" "$VALUES_FILE"

# Commit and tag
git add .
git commit -m "release: bump chart version to v$new_version"
git tag "v$new_version"

# Push to GitHub
git push origin main
git push origin "v$new_version"

echo "✅ Released version v$new_version"
