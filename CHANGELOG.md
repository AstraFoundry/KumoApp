# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2026-11-18

Kumo is rewritten in Go with MyGo and a React frontend. Existing profiles,
overrides, settings and state carry over.

### Added
- Topology: a canvas of entry → rules → groups → nodes with selected paths,
  dynamic members, broken references, focus and a text inspector.
- Setup with an animated preview of each step, and a new About page.
- Agent prompts on every page, the Agents page, and `kumo prompt`.
- A guardian process that stops the core and clears the system proxy after
  a crash or force quit.
- Port conflicts name the program holding the port and offer free ports.
- JavaScript overrides (`main(config)`), Linux builds (amd64 and arm64).

### Changed
- A system proxy another app took over is shown and left alone; Kumo only
  clears a proxy it still owns, and can reclaim it on request.
- TUN uses a one-time permission grant for Mihomo instead of a helper service.
- New and unset TUN settings use the MIPS stack. MIPS needs Mihomo 1.19.32
  or later; an older core rejects that stack and does not start. A stack
  already stored is kept. The TUN page says so when MIPS is selected.
- Releases are tagged `vX.Y.Z`. The 0.0.x tags stay unprefixed.

### Fixed
- The profile import dialog keeps its footer inside the popup.

### Removed
- The Swift sources and the Xcode project.
- Bundled Sub-Store (deferred; see docs/roadmap).

### Limitations
- macOS builds are unsigned beyond ad hoc signing. Gatekeeper blocks them
  on other Macs. This release does not add Developer ID signing or
  notarization.
- In-app updates are off. About shows that updates are not available in
  this build, and the launch check does nothing. There are no update files
  and no updater key. The 0.0.x updater reads `latest.yml`, which this
  release does not publish, so those installs move to 0.1.0 by hand.
- Windows is not released.

## [0.0.15] - 2026-05-23

### Internationalization
- Replaced all hardcoded UI strings with `String(localized:)` across 25 Swift files for full i18n support.

### Infrastructure
- Bumped marketing version to 0.0.15.

## [0.0.14] and earlier

See git history for earlier changes.
