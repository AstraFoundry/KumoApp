#!/usr/bin/env bash
# Hermetic checks for the release asset scripts and the workflow contract.
set -euo pipefail

root=$(cd "$(dirname "$0")/../.." && pwd)
workflow=$root/.github/workflows/build-release.yml
# Child scripts use this interpreter. Running the suite with Bash 3.2
# therefore executes the release scripts with Bash 3.2, not the system bash.
shell=$BASH
echo "release script tests under ${shell} ($("$shell" -c 'echo "$BASH_VERSION"'))"
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
if GITHUB_WORKSPACE=$repo BUILD_DIR="$repo/build" "$shell" "$root/scripts/release/list_release_assets.sh" >/dev/null; then
  fail "update manifest was listed"
fi
pass "update manifests are refused"
rm "$repo/build/linux-amd64/update-linux-amd64.json"
listed=$(GITHUB_WORKSPACE=$repo BUILD_DIR="$repo/build" "$shell" "$root/scripts/release/list_release_assets.sh")
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

# One platform at a time. The macOS runner builds only darwin directories;
# Bash 3.2 aborted there on an empty Linux array. Linux-only is the other
# job. An empty tree and a platform missing its installer must fail.
write_config() {
  cat > "$1/mygo.config.ts" <<'EOF'
export default {
  name: "Kumo",
  version: "0.1.0",
}
EOF
}
list_of() {
  GITHUB_WORKSPACE=$1 BUILD_DIR=$2 "$shell" "$root/scripts/release/list_release_assets.sh"
}

donly=$tmp/darwin-only
mkdir -p "$donly/build/darwin-arm64" "$donly/build/darwin-amd64"
write_config "$donly"
printf 'arm\n' > "$donly/build/darwin-arm64/Kumo 0.1.0 arm64.dmg"
printf 'amd\n' > "$donly/build/darwin-amd64/Kumo 0.1.0 amd64.dmg"
mkdir -p "$donly/build/darwin-arm64/Kumo.app" "$donly/build/darwin-amd64/Kumo.app"
dlisted=$(list_of "$donly" "$donly/build")
dexpect=$'build/darwin-amd64/Kumo 0.1.0 amd64.dmg
build/darwin-arm64/Kumo 0.1.0 arm64.dmg'
dgot=$(printf '%s\n' "$dlisted" | sed "s|^$donly/||")
if [ "$dgot" != "$dexpect" ]; then
  printf 'darwin-only:\n%s\n' "$dgot" >&2
  fail "darwin-only list"
fi
pass "darwin-only lists both disk images"

lonly=$tmp/linux-only
mkdir -p "$lonly/build/linux-amd64" "$lonly/build/linux-arm64"
write_config "$lonly"
for arch in amd64 arm64; do
  printf 'deb\n' > "$lonly/build/linux-${arch}/kumo_0.1.0_${arch}.deb"
  printf 'tar\n' > "$lonly/build/linux-${arch}/kumo-0.1.0-linux-${arch}.tar.gz"
  printf 'install\n' > "$lonly/build/linux-${arch}/install.sh"
  printf 'bin\n' > "$lonly/build/linux-${arch}/kumo"
done
llisted=$(list_of "$lonly" "$lonly/build")
lexpect=$'build/linux-amd64/kumo_0.1.0_amd64.deb
build/linux-amd64/kumo-0.1.0-linux-amd64.tar.gz
build/linux-arm64/kumo_0.1.0_arm64.deb
build/linux-arm64/kumo-0.1.0-linux-arm64.tar.gz
build/linux-amd64/install.sh'
lgot=$(printf '%s\n' "$llisted" | sed "s|^$lonly/||")
if [ "$lgot" != "$lexpect" ]; then
  printf 'linux-only:\n%s\n' "$lgot" >&2
  fail "linux-only list"
fi
pass "linux-only lists deb, archive, and one install.sh"

empty=$tmp/empty
mkdir -p "$empty/build"
write_config "$empty"
if list_of "$empty" "$empty/build" >/dev/null; then
  fail "empty build was accepted"
