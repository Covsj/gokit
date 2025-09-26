#!/usr/bin/env bash
set -euo pipefail

version_input=${1:-}
if [[ -z "${version_input}" ]]; then
  echo "Usage: $0 <version>  e.g. $0 v1.2.3 or $0 1.2.3" >&2
  exit 1
fi

# Normalize version to vMAJOR.MINOR.PATCH
if [[ ! "${version_input}" =~ ^v?[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "Invalid version: '${version_input}'. Use vMAJOR.MINOR.PATCH (e.g. v1.2.3)" >&2
  exit 1
fi
version=${version_input}
if [[ ! ${version} =~ ^v ]]; then
  version="v${version}"
fi

# Ensure inside a git repo
if ! git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  echo "Not a git repository." >&2
  exit 1
fi

# Ensure tag doesn't already exist
if git rev-parse -q --verify "refs/tags/${version}" >/dev/null; then
  echo "Tag ${version} already exists." >&2
  exit 1
fi

# Commit changes if any
if ! git diff --quiet HEAD --; then
  git add -A
  git commit -m "chore: release ${version}"
fi

# Create annotated tag
git tag -a "${version}" -m "release ${version}"

# Determine default remote (fallback to origin)
remote="origin"
if git remote show >/dev/null 2>&1; then
  default_remote=$(git remote | head -n1 || true)
  if [[ -n "${default_remote}" ]]; then
    remote="${default_remote}"
  fi
fi

# Push branch and tag
git push "${remote}" HEAD
git push "${remote}" "${version}"

echo "Released ${version} successfully."
