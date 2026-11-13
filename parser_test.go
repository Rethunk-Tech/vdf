package vdf

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Ensure the parser can parse strings into Statement ASTs.
func TestParser_ParseStatement(t *testing.T) {
	tests := []struct {
		s   string
		m   map[string]any
		err error
	}{
		// Single field statement
		{
			s: `"Example"
{
	"TimeNextStatsReport"      "1234567890"
	"ContentStatsID"           "-7123456789012345678"
}`,
			m: map[string]any{
				"Example": map[string]any{
					"TimeNextStatsReport": "1234567890",
					"ContentStatsID":      "-7123456789012345678",
				},
			},
		},
		{
			s: `"Root"
{
 "attr1"       "hey-ho"

 "attr2"       "ho-hey"
 "map1"
 {
   "foo"       "Q79v5tbar"
 }
 "data"
 {
   "val"       "1"
   "map"       "2"
   "player"    "3"
 }
}`,
			m: map[string]any{
				"Root": map[string]any{
					"attr1": "hey-ho",
					"attr2": "ho-hey",
					"map1": map[string]any{
						"foo": "Q79v5tbar",
					},
					"data": map[string]any{
						"val":    "1",
						"map":    "2",
						"player": "3",
					},
				},
			},
		},
		{
			s: `"Root"
{
 attr1       "hey-ho"
 "attr2"       ho
 "map1"
 {
   "foo"       "Q79v5tbar"
 }
 "data"
 {
   "v\\al"       "1"
   "map"       "2"
   "pl\"ayer"    "3"
 }
}`,
			m: map[string]any{
				"Root": map[string]any{
					"attr1": "hey-ho",
					"attr2": "ho",
					"map1": map[string]any{
						"foo": "Q79v5tbar",
					},
					"data": map[string]any{
						"v\\al":    "1",
						"map":      "2",
						"pl\"ayer": "3",
					},
				},
			},
		},
		{
			s: `"Root"
{
 attr1       "hey-ho"
 "attr2"       ho
 "map1"
 {
   // This is a comment
   "foo"       "Q79v5tbar"
 }
 "data"
 {
   "v\\al"       "1"
   "map"       "2"
   "pl\"ayer"    "3"
 }
}`,
			m: map[string]any{
				"Root": map[string]any{
					"attr1": "hey-ho",
					"attr2": "ho",
					"map1": map[string]any{
						"foo": "Q79v5tbar",
					},
					"data": map[string]any{
						"v\\al":    "1",
						"map":      "2",
						"pl\"ayer": "3",
					},
				},
			},
		},
		{
			s: `"Root"
{
 attr1       "hey-ho"
 "attr2"       ho
 "map1"
 {
   // Comment line first
   // Comment line second
   "foo"       "Q79v5tbar"
   // Comment line third
 }
 "data"
 {
   "v\\al"       "1"
   "map"       "2"
   "pl\"ayer"    "3"
 }
}`,
			m: map[string]any{
				"Root": map[string]any{
					"attr1": "hey-ho",
					"attr2": "ho",
					"map1": map[string]any{
						"foo": "Q79v5tbar",
					},
					"data": map[string]any{
						"v\\al":    "1",
						"map":      "2",
						"pl\"ayer": "3",
					},
				},
			},
		},
		{
			s: `// Root comment line
"Root"
{
 attr1       "hey-ho"
 "attr2"       ho
}`,
			m: map[string]any{
				"Root": map[string]any{
					"attr1": "hey-ho",
					"attr2": "ho",
				},
			},
		},
		{
			s: `"Root"
{
 "map"
 {
   "attr1" "hello"
 }
 "map"
 {
   "attr2" "world"
 }
}`,
			m: map[string]any{
				"Root": map[string]any{
					"map": map[string]any{
						"attr1": "hello",
						"attr2": "world",
					},
				},
			},
		},
	}

	for i, tt := range tests {
		m, err := NewParser(strings.NewReader(tt.s)).Parse()
		if tt.err != nil {
			if err == nil || err.Error() != tt.err.Error() {
				t.Errorf("%d. %q: error mismatch:\n  exp=%s\n  got=%s\n\n", i, tt.s, tt.err, err)
			}
		} else if err != nil {
			t.Errorf("%d. %q: unexpected error: %s", i, tt.s, err)
		} else if !reflect.DeepEqual(tt.m, m) {
			t.Errorf("%d. %q\n\nparse mismatch:\n\nexp=%#v\n\ngot=%#v\n\n", i, tt.s, tt.m, m)
		}
	}
}

