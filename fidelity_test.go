package vdf

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type failAfter struct {
	s    string
	n, i int
}

func (f *failAfter) Read(p []byte) (int, error) {
	if f.i >= f.n {
		return 0, errors.New("disk error")
	}
	p[0] = f.s[f.i]
	f.i++

	return 1, nil
}

func TestParse_Fidelity(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want map[string]any
	}{
		{"utf8 bom", "\ufeff\"a\"\n{\n\"k\" \"v\"\n}", map[string]any{"a": map[string]any{"k": "v"}}},
		{"crlf", "\"a\"\r\n{\r\n\"k\" \"v\"\r\n\"z\" \"w\"\r\n}\r\n", map[string]any{"a": map[string]any{"k": "v", "z": "w"}}},
		{"crlf comment", "\"a\"{ // c\r\n\"k\" \"v\"\r\n}", map[string]any{"a": map[string]any{"k": "v"}}},
		{"unicode", "\"a\"{ \"k\" \"日本語\" \"é\" \"w\" }", map[string]any{"a": map[string]any{"k": "日本語", "é": "w"}}},
		{"url in value", "\"a\"{ \"k\" \"http://x/y\" }", map[string]any{"a": map[string]any{"k": "http://x/y"}}},
		{"escaped quote and backslash", `"a"{ "k" "say \"hi\" C:\\x" }`, map[string]any{"a": map[string]any{"k": `say "hi" C:\x`}}},
		{"duplicate sections merge", `"a"{ "s" {"p" "1"} "s" {"q" "2"} }`, map[string]any{"a": map[string]any{"s": map[string]any{"p": "1", "q": "2"}}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewParser(strings.NewReader(tt.in)).Parse()
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestParse_RejectsInsteadOfTruncating(t *testing.T) {
	for name, in := range map[string]string{
		"unquoted negative value": `"a"{ "k" -1 "z" "w" }`,
		"unquoted hash value":     `"a"{ "k" #x "z" "w" }`,
	} {
		t.Run(name, func(t *testing.T) {
			_, err := NewParser(strings.NewReader(in)).Parse()
			require.ErrorIs(t, err, ErrNotValidFormat)
		})
	}
}

func TestParse_ReadErrorSurfaces(t *testing.T) {
	for _, n := range []int{12, 8} {
		_, err := NewParser(&failAfter{s: `"a"{ "k" "v" "z" "w" }`, n: n}).Parse()
		require.ErrorContains(t, err, "disk error")
	}
}

func TestParse_DeepNesting(t *testing.T) {
	const depth = 50000
	in := `"a"` + strings.Repeat(`{ "s" `, depth) + `{ "k" "v" }` + strings.Repeat(`}`, depth)
	_, err := NewParser(strings.NewReader(in)).Parse()
	require.NoError(t, err)
}
