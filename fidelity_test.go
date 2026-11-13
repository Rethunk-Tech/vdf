package vdf

import (
	"errors"
	"runtime/debug"
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
		{"bare tokens", `"a"{ "k" -1 "z" #x timeout 5.0 Game |gameinfo_path|. url http://x/y path C:\x\y }`,
			map[string]any{"a": map[string]any{"k": "-1", "z": "#x", "timeout": "5.0", "Game": "|gameinfo_path|.", "url": "http://x/y", "path": `C:\x\y`}}},
		{"bare token ends at a brace or quote", `a{k"v"s{x y}}`, map[string]any{"a": map[string]any{"k": "v", "s": map[string]any{"x": "y"}}}},
		{"conditionals are skipped", `"a"{ "k" "v" [$WIN32] "z" [$OSX||$LINUX] "w" "s" [!$X360] { "x" "y" } [$WIN32] "q" "[$X]" }`,
			map[string]any{"a": map[string]any{"k": "v", "z": "w", "s": map[string]any{"x": "y"}, "q": "[$X]"}}},
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
		"unterminated section": `"a"{ "k" "v"`,
		"unterminated nested":  `"a"{ "k" "v" "s" { "x" "y" `,
		"key without value":    `"a"{ "k" }`,
		"dangling key at eof":  `"a"{ "k" "v" "z"`,
		"single slash":         `"a"{ "k" "v" / "z" "w" }`,
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

// Recursion would hit the stack limit, a fatal error no caller can recover from, so nesting depth must cost heap only.
func TestParse_DeepNestingStaysOffTheStack(t *testing.T) {
	defer debug.SetMaxStack(debug.SetMaxStack(256 << 10))

	const depth = 20000
	section := strings.Repeat(`"s"{`, depth) + strings.Repeat(`}`, depth)
	got, err := NewParser(strings.NewReader(`"a"{` + section + section + `}`)).Parse()
	require.NoError(t, err)
	require.NotNil(t, got)
}

func TestParse_ManyCommentLinesStayOffTheStack(t *testing.T) {
	defer debug.SetMaxStack(debug.SetMaxStack(256 << 10))

	got, err := NewParser(strings.NewReader(`"a"{` + strings.Repeat("//\n", 20000) + `"k" "v"}`)).Parse()
	require.NoError(t, err)
	require.Equal(t, map[string]any{"a": map[string]any{"k": "v"}}, got)
}