func TestParser_Parse(t *testing.T) {
	getReader := func(fileName string) (io.Reader, error) {
		root, err := os.OpenRoot("testdata")
		if err != nil {
			return nil, err
		}
		return root.Open(filepath.Base(fileName))
	}

	tests := []struct {
		name     string
		fileName string
		want     func(got map[string]any, err error)
	}{
		{
			name:     "empty file",
			fileName: "testdata/empty.vdf",
			want: func(got map[string]any, err error) {
				require.Error(t, err)
			},
		},
		{
			name:     "corrupted file",
			fileName: "testdata/corrupted.vdf",
			want: func(got map[string]any, err error) {
				require.ErrorIs(t, err, ErrNotValidFormat)
			},
		},
		{
			name:     "ok",
			fileName: "testdata/ok.vdf",
			want: func(got map[string]any, err error) {
				require.NoError(t, err)
				saveFile, ok := got["SaveFile"].(map[string]any)
				require.True(t, ok)
				require.Equal(t, "ciccio", saveFile["team1"])
				require.Equal(t, "pasticcio", saveFile["team2"])
			},
		},
		// Full example sample file from
		// https://developer.valvesoftware.com/wiki/Counter-Strike:_Global_Offensive_Game_State_Integration#Sample_Configuration_File
		{
			name:     "gamestate_integration_consolesample",
			fileName: "testdata/gamestate_integration_consolesample.cfg",
			want: func(got map[string]any, err error) {
				require.NoError(t, err)
				expected := map[string]any{
					"Console Sample v.1": map[string]any{
						"uri":       "http://127.0.0.1:3000",
						"timeout":   "5.0",
						"buffer":    "0.1",
						"throttle":  "0.5",
						"heartbeat": "60.0",
						"auth": map[string]any{ //nolint:gosec // the sample token is a public example from the Valve wiki
							"token": "CCWJu64ZV3JHDT8hZc",
						},
						"output": map[string]any{
							"precision_time":     "3",
							"precision_position": "1",
							"precision_vector":   "3",
						},
						"data": map[string]any{
							"provider":           "1",
							"map":                "1",
							"round":              "1",
							"player_id":          "1",
							"player_state":       "1",
							"player_weapons":     "1",
							"player_match_stats": "1",
						},
					},
				}
				require.Equal(t, expected, got)
			},
		},
		// Comments that start with "// my comment" are supported.
		// A broken comment like "/ my comment" is an error, never a partial map.
		{
			name:     "broken and unsupported comment",
			fileName: "testdata/broken_comment.cfg",
			want: func(got map[string]any, err error) {
				require.ErrorIs(t, err, ErrNotValidFormat)
				require.Nil(t, got)
			},
		},
		// Each level, even the first one, requires
		// surrounding curly braces. Intending is not enough.
		{
			name:     "no curly brace",
			fileName: "testdata/no_curly_brace.vdf",
			want: func(got map[string]any, err error) {
				require.Error(t, err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader, err := getReader(tt.fileName)
			require.NoError(t, err)
			p := NewParser(reader)
			got, err := p.Parse()
			tt.want(got, err)
		})
	}
}

func FuzzParser_Parse(f *testing.F) {
	testcases := []string{
		`"Example"
{
	"TimeNextStatsReport"      "1234567890"
	"ContentStatsID"           "-7123456789012345678"
}`,
		`"Root"
{
 "attr1"       "hey-ho"

 "attr2"       "ho-hey"
 "map1"
 {
   "foo"       "Q79v5tbar"
 }
 "data"
 {
   "val"       "1"
   "map"       "2"
   "player"    "3"
 }
}`,
		`"Root"
{
 attr1       "hey-ho"
 "attr2"       ho
 "map1"
 {
   "foo"       "Q79v5tbar"
 }
 "data"
 {
   "v\\al"       "1"
   "map"       "2"
   "pl\"ayer"    "3"
 }
}`,
		`"Root"
{
 attr1       "hey-ho"
 "attr2"       ho
 "map1"
 {
   // This is a comment
   "foo"       "Q79v5tbar"
 }
 "data"
 {
   "v\\al"       "1"
   "map"       "2"
   "pl\"ayer"    "3"
 }
}`,
		`"Root"
{
 attr1       "hey-ho"
 "attr2"       ho
 "map1"
 {
   // Comment line first
   // Comment line second
   "foo"       "Q79v5tbar"
   // Comment line third
 }
 "data"
 {
   "v\\al"       "1"
   "map"       "2"
   "pl\"ayer"    "3"
 }
}`,
		`// Root comment line
"Root"
{
 attr1       "hey-ho"
 "attr2"       ho
}`,
		`"Root"
{
 "map"
 {
   "attr1" "hello"
 }
 "map"
 {
   "attr2" "world"
 }
}`,
	}
	for _, tc := range testcases {
		f.Add(tc)
	}
	if entries, err := os.ReadDir("testdata"); err == nil {
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			b, err := os.ReadFile(filepath.Join("testdata", e.Name()))
			if err == nil {
				f.Add(string(b))
			}
		}
	}
	f.Fuzz(func(t *testing.T, input string) {
		if len(input) > 65536 {
			return
		}
		m, err := NewParser(strings.NewReader(input)).Parse()
		if err != nil {
			return
		}
		var buf bytes.Buffer
		if err := Write(&buf, m); err != nil {
			t.Fatalf("Write after successful Parse: %v\nmap: %#v", err, m)
		}
		written := buf.String()
		m2, err := NewParser(strings.NewReader(written)).Parse()
		if err != nil {
			t.Fatalf("Parse after Write: %v\nwritten:\n%s", err, written)
		}
		if !reflect.DeepEqual(m, m2) {
			t.Fatalf("round-trip map mismatch\nin: %#v\nout: %#v\nwritten:\n%s", m, m2, written)
		}
	})
}

// Ensure the parser returns on a comment at the end of the input or one holding a NUL byte.
func TestParser_CommentAtEOF(t *testing.T) {
	for _, s := range []string{
		"\"a\" {\n// last line",
		"\"a\" {\n//",
		"\"a\" {\n// x\x00y\n",
	} {
		_, _ = NewParser(strings.NewReader(s)).Parse()
	}
}
