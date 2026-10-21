# Release Management

Kumo is built with `mygo build`, which runs `bun run build:web`
(typecheck + Vite), embeds `dist/` into the Go binary and packages it.

| Platform | Output in `build/` |
| --- | --- |
| macOS (arm64, amd64) | `Kumo.app` and a disk image |
| Linux (amd64, arm64) | the binary, a `.tar.gz`, a `.deb`, `kumo.desktop` and `install.sh` |

Windows builds work with the same command but are not released yet: the
system proxy adapter is written, TUN permission and the CLI link are not.

## Versioning

`version` in `mygo.config.ts` is the release version. Tags carry a `v`
prefix (`v0.1.0`). Releases of the Swift app were tagged without it
(`0.0.15`); they stay as they are.

## Releasing

1. Update `version` in `mygo.config.ts` and add a `## [x.y.z]` section to
   `CHANGELOG.md` (the update window shows it).
2. `make check`, then `make e2e MIHOMO=<path>` on Linux.
3. Commit, `git tag vX.Y.Z`, `git push origin vX.Y.Z`.
4. `.github/workflows/build-release.yml` tests, creates a draft release,
   builds macOS and Linux and uploads to the draft. Review and publish it.

## Updates

MyGo's updater checks the latest GitHub release, verifies the signature
against the public key built into the app and replaces the app. Turning it
on needs the maintainers' key pair, which is not in the repository:

```sh
mygo keygen                                   # writes mygo-update.key / .pub
gh secret set MYGO_UPDATER_PRIVATE_KEY < mygo-update.key
```

and in `mygo.config.ts`:

```ts
updates: { publicKey: "<contents of mygo-update.pub>", github: "AstraFoundry/KumoApp" },
```

Until then builds have no update files, About shows that updates are not
available, and the launch check does nothing.

The Swift app's updater reads `latest.yml`, which MyGo releases don't
publish, so users of 0.0.x install 0.1.0 by hand once. Their data carries
over (see [persistence](persistence-logging.md)).

## Signing

Unsigned macOS builds are signed ad hoc and blocked by Gatekeeper on other
Macs. To sign and notarize, add the Developer ID certificate and notary
credentials as secrets and import them before the build step, as described
in MyGo's `docs/github-actions.md` ("Signing and notarizing macOS apps").

## Telegram

`.github/workflows/telegram-release.yml` posts the published release's notes
and disk images to the Telegram channel (`scripts/telegram_release_notify.py`).
