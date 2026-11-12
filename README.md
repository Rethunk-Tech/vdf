<h1 align="center">vdf</h1>

<div align="center">

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go](https://img.shields.io/badge/go-1.27.3+-00ADD8?logo=go&logoColor=white)](go.mod)

</div>

---

A lexer, parser and writer for Valve's Data Format (VDF, KeyValues). Package `vdf`, module `github.com/Rethunk-Tech/vdf/v2`.

Parses [Valve's Data Format](https://developer.valvesoftware.com/wiki/KeyValues) (Steam's
`libraryfolders.vdf`, `appmanifest_*.acf`, game configs) into nested maps and writes such a map back
out. It is a fork of [andygrunwald/vdf](https://github.com/andygrunwald/vdf), created by Andy
Grunwald, which this repository keeps the history and MIT license of.

## Quick start

```bash
go get github.com/Rethunk-Tech/vdf/v2
```

A worked example and the contracts: [HUMANS.md](HUMANS.md).

## Features

- `NewParser(r).Parse()` reads a document into `map[string]any` of strings and nested maps.
- `Write` renders that map back out with sorted keys; parse, write, parse yields the same map.
- Quoted and bare identifiers, escapes, `//` comments and duplicate-key merging.
- A scanner (`NewScanner`) for callers that want tokens.
- Fuzz targets for the scanner, parser and the write round trip.

## Documentation

| Doc | Audience |
| --- | --- |
| [HUMANS.md](HUMANS.md) | Running and using the package |
| [AGENTS.md](AGENTS.md) | Internals, invariants, and why |
| [CONTRIBUTING.md](CONTRIBUTING.md) | Process — commits, tests, docs |
| [SECURITY.md](SECURITY.md) | Vulnerability disclosure |
| [CHANGELOG.md](CHANGELOG.md) | Release notes |

## License

MIT — see [LICENSE](LICENSE). © 2016 Andy Grunwald, © 2026 Rethunk.Tech, LLC.
