package vdf

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParser_DuplicateSectionsMergeRecursively(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want map[string]any
	}{
		{
			name: "localconfig nested sections merge",
			in: `"UserLocalConfigStore"
{
	"Software"
	{
		"Valve"
		{
			"Steam"
			{
				"apps"
				{
					"220" { "LastPlayed" "1" }
				}
			}
		}
	}
	"Software"
	{
		"Valve"
		{
			"Steam"
			{
				"apps"
				{
					"400" { "LastPlayed" "2" }
				}
				"Language" "english"
			}
		}
	}
}`,
			want: map[string]any{"UserLocalConfigStore": map[string]any{"Software": map[string]any{"Valve": map[string]any{"Steam": map[string]any{
				"apps": map[string]any{
					"220": map[string]any{"LastPlayed": "1"},
					"400": map[string]any{"LastPlayed": "2"},
				},
				"Language": "english",
			}}}}},
		},
		{
			name: "libraryfolders duplicate entry, later scalar wins",
			in: `"libraryfolders"
{
	"0" { "path" "/a" "apps" { "220" "10" } }
	"0" { "path" "/b" "apps" { "400" "20" } }
}`,
			want: map[string]any{"libraryfolders": map[string]any{"0": map[string]any{
				"path": "/b",
				"apps": map[string]any{"220": "10", "400": "20"},
			}}},
		},
		{
			name: "string key then section of the same name: section wins",
			in:   `"root" { "apps" "none" "apps" { "220" "1" } }`,
			want: map[string]any{"root": map[string]any{"apps": map[string]any{"220": "1"}}},
		},
		{
			name: "section then string key of the same name: string wins",
			in:   `"root" { "apps" { "220" "1" } "apps" "none" }`,
			want: map[string]any{"root": map[string]any{"apps": "none"}},
		},
		{
			name: "section then scalar inside a merged section: later kind wins",
			in:   `"root" { "a" { "x" { "k" "v" } } "a" { "x" "flat" } }`,
			want: map[string]any{"root": map[string]any{"a": map[string]any{"x": "flat"}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewParser(strings.NewReader(tt.in)).Parse()
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}
