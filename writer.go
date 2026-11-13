package vdf

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"
)

// ErrUnsupportedValue is returned by Write for a map that Parse could not have produced.
var ErrUnsupportedValue = errors.New("unsupported vdf value")

// Write writes m to w in the format Parse reads, so that parsing the output yields m again.
//
// m must hold exactly one key, whose value is a map, like the result of Parse.
// Every other value is a string or a map[string]any. Keys are written in sorted
// order, so equal maps produce equal output. Nothing is written to w on error.
func Write(w io.Writer, m map[string]any) error {
	if len(m) != 1 {
		return fmt.Errorf("%w: want exactly one root key, got %d", ErrUnsupportedValue, len(m))
	}

	var buf bytes.Buffer
	for key, value := range m {
		root, ok := value.(map[string]any)
		if !ok {
			return fmt.Errorf("%w: root key %q holds %T, want map[string]any", ErrUnsupportedValue, key, value)
		}
		if err := writeSection(&buf, key, root); err != nil {
			return err
		}
	}

	_, err := w.Write(buf.Bytes())

	return err
}

// maxIndent bounds the tabs on a line. Indenting every level makes the output quadratic in depth;
// past this depth nobody reads the nesting from the tabs, and the parser ignores them anyway.
const maxIndent = 64

// writeSection keeps open sections on a slice rather than the call stack: Go aborts the process on
// stack exhaustion, so nesting depth may only cost heap.
func writeSection(buf *bytes.Buffer, key string, m map[string]any) error {
	type section struct {
		m    map[string]any
		keys []string // not yet written, in order
	}
	indent := func(depth int) string { return strings.Repeat("\t", min(depth, maxIndent)) }
	start := func(key string, m map[string]any, depth int) section {
		buf.WriteString(indent(depth) + quote(key) + "\n" + indent(depth) + "{\n")
		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		slices.Sort(keys)

		return section{m: m, keys: keys}
	}

	open := []section{start(key, m, 0)}
	for len(open) > 0 {
		depth := len(open) - 1
		top := &open[depth]
		if len(top.keys) == 0 {
			buf.WriteString(indent(depth) + "}\n")
			open = open[:depth]
			continue
		}
		k := top.keys[0]
		top.keys = top.keys[1:]
		switch v := top.m[k].(type) {
		case string:
			buf.WriteString(indent(depth+1) + quote(k) + "\t\t" + quote(v) + "\n")
		case map[string]any:
			open = append(open, start(k, v, depth+1))
		default:
			return fmt.Errorf("%w: key %q holds %T, want string or map[string]any", ErrUnsupportedValue, k, v)
		}
	}

	return nil
}

// quote wraps s in quotation marks. The parser drops a backslash and keeps the
// character after it, so only the quotation mark and the backslash need one.
func quote(s string) string {
	return `"` + strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(s) + `"`
}
