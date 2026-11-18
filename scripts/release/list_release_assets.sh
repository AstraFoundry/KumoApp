#!/usr/bin/env bash
# Print the installer files for the current build, one path per line.
#
# This is the set mygo build -upload would attach, without update manifests
# or deltas. The updater stays off, so those files are refused. Directories
# such as Kumo.app are not listed; the disk image contains the app.
set -euo pipefail

cd "${GITHUB_WORKSPACE:-$(git rev-parse --show-toplevel)}"
build=${BUILD_DIR:-build}
if [ ! -d "$build" ]; then
  echo "no build directory ${build}" >&2
  exit 1
fi

name=$(sed -n 's/^  name: "\(.*\)",/\1/p' mygo.config.ts)
version=$(sed -n 's/^  version: "\(.*\)",/\1/p' mygo.config.ts)
if [ -z "$name" ] || [ -z "$version" ]; then
  echo "mygo.config.ts needs name and version" >&2
  exit 1
fi
slug=$(printf '%s' "$name" | tr '[:upper:]' '[:lower:]' | sed -E 's/[^a-z0-9]+/-/g; s/^-+//; s/-+$//')
if [ -z "$slug" ]; then
  echo "could not slug ${name}" >&2
  exit 1
fi

if find "$build" -type f \( -name 'update-*.json' -o -name '*.delta' \) | grep -q .; then
  echo "refusing update manifests or deltas under ${build}; updates are not part of this release" >&2
  exit 1
fi

files=()
require() {
  if [ ! -f "$1" ]; then
    echo "missing release asset $1" >&2
    exit 1
  fi
  files+=("$1")
}

shopt -s nullglob
linux=( "$build"/linux-* )
darwin=( "$build"/darwin-* )
shopt -u nullglob
if [ ${#linux[@]} -eq 0 ] && [ ${#darwin[@]} -eq 0 ]; then
  echo "no linux or darwin build under ${build}" >&2
  exit 1
fi

install_script=
for dir in "${linux[@]}"; do
  arch=${dir##*-}
  require "${dir}/${slug}_${version}_${arch}.deb"
  require "${dir}/${slug}-${version}-linux-${arch}.tar.gz"
  if [ ! -f "${dir}/install.sh" ]; then
    echo "missing ${dir}/install.sh" >&2
    exit 1
  fi
  if [ -z "$install_script" ]; then
    install_script=${dir}/install.sh
  elif ! cmp -s "$install_script" "${dir}/install.sh"; then
    echo "install.sh differs between Linux builds" >&2
    exit 1
  fi
done
if [ -n "$install_script" ]; then
  files+=("$install_script")
fi

for dir in "${darwin[@]}"; do
  arch=${dir##*-}
  if [ "$arch" = universal ]; then
    require "${dir}/${name} ${version}.dmg"
  else
    require "${dir}/${name} ${version} ${arch}.dmg"
  fi
done

printf '%s\n' "${files[@]}"
