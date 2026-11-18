#!/usr/bin/env bash
# Confirm this checkout is the annotated tag named by mygo.config.ts.
#
# RELEASE_REF is that tag (vX.Y.Z or refs/tags/vX.Y.Z). The peeled commit is
# read with git ls-remote, which does not create or move the tag. Tests may
# set RELEASE_TAG_COMMIT instead.
set -euo pipefail

cd "${GITHUB_WORKSPACE:-$(git rev-parse --show-toplevel)}"

ref=${RELEASE_REF:?set RELEASE_REF to the tag being built}
ref=${ref#refs/tags/}
version=$(sed -n 's/^  version: "\(.*\)",/\1/p' mygo.config.ts)
if [ -z "$version" ]; then
  echo "mygo.config.ts has no version" >&2
  exit 1
fi
tag="v${version}"
if [ "$ref" != "$tag" ]; then
  echo "ref ${ref} is not ${tag}, the version in this checkout" >&2
  exit 1
fi

if [ -n "${RELEASE_TAG_COMMIT:-}" ]; then
  peeled=$RELEASE_TAG_COMMIT
else
  peeled=$(git ls-remote --tags origin "refs/tags/${tag}" | awk -v want="refs/tags/${tag}^{}" '$2 == want { print $1 }')
fi
if [ -z "$peeled" ]; then
  echo "origin has no peeled commit for annotated tag ${tag}" >&2
  exit 1
fi

head=$(git rev-parse HEAD)
if [ "$head" != "$peeled" ]; then
  echo "HEAD ${head} is not ${tag} (${peeled}); refusing to treat another commit as that release" >&2
  exit 1
fi

echo "source_tag=${tag}"
echo "source_sha=${head}"
if [ -n "${GITHUB_OUTPUT:-}" ]; then
  printf 'tag=%s\nsha=%s\n' "$tag" "$head" >> "$GITHUB_OUTPUT"
fi
