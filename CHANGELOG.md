# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - Unreleased

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
- JavaScript overrides (`main(config)`), Linux builds.

### Changed
- A system proxy another app took over is shown and left alone; Kumo only
  clears a proxy it still owns, and can reclaim it on request.
- TUN uses a one-time permission grant for Mihomo instead of a helper service.
- Releases are tagged `vX.Y.Z`; 0.0.x users install 0.1.0 manually once.

### Removed
- The Swift sources and the Xcode project.
- Bundled Sub-Store (deferred; see docs/roadmap).

## [0.0.15] - 2026-05-23

### Internationalization
- Replaced all hardcoded UI strings with `String(localized:)` across 25 Swift files for full i18n support.

### Infrastructure
- Bumped marketing version to 0.0.15.

## [0.0.14] and earlier

See git history for earlier changes.
