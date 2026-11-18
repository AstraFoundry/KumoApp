#!/usr/bin/env bash
# Print the installer files for the current build, one path per line.
#
# This is the set mygo build -upload would attach, without update manifests
# or deltas. The updater stays off, so those files are refused. Directories
# such as Kumo.app are not listed; the disk image contains the app.
#
# Bash 3.2, the macOS runner's /bin/bash, treats an empty array expansion as
# an unbound variable under set -u. This script does not use arrays. A
# platform the build did not produce simply does not enter its loop.
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

saw=0
install_script=
shopt -s nullglob
for dir in "$build"/linux-*; do
  [ -d "$dir" ] || continue
  saw=1
  arch=${dir##*-}
  deb="${dir}/${slug}_${version}_${arch}.deb"
  archive="${dir}/${slug}-${version}-linux-${arch}.tar.gz"
  [ -f "$deb" ] || { echo "missing release asset ${deb}" >&2; exit 1; }
  [ -f "$archive" ] || { echo "missing release asset ${archive}" >&2; exit 1; }
  [ -f "${dir}/install.sh" ] || { echo "missing ${dir}/install.sh" >&2; exit 1; }
  printf '%s\n' "$deb" "$archive"
  if [ -z "$install_script" ]; then
    install_script=${dir}/install.sh
  elif ! cmp -s "$install_script" "${dir}/install.sh"; then
    echo "install.sh differs between Linux builds" >&2
    exit 1
  fi
done
if [ -n "$install_script" ]; then
  printf '%s\n' "$install_script"
fi
for dir in "$build"/darwin-*; do
  [ -d "$dir" ] || continue
  saw=1
  arch=${dir##*-}
  if [ "$arch" = universal ]; then
    dmg="${dir}/${name} ${version}.dmg"
  else
    dmg="${dir}/${name} ${version} ${arch}.dmg"
  fi
  [ -f "$dmg" ] || { echo "missing release asset ${dmg}" >&2; exit 1; }
  printf '%s\n' "$dmg"
done
shopt -u nullglob
if [ "$saw" -eq 0 ]; then
  echo "no linux or darwin build under ${build}" >&2
  exit 1
fi
