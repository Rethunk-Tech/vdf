# Changelog

Notable changes to `vdf`. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and versions follow
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [2.1.0]

### Added

- `Write` writes a parsed map back out as VDF text with sorted keys; parsing the output yields the
  same map. `ErrUnsupportedValue` reports a map it cannot write.

### Changed

- The module path is `github.com/Rethunk-Tech/vdf/v2`.

### Fixed

- `Parse` returns on a comment at the end of the input instead of looping.

[Unreleased]: https://github.com/Rethunk-Tech/vdf/compare/v2.1.0...HEAD
[2.1.0]: https://github.com/Rethunk-Tech/vdf/releases/tag/v2.1.0
