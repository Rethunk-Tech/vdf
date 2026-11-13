# Changelog

Notable changes to `vdf`. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and versions follow
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Fixed

- Duplicate sections merge recursively, as in Valve's KeyValues; a section after a same-named string key is kept.
- `Parse` returns `ErrNotValidFormat` for a truncated section, a key without a value or a stray token, instead of a partial map.
- Bare tokens such as `5.0`, `-1` and `|gameinfo_path|.` are read whole, as in Valve's KeyValues.
- Platform conditionals such as `[$WIN32]` are skipped instead of being read as keys.
- Deep nesting and long comment runs no longer overflow the stack, and `Write` output stays linear in nesting depth.

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
