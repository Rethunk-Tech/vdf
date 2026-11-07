# HUMANS.md

Run and use `vdf`. Internals: [AGENTS.md](AGENTS.md).

## Quick start

```bash
go get github.com/Rethunk-Tech/vdf/v2
```

```go
package main

import (
	"os"
	"strings"

	"github.com/Rethunk-Tech/vdf/v2"
)

func main() {
	m, err := vdf.NewParser(strings.NewReader(`"Root" { "name" "mortar" }`)).Parse()
	if err != nil {
		panic(err)
	}

	if err := vdf.Write(os.Stdout, m); err != nil {
		panic(err)
	}
}
```

## Contracts

| Entry point | Contract |
| --- | --- |
| `NewParser(r).Parse()` | Reads the first top-level section into `map[string]any` (strings and nested maps); a malformed document returns `ErrNotValidFormat` or a descriptive error, never a partial map. Duplicate keys merge. |
| `Write(w, m)` | Writes `m` so `Parse` reads it back equal; keys sorted, tab indented. Exactly one root key holding a map, otherwise `ErrUnsupportedValue` and nothing written. |
| `NewScanner(r).Scan` | Token stream for callers that need positions of tokens rather than a map. |

## Requirements

Go 1.27.3 or newer, plus `lefthook` and `gitleaks` on PATH if you want the git hooks.

```bash
lefthook install
```

That is per clone; git does not commit hooks. Pre-commit scans staged changes for secrets, and
pre-push runs `gate`.

## Verify

```bash
gate
```
