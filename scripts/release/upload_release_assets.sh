#!/usr/bin/env bash
# Attach this build's installers to the existing draft of its version tag.
#
# Does not create a release, publish it, or create or move a tag. The
# checkout must already be the annotated tag (see assert_release_commit.sh).
# DRY_RUN=1 prints the files and does not call gh.
set -euo pipefail

here=$(cd "$(dirname "$0")" && pwd)
# shellcheck disable=SC1091
RELEASE_REF=${RELEASE_REF:?set RELEASE_REF to the tag being built} \
  bash "$here/assert_release_commit.sh" >/dev/null

cd "${GITHUB_WORKSPACE:-$(git rev-parse --show-toplevel)}"
version=$(sed -n 's/^  version: "\(.*\)",/\1/p' mygo.config.ts)
tag="v${version}"

# Positional parameters, not an array. Bash 3.2 rejects an empty array
# expansion under set -u. "$@" with no parameters is empty and is safe.
if ! listed=$("$here/list_release_assets.sh"); then
  exit $?
fi
set --
if [ -n "$listed" ]; then
  while IFS= read -r line; do
    [ -n "$line" ] && set -- "$@" "$line"
  done <<EOF
$listed
EOF
fi
if [ "$#" -eq 0 ]; then
  echo "no release assets" >&2
  exit 1
fi

echo "release ${tag} assets from $(git rev-parse HEAD):"
printf '  %s\n' "$@"

if [ "${DRY_RUN:-}" = 1 ]; then
  echo "dry run: not calling gh"
  exit 0
fi

repo=${GITHUB_REPOSITORY:?GITHUB_REPOSITORY is required}
draft=$(gh release view "$tag" --repo "$repo" --json isDraft --jq .isDraft)
if [ "$draft" != "true" ]; then
  echo "release ${tag} is not an unpublished draft; refusing to upload" >&2
  exit 1
fi

gh release upload "$tag" --repo "$repo" --clobber "$@"
echo "uploaded $# files to draft ${tag}; left it unpublished"
