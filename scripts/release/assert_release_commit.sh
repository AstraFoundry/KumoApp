#!/usr/bin/env bash
# Confirm this checkout is the annotated tag named by mygo.config.ts.
#
# RELEASE_REF is that tag (vX.Y.Z or refs/tags/vX.Y.Z). The tag is fetched
# from origin into a temporary ref outside refs/tags. An exact
# `git ls-remote --tags origin refs/tags/<tag>` returns only the tag object
# for an annotated tag, not the peeled commit, so the commit is read with
# git rev-parse on the fetched tag object. The published tag is not created,
# moved, or deleted. A lightweight tag is rejected.
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

check=refs/kumo-release-check/${tag}
cleanup() {
  git update-ref -d "$check" >/dev/null 2>&1 || true
}
trap cleanup EXIT

if ! git fetch --quiet --no-tags origin "refs/tags/${tag}:${check}"; then
  echo "origin has no tag ${tag}" >&2
  exit 1
fi
kind=$(git cat-file -t "$check")
if [ "$kind" != tag ]; then
  echo "tag ${tag} on origin is ${kind}, not an annotated tag" >&2
  exit 1
fi
pointer=$(git rev-parse "$check")
peeled=$(git rev-parse "${check}^{commit}")
if [ -z "$peeled" ] || [ "$pointer" = "$peeled" ]; then
  echo "annotated tag ${tag} did not peel to a commit" >&2
  exit 1
fi

head=$(git rev-parse HEAD)
if [ "$head" != "$peeled" ]; then
  echo "HEAD ${head} is not ${tag} (${peeled}); refusing to treat another commit as that release" >&2
  exit 1
fi

echo "source_tag=${tag}"
echo "source_tag_object=${pointer}"
echo "source_sha=${head}"
if [ -n "${GITHUB_OUTPUT:-}" ]; then
  printf 'tag=%s\nsha=%s\n' "$tag" "$head" >> "$GITHUB_OUTPUT"
fi
