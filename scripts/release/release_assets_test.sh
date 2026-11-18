#!/usr/bin/env bash
# Hermetic checks for the release asset scripts and the workflow contract.
set -euo pipefail

root=$(cd "$(dirname "$0")/../.." && pwd)
workflow=$root/.github/workflows/build-release.yml
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

fail() { printf 'FAIL %s\n' "$*" >&2; exit 1; }
pass() { printf 'ok  %s\n' "$*"; }

# A checkout that is the annotated tag may list and describe its assets.
repo=$tmp/repo
mkdir -p "$repo/build/linux-amd64" "$repo/build/linux-arm64" \
  "$repo/build/darwin-arm64" "$repo/build/darwin-amd64"
cat > "$repo/mygo.config.ts" <<'EOF'
export default {
  name: "Kumo",
  version: "0.1.0",
}
EOF
for arch in amd64 arm64; do
  printf 'deb %s\n' "$arch" > "$repo/build/linux-${arch}/kumo_0.1.0_${arch}.deb"
  printf 'tar %s\n' "$arch" > "$repo/build/linux-${arch}/kumo-0.1.0-linux-${arch}.tar.gz"
  printf 'install\n' > "$repo/build/linux-${arch}/install.sh"
  printf 'bin\n' > "$repo/build/linux-${arch}/kumo"
  printf 'dmg %s\n' "$arch" > "$repo/build/darwin-${arch}/Kumo 0.1.0 ${arch}.dmg"
  mkdir -p "$repo/build/darwin-${arch}/Kumo.app"
done
printf 'update\n' > "$repo/build/linux-amd64/update-linux-amd64.json"
git -C "$repo" init -q
git -C "$repo" config user.email test@example.com
git -C "$repo" config user.name test
git -C "$repo" add mygo.config.ts
git -C "$repo" commit -q -m tag
tag_commit=$(git -C "$repo" rev-parse HEAD)
git -C "$repo" tag -a v0.1.0 -m v0.1.0
# A later commit must not be accepted as v0.1.0.
echo 'x' >> "$repo/mygo.config.ts"
git -C "$repo" commit -qam 'not the tag'
if GITHUB_WORKSPACE=$repo RELEASE_REF=v0.1.0 RELEASE_TAG_COMMIT=$tag_commit \
  bash "$root/scripts/release/assert_release_commit.sh" >/dev/null; then
  fail "a newer commit was accepted as v0.1.0"
fi
pass "newer commit is not v0.1.0"
git -C "$repo" checkout -q v0.1.0
# The update manifest is untracked build output, still visible after checkout.
if GITHUB_WORKSPACE=$repo BUILD_DIR="$repo/build" bash "$root/scripts/release/list_release_assets.sh" >/dev/null; then
  fail "update manifest was listed"
fi
pass "update manifests are refused"
rm "$repo/build/linux-amd64/update-linux-amd64.json"
listed=$(GITHUB_WORKSPACE=$repo BUILD_DIR="$repo/build" bash "$root/scripts/release/list_release_assets.sh")
expect=$'build/linux-amd64/kumo_0.1.0_amd64.deb
build/linux-amd64/kumo-0.1.0-linux-amd64.tar.gz
build/linux-arm64/kumo_0.1.0_arm64.deb
build/linux-arm64/kumo-0.1.0-linux-arm64.tar.gz
build/linux-amd64/install.sh
build/darwin-amd64/Kumo 0.1.0 amd64.dmg
build/darwin-arm64/Kumo 0.1.0 arm64.dmg'
# Paths are relative to the workspace. Normalize.
got=$(printf '%s\n' "$listed" | sed "s|^$repo/||")
if [ "$got" != "$expect" ]; then
  printf 'asset list:\n%s\nexpected:\n%s\n' "$got" "$expect" >&2
  fail "asset list"
fi
pass "installers listed, app bundles and binaries omitted"
if ! printf '%s\n' "$listed" | grep -q 'Kumo.app'; then
  pass "Kumo.app is not an upload"
else
  fail "Kumo.app was listed"
fi

out=$(GITHUB_WORKSPACE=$repo RELEASE_REF=refs/tags/v0.1.0 RELEASE_TAG_COMMIT=$tag_commit \
  bash "$root/scripts/release/assert_release_commit.sh")
printf '%s\n' "$out" | grep -q "source_sha=${tag_commit}" || fail "assert sha"
printf '%s\n' "$out" | grep -q "source_tag=v0.1.0" || fail "assert tag"
pass "annotated tag checkout matches"

if GITHUB_WORKSPACE=$repo RELEASE_REF=main RELEASE_TAG_COMMIT=$tag_commit \
  bash "$root/scripts/release/assert_release_commit.sh" >/dev/null; then
  fail "branch ref was accepted"
fi
pass "only the version tag is accepted"

dry=$(GITHUB_WORKSPACE=$repo RELEASE_REF=v0.1.0 RELEASE_TAG_COMMIT=$tag_commit DRY_RUN=1 \
  bash "$root/scripts/release/upload_release_assets.sh")
printf '%s\n' "$dry" | grep -q 'dry run: not calling gh' || fail "dry run"
printf '%s\n' "$dry" | grep -q 'Kumo 0.1.0 arm64.dmg' || fail "dry run dmg"
pass "dry run does not call gh"

# Different Linux install scripts are not one release asset.
printf 'other\n' > "$repo/build/linux-arm64/install.sh"
if GITHUB_WORKSPACE=$repo BUILD_DIR="$repo/build" bash "$root/scripts/release/list_release_assets.sh" >/dev/null; then
  fail "differing install.sh was accepted"
fi
pass "linux install.sh must be identical"

# The workflow that builds v0.1.0 cannot live on that tag. It has to copy
# its scripts out, then check the tag out, and it must not use -upload.
python3 - "$workflow" <<'PY'
import sys
text = open(sys.argv[1]).read()
errors = []
commands = []
for line in text.splitlines():
    stripped = line.strip()
    if stripped.startswith("#"):
        continue
    commands.append(stripped)
body = "\n".join(commands)
if "-upload" in body:
    errors.append("workflow still passes -upload")
if "MYGO_UPDATER_PRIVATE_KEY" in body:
    errors.append("workflow still requests the updater key")
if "workflow_dispatch" not in text or "inputs:" not in text:
    errors.append("workflow_dispatch input missing")
if "git tag" in body or "gh release edit" in body or "--draft=false" in body:
    errors.append("workflow can move a tag or publish a release")
idx_copy = text.find('mkdir -p "$RUNNER_TEMP/kumo-release"')
idx_source = text.find("Checkout the release tag")
idx_workflow = text.find("Checkout the workflow commit")
if not (0 <= idx_workflow < idx_copy < idx_source):
    errors.append("release scripts are not preserved before the tag checkout")
if text.count('mkdir -p "$RUNNER_TEMP/kumo-release"') != 2:
    errors.append("both jobs must keep the scripts before checking out the tag")
if "needs.draft-release.result == 'skipped'" not in text:
    errors.append("dispatch cannot build when draft creation is skipped")
if "upload_release_assets.sh" not in text:
    errors.append("draft upload script is not used")
if errors:
    sys.exit("\n".join(errors))
PY
pass "workflow builds a named tag without -upload and does not publish"

echo "release asset checks passed"
