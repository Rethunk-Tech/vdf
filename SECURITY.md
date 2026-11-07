# Security

## Supported versions

The latest tagged release. There are no long-lived release branches, so a fix ships in the next tag
rather than being backported.

## Reporting a vulnerability

Use GitHub's **private vulnerability reporting** on this repository (Security → Report a
vulnerability). It stays private until a fix ships.

Please do not open a public issue for a suspected vulnerability first.

## What is in scope

- `Parser.Parse` and `Scanner.Scan` read arbitrary text. A crafted input must not panic, loop
  forever or read outside the reader; the fuzz targets in the test files cover this.
- `Write` renders caller-built maps. A key or value holding a quotation mark, a backslash or a
  newline must not change the document's structure, which `quote` and the round-trip fuzz target
  check.

## What is out of scope

- Input size and nesting depth. Parsing recurses once per nested section, so a caller reading
  untrusted files bounds the reader and rejects absurd depth before or around `Parse`.
- Which keys and values an application accepts after parsing.
