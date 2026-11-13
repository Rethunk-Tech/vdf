package vdf

import (
	"bytes"
	"errors"
	"os"
	"runtime/debug"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWrite_Format(t *testing.T) {
	var buf bytes.Buffer
	err := Write(&buf, map[string]any{
		"Root": map[string]any{
			"b":   "2",
			"a":   "1",
			"sub": map[string]any{"z": "last"},
		},
	})
	require.NoError(t, err)
	require.Equal(t, "\"Root\"\n{\n\t\"a\"\t\t\"1\"\n\t\"b\"\t\t\"2\"\n\t\"sub\"\n\t{\n\t\t\"z\"\t\t\"last\"\n\t}\n}\n", buf.String())
}

func TestWrite_RoundTrip(t *testing.T) {
	cases := map[string]map[string]any{
		"quote and backslash": {"R": map[string]any{`k"ey`: `va\lue "x"`, `\`: `\\`}},
		"empty":               {"R": map[string]any{}},
		"empty strings":       {"": map[string]any{"": ""}},
		"newline and tab":     {"R": map[string]any{"k": "a\nb\tc"}},
		"comment marker":      {"R": map[string]any{"uri": "http://127.0.0.1:3000"}},
		"braces":              {"R": map[string]any{"k": "{ }", "{": "}"}},
		"nested":              {"R": map[string]any{"a": map[string]any{"b": map[string]any{"c": "d"}}}},
		"unicode":             {"R": map[string]any{"ключ": "値"}},
	}
	for name, want := range cases {
		t.Run(name, func(t *testing.T) {
			var buf bytes.Buffer
			require.NoError(t, Write(&buf, want))
			got, err := NewParser(&buf).Parse()
			require.NoError(t, err)
			require.Equal(t, want, got)
		})
	}
}

func TestWrite_RoundTripTestdata(t *testing.T) {
	for _, name := range []string{"ok.vdf", "gamestate_integration_consolesample.cfg"} {
		t.Run(name, func(t *testing.T) {
			root, err := os.OpenRoot("testdata")
			require.NoError(t, err)
			src, err := root.ReadFile(name)
			require.NoError(t, err)
			first, err := NewParser(bytes.NewReader(src)).Parse()
			require.NoError(t, err)

			var out1, out2 bytes.Buffer
			require.NoError(t, Write(&out1, first))
			second, err := NewParser(bytes.NewReader(out1.Bytes())).Parse()
			require.NoError(t, err)
			require.Equal(t, first, second)

			require.NoError(t, Write(&out2, second))
			require.Equal(t, out1.String(), out2.String())
		})
	}
}

func TestWrite_Unsupported(t *testing.T) {
	cases := map[string]map[string]any{
		"nil map":      nil,
		"two roots":    {"a": map[string]any{}, "b": map[string]any{}},
		"string root":  {"a": "b"},
		"int value":    {"a": map[string]any{"k": 1}},
		"nested slice": {"a": map[string]any{"k": map[string]any{"j": []string{}}}},
	}
	for name, m := range cases {
		t.Run(name, func(t *testing.T) {
			var buf bytes.Buffer
			err := Write(&buf, m)
			require.ErrorIs(t, err, ErrUnsupportedValue)
			require.Zero(t, buf.Len())
		})
	}
}

func FuzzWrite_RoundTrip(f *testing.F) {
	f.Add("\"R\"\n{\n\t\"k\"\t\"v\"\n\t\"s\"\n\t{\n\t\t\"a\\\"b\"\t\"c\"\n\t}\n}")
	f.Fuzz(func(t *testing.T, input string) {
		first, err := NewParser(strings.NewReader(input)).Parse()
		if err != nil {
			t.Skip()
		}
		var buf bytes.Buffer
		if err := Write(&buf, first); err != nil {
			if errors.Is(err, ErrUnsupportedValue) && len(first) != 1 {
				t.Skip()
			}
			t.Fatal(err)
		}
		second, err := NewParser(&buf).Parse()
		require.NoError(t, err)
		require.Equal(t, first, second)
	})
}

// A recursive writer would hit the stack limit, and indenting every level would make the output quadratic in depth.
func TestWrite_DeepNesting(t *testing.T) {
	restore := debug.SetMaxStack(256 << 10)
	defer debug.SetMaxStack(restore)

	const depth = 20000
	in := `"a"{` + strings.Repeat(`"s"{`, depth) + `"k" "v"` + strings.Repeat(`}`, depth) + `}`
	m, err := NewParser(strings.NewReader(in)).Parse()
	require.NoError(t, err)

	var buf bytes.Buffer
	require.NoError(t, Write(&buf, m))
	require.Less(t, buf.Len(), 300*depth)

	again, err := NewParser(&buf).Parse()
	require.NoError(t, err)
	debug.SetMaxStack(restore) // reflect.DeepEqual recurses
	require.Equal(t, m, again)
}
