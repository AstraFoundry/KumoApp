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
newer_commit=$(git -C "$repo" rev-parse HEAD)
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

if GITHUB_WORKSPACE=$repo RELEASE_REF=main \
  bash "$root/scripts/release/assert_release_commit.sh" >/dev/null; then
  fail "branch ref was accepted"
fi
pass "only the version tag is accepted"

# The published tag is an annotated tag on a remote. Exact ls-remote of that
# ref returns the tag object only; the script must still peel it, and must
# not write refs/tags.
remote=$tmp/remote
git init -q --bare "$remote"
git --git-dir="$remote" symbolic-ref HEAD refs/heads/main
git -C "$repo" remote add origin "$remote"
git -C "$repo" push -q origin "$tag_commit:refs/heads/main" "refs/tags/v0.1.0:refs/tags/v0.1.0"
tag_object=$(git --git-dir="$remote" rev-parse refs/tags/v0.1.0)
exact=$(git ls-remote --tags "$remote" "refs/tags/v0.1.0")
printf '%s\n' "$exact" | grep -q '\^{}' && fail "exact ls-remote peeled the annotated tag"
printf '%s\n' "$exact" | grep -q "$tag_object" || fail "exact ls-remote missed the tag object"
pass "exact ls-remote returns the tag object only"

git -C "$repo" checkout -q --detach "$tag_commit"
git -C "$repo" tag -d v0.1.0 >/dev/null
before_refs=$(git -C "$repo" show-ref || true)
out=$(GITHUB_WORKSPACE=$repo RELEASE_REF=v0.1.0 bash "$root/scripts/release/assert_release_commit.sh")
printf '%s\n' "$out" | grep -q "source_sha=${tag_commit}" || fail "remote assert sha"
printf '%s\n' "$out" | grep -q "source_tag=v0.1.0" || fail "remote assert tag"
printf '%s\n' "$out" | grep -q "source_tag_object=${tag_object}" || fail "remote assert object"
after_refs=$(git -C "$repo" show-ref || true)
[ "$before_refs" = "$after_refs" ] || fail "assert wrote a local ref"
[ "$(git --git-dir="$remote" rev-parse refs/tags/v0.1.0)" = "$tag_object" ] || fail "remote tag object changed"
pass "annotated tag peeled from origin without writing refs/tags"

# A depth-1 checkout, as Actions uses, still peels the remote tag.
shallow=$tmp/shallow
git clone -q --depth 1 "file://${remote}" "$shallow"
git -C "$shallow" tag -d v0.1.0 >/dev/null 2>&1 || true
git -C "$shallow" checkout -q --detach "$tag_commit"
shallow_out=$(GITHUB_WORKSPACE=$shallow RELEASE_REF=refs/tags/v0.1.0 \
  bash "$root/scripts/release/assert_release_commit.sh")
printf '%s\n' "$shallow_out" | grep -q "source_sha=${tag_commit}" || fail "shallow assert"
if git -C "$shallow" show-ref --tags | grep -q .; then
  fail "shallow checkout gained a tag"
fi
pass "shallow checkout peels the annotated tag"

dry=$(GITHUB_WORKSPACE=$repo RELEASE_REF=v0.1.0 DRY_RUN=1 \
  bash "$root/scripts/release/upload_release_assets.sh")
printf '%s\n' "$dry" | grep -q 'dry run: not calling gh' || fail "dry run"
printf '%s\n' "$dry" | grep -q 'Kumo 0.1.0 arm64.dmg' || fail "dry run dmg"
pass "dry run does not call gh"

git -C "$repo" checkout -q --detach "$newer_commit"
if GITHUB_WORKSPACE=$repo RELEASE_REF=v0.1.0 \
  bash "$root/scripts/release/assert_release_commit.sh" >/dev/null; then
  fail "a newer commit was accepted from the remote tag"
fi
pass "newer commit is not the remote tag"

missing=$tmp/missing
git clone -q "$remote" "$missing"
git --git-dir="$missing/.git" tag -d v0.1.0 >/dev/null
git --git-dir="$remote" update-ref -d refs/tags/v0.1.0
if GITHUB_WORKSPACE=$missing RELEASE_REF=v0.1.0 \
  bash "$root/scripts/release/assert_release_commit.sh" >/dev/null; then
  fail "missing tag was accepted"
fi
pass "missing tag is rejected"
git --git-dir="$remote" update-ref refs/tags/v0.1.0 "$tag_object"

light=$tmp/light
git init -q --bare "$light/bare"
git clone -q "$remote" "$light/src"
git -C "$light/src" remote set-url origin "$light/bare"
git -C "$light/src" push -q origin "$tag_commit:refs/heads/main"
git -C "$light/src" tag lighthead "$tag_commit"
git -C "$light/src" push -q origin "refs/tags/lighthead:refs/tags/v0.1.0"
git -C "$light/src" checkout -q --detach "$tag_commit"
if GITHUB_WORKSPACE=$light/src RELEASE_REF=v0.1.0 \
  bash "$root/scripts/release/assert_release_commit.sh" >/dev/null; then
  fail "lightweight tag was accepted"
fi
pass "lightweight tag is rejected"

# The real annotated v0.1.0 tag. Read-only: a temporary clone fetches it and
# must not change the published tag object.
project_origin=$(git -C "$root" remote get-url origin)
real=$tmp/real
git init -q "$real"
git -C "$real" remote add origin "$project_origin"
git -C "$real" fetch -q --depth 1 origin 9d299814c3ca6f4d4c937966044d2ec4a8455a97
git -C "$real" checkout -q --detach FETCH_HEAD
real_out=$(GITHUB_WORKSPACE=$real RELEASE_REF=v0.1.0 bash "$root/scripts/release/assert_release_commit.sh")
printf '%s\n' "$real_out" | grep -q "source_sha=9d299814c3ca6f4d4c937966044d2ec4a8455a97" || fail "real tag sha"
printf '%s\n' "$real_out" | grep -q "source_tag_object=8f8e68ea9f3bb821d0bb5fa1e9c1015b04b99201" || fail "real tag object"
if git -C "$real" show-ref --tags | grep -q .; then
  fail "real check created a local tag"
fi
git -C "$real" fetch -q --depth 1 origin main
git -C "$real" checkout -q --detach FETCH_HEAD
if GITHUB_WORKSPACE=$real RELEASE_REF=v0.1.0 \
  bash "$root/scripts/release/assert_release_commit.sh" >/dev/null; then
  fail "main was accepted as v0.1.0"
fi
published=$(git ls-remote --tags "$project_origin" "refs/tags/v0.1.0")
printf '%s\n' "$published" | grep -q '^8f8e68ea9f3bb821d0bb5fa1e9c1015b04b99201[[:space:]]refs/tags/v0.1.0$' || fail "published tag object changed"
printf '%s\n' "$published" | grep -q '\^{}' && fail "published exact ls-remote grew a peel line" || true
pass "real v0.1.0 peels to 9d299814 and the tag object is unchanged"

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
