package main

import (
	"testing"

	"github.com/shoenig/test/must"
)

func TestExtract(t *testing.T) {
	t.Parallel()
	changelog := `# changelog

## 0.0.2

- second release

## 0.0.1

- sorts functions
- sorts struct fields
`

	t.Run("errors", func(t *testing.T) {
		t.Parallel()
		tests := map[string]struct {
			changelog string
			tag       string
			wantErr   string
		}{
			"empty section": {
				changelog: "## 0.0.2\n\n## 0.0.1\n\n- first\n",
				tag:       "v0.0.2",
				wantErr:   `section "## 0.0.2" in CHANGELOG.md is empty`,
			},
			"heading must match exactly": {
				changelog: "## 0.0.10\n\n- tenth\n",
				tag:       "v0.0.1",
				wantErr:   `CHANGELOG.md has no section "## 0.0.1"`,
			},
			"missing section": {
				changelog: changelog,
				tag:       "v9.9.9",
				wantErr:   `CHANGELOG.md has no section "## 9.9.9"`,
			},
		}
		for name, tt := range tests {
			t.Run(name, func(t *testing.T) {
				t.Parallel()
				_, err := extract(tt.changelog, tt.tag)
				must.EqError(t, err, tt.wantErr)
			})
		}
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		tests := map[string]struct {
			changelog string
			tag       string
			want      string
		}{
			"CRLF line endings": {
				changelog: "## 0.0.1\r\n\r\n- first\r\n",
				tag:       "v0.0.1",
				want:      "- first\n",
			},
			"last section": {
				changelog: changelog,
				tag:       "v0.0.1",
				want:      "- sorts functions\n- sorts struct fields\n",
			},
			"middle section": {
				changelog: changelog,
				tag:       "v0.0.2",
				want:      "- second release\n",
			},
			"tag without v prefix": {
				changelog: changelog,
				tag:       "0.0.2",
				want:      "- second release\n",
			},
		}
		for name, tt := range tests {
			t.Run(name, func(t *testing.T) {
				t.Parallel()
				have, err := extract(tt.changelog, tt.tag)
				must.NoError(t, err)
				must.Eq(t, tt.want, have)
			})
		}
	})
}