fi
pass "empty build is rejected"

nodmg=$tmp/no-dmg
mkdir -p "$nodmg/build/darwin-arm64"
write_config "$nodmg"
if list_of "$nodmg" "$nodmg/build" >/dev/null; then
  fail "darwin directory without a disk image was accepted"
fi
pass "missing disk image is rejected"

badmix=$tmp/bad-mix
mkdir -p "$badmix/build/linux-amd64" "$badmix/build/darwin-arm64"
write_config "$badmix"
printf 'deb\n' > "$badmix/build/linux-amd64/kumo_0.1.0_amd64.deb"
printf 'tar\n' > "$badmix/build/linux-amd64/kumo-0.1.0-linux-amd64.tar.gz"
printf 'install\n' > "$badmix/build/linux-amd64/install.sh"
if list_of "$badmix" "$badmix/build" >/dev/null; then
  fail "linux plus a darwin directory without its disk image was accepted"
fi
pass "mixed output with a missing disk image is rejected"

if grep -n -E '=\(\)|\[@\]' \
  "$root/scripts/release/list_release_assets.sh" \
  "$root/scripts/release/upload_release_assets.sh" \
  "$root/scripts/release/assert_release_commit.sh"; then
  fail "an array expansion returned to a release script"
fi
pass "release scripts do not expand arrays"

if GITHUB_WORKSPACE=$repo RELEASE_REF=main \
  "$shell" "$root/scripts/release/assert_release_commit.sh" >/dev/null; then
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
out=$(GITHUB_WORKSPACE=$repo RELEASE_REF=v0.1.0 "$shell" "$root/scripts/release/assert_release_commit.sh")
printf '%s\n' "$out" | grep -q "source_sha=${tag_commit}" || fail "remote assert sha"
printf '%s\n' "$out" | grep -q "source_tag=v0.1.0" || fail "remote assert tag"
printf '%s\n' "$out" | grep -q "source_tag_object=${tag_object}" || fail "remote assert object"
after_refs=$(git -C "$repo" show-ref || true)
[ "$before_refs" = "$after_refs" ] || fail "assert wrote a local ref"
[ "$(git --git-dir="$remote" rev-parse refs/tags/v0.1.0)" = "$tag_object" ] || fail "remote tag object changed"
pass "annotated tag peeled from origin without writing refs/tags"

# A depth-1 checkout still proves the tag, and the proof does not rewrite
# its shallow file.
shallow=$tmp/shallow
git clone -q --depth 1 "file://${remote}" "$shallow"
git -C "$shallow" tag -d v0.1.0 >/dev/null 2>&1 || true
git -C "$shallow" checkout -q --detach "$tag_commit"
shallow_before=$(cat "$shallow/.git/shallow" 2>/dev/null || true)
shallow_out=$(GITHUB_WORKSPACE=$shallow RELEASE_REF=refs/tags/v0.1.0 \
  "$shell" "$root/scripts/release/assert_release_commit.sh")
shallow_after=$(cat "$shallow/.git/shallow" 2>/dev/null || true)
printf '%s\n' "$shallow_out" | grep -q "source_sha=${tag_commit}" || fail "shallow assert"
[ "$shallow_before" = "$shallow_after" ] || fail "shallow file changed"
if git -C "$shallow" show-ref --tags | grep -q .; then
  fail "shallow checkout gained a tag"
fi
pass "shallow checkout peels the tag without changing the shallow file"

dry=$(GITHUB_WORKSPACE=$repo RELEASE_REF=v0.1.0 DRY_RUN=1 \
  "$shell" "$root/scripts/release/upload_release_assets.sh")
printf '%s\n' "$dry" | grep -q 'dry run: not calling gh' || fail "dry run"
printf '%s\n' "$dry" | grep -q 'Kumo 0.1.0 arm64.dmg' || fail "dry run dmg"
pass "dry run does not call gh"

