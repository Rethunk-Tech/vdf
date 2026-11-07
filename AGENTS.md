# AGENTS.md

`vdf` is a lexer, parser and writer for Valve's Data Format (VDF, KeyValues) text files. It is a
fork of andygrunwald/vdf kept in step with upstream where the fixes overlap, and it imports nothing
beyond the standard library (testify is test-only).

## Gate

`gate` is the command. Each override in `.gate.toml` is there because a runner
cannot resolve that check from PATH.

## Shape

- `Parse` returns `map[string]any` whose values are strings or nested maps. `Write` accepts exactly
  that shape and refuses everything else with `ErrUnsupportedValue`, so parse, write, parse yields the
  same map.
- `Write` sorts keys and builds the whole output before writing, so equal maps give equal bytes and a
  failed write leaves the writer untouched.
- The parser drops a backslash and keeps the character after it, so `quote` escapes only `"` and `\`;
  any other escape would not round-trip.
- Duplicate keys merge in `mergeMap`; the map cannot represent them, so `Write` never emits one.
- The parser has no depth or size limit. That bound is the caller's policy, not the package's.

## Files

| File | Holds |
| --- | --- |
| `token.go` | The `Token` kinds the scanner emits |
| `lexer.go` | `Scanner`: runes to tokens |
| `parser.go` | `Parser`: tokens to a nested map |
| `writer.go` | `Write`: a nested map back to VDF text |
| `errors.go` | Sentinel errors |
| `doc.go` | Package documentation |
