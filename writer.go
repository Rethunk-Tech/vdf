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
		if err := writeSection(&buf, key, root, 0); err != nil {
			return err
		}
	}

	_, err := w.Write(buf.Bytes())

	return err
}

func writeSection(buf *bytes.Buffer, key string, m map[string]any, depth int) error {
	indent := strings.Repeat("\t", depth)
	buf.WriteString(indent + quote(key) + "\n" + indent + "{\n")

	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)

	for _, k := range keys {
		switch v := m[k].(type) {
		case string:
			buf.WriteString(indent + "\t" + quote(k) + "\t\t" + quote(v) + "\n")
		case map[string]any:
			if err := writeSection(buf, k, v, depth+1); err != nil {
				return err
			}
		default:
			return fmt.Errorf("%w: key %q holds %T, want string or map[string]any", ErrUnsupportedValue, k, v)
		}
	}

	buf.WriteString(indent + "}\n")

	return nil
}

// quote wraps s in quotation marks. The parser drops a backslash and keeps the
// character after it, so only the quotation mark and the backslash need one.
func quote(s string) string {
	return `"` + strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(s) + `"`
}
