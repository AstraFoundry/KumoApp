#!/usr/bin/env bash
# Confirm this checkout is the annotated tag named by mygo.config.ts.
#
# The proof is read-only. git ls-remote prints refs/tags/<tag> as the tag
# object. Asking for that name alone does not print the peeled commit;
# refs/tags/<tag>^{} does, and only for an annotated tag. Nothing is fetched
# and the shallow file of a depth-1 Actions checkout is not touched.
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

if ! listing=$(git ls-remote origin "refs/tags/${tag}" "refs/tags/${tag}^{}"); then
  status=$?
  echo "git ls-remote failed for ${tag} (exit ${status})" >&2
  exit "$status"
fi
pointer=$(printf '%s\n' "$listing" | awk -v want="refs/tags/${tag}" '$2 == want { print $1 }')
peeled=$(printf '%s\n' "$listing" | awk -v want="refs/tags/${tag}^{}" '$2 == want { print $1 }')
if [ -z "$pointer" ]; then
  echo "origin has no tag ${tag}" >&2
  printf 'ls-remote output:\n%s\n' "$listing" >&2
  exit 1
fi
if [ -z "$peeled" ] || [ "$pointer" = "$peeled" ]; then
  echo "tag ${tag} on origin is not annotated" >&2
  printf 'ls-remote output:\n%s\n' "$listing" >&2
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