git -C "$repo" checkout -q --detach "$newer_commit"
if GITHUB_WORKSPACE=$repo RELEASE_REF=v0.1.0 \
  "$shell" "$root/scripts/release/assert_release_commit.sh" >/dev/null; then
  fail "a newer commit was accepted from the remote tag"
fi
pass "newer commit is not the remote tag"

missing=$tmp/missing
git clone -q "$remote" "$missing"
git --git-dir="$missing/.git" tag -d v0.1.0 >/dev/null
git --git-dir="$remote" update-ref -d refs/tags/v0.1.0
if GITHUB_WORKSPACE=$missing RELEASE_REF=v0.1.0 \
  "$shell" "$root/scripts/release/assert_release_commit.sh" >/dev/null; then
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
  "$shell" "$root/scripts/release/assert_release_commit.sh" >/dev/null; then
  fail "lightweight tag was accepted"
fi
pass "lightweight tag is rejected"

# Read-only proof of the published tag. No fetch and no shallow file.
# Each attempt is logged on its own so a failure names that read.
project_origin=$(git -C "$root" remote get-url origin)
live_attempt() {
  n=$1
  echo "live ls-remote attempt ${n}"
  if ! listing=$(git ls-remote "$project_origin" "refs/tags/v0.1.0" "refs/tags/v0.1.0^{}"); then
    echo "live ls-remote attempt ${n} command failed" >&2
    fail "live ls-remote ${n}"
  fi
  printf 'attempt %s raw:\n%s\n' "$n" "$listing"
  pointer=$(printf '%s\n' "$listing" | awk -v want="refs/tags/v0.1.0" '$2 == want { print $1 }')
  peeled=$(printf '%s\n' "$listing" | awk -v want="refs/tags/v0.1.0^{}" '$2 == want { print $1 }')
  if [ "$pointer" != "8f8e68ea9f3bb821d0bb5fa1e9c1015b04b99201" ] || \
     [ "$peeled" != "9d299814c3ca6f4d4c937966044d2ec4a8455a97" ]; then
    echo "live ls-remote attempt ${n} unexpected" >&2
    fail "live ls-remote ${n} values"
  fi
}
n=1
while [ "$n" -le 5 ]; do
  live_attempt "$n"
  n=$((n + 1))
done
pass "five read-only ls-remote reads match the v0.1.0 object and commit"

probe=$tmp/probe
git init -q "$probe"
git -C "$probe" remote add origin "$project_origin"
printf '9d299814c3ca6f4d4c937966044d2ec4a8455a97\n' > "$probe/.git/HEAD"
printf '  version: "0.1.0",\n' > "$probe/mygo.config.ts"
probe_out=$(GITHUB_WORKSPACE=$probe RELEASE_REF=v0.1.0 "$shell" "$root/scripts/release/assert_release_commit.sh")
printf '%s\n' "$probe_out" | grep -q "source_sha=9d299814c3ca6f4d4c937966044d2ec4a8455a97" || fail "probe sha"
printf '%s\n' "$probe_out" | grep -q "source_tag_object=8f8e68ea9f3bb821d0bb5fa1e9c1015b04b99201" || fail "probe object"
[ ! -f "$probe/.git/shallow" ] || fail "probe created a shallow file"
pass "production check accepts v0.1.0 without fetching"

printf '4d4981e52eaced7ec6401c57221d2c731967c9db\n' > "$probe/.git/HEAD"
if GITHUB_WORKSPACE=$probe RELEASE_REF=v0.1.0 \
  "$shell" "$root/scripts/release/assert_release_commit.sh" >/dev/null; then
  fail "another commit was accepted for the published tag"
fi
pass "production check rejects a different commit for v0.1.0"

# Different Linux install scripts are not one release asset.
printf 'other\n' > "$repo/build/linux-arm64/install.sh"
if GITHUB_WORKSPACE=$repo BUILD_DIR="$repo/build" "$shell" "$root/scripts/release/list_release_assets.sh" >/dev/null; then
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
