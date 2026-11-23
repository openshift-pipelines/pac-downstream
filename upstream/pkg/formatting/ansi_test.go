package formatting

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestStripANSI(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "plain text unchanged",
			input: "cmd/main.go:42:1: broken",
			want:  "cmd/main.go:42:1: broken",
		},
		{
			name:  "basic colors",
			input: "\x1b[31merror:\x1b[0m broken",
			want:  "error: broken",
		},
		{
			name:  "bold and 256 colors",
			input: "\x1b[1;38;5;208mwarning\x1b[m done",
			want:  "warning done",
		},
		{
			name:  "nilaway output from issue",
			input: "/workspace/source/pkg/params/run.go:58:16: \x1b[31merror: \x1b[0mPotential nil panic detected. literal \x1b[95m`nil`\x1b[0m returned",
			want:  "/workspace/source/pkg/params/run.go:58:16: error: Potential nil panic detected. literal `nil` returned",
		},
		{
			name:  "cursor and erase sequences",
			input: "\x1b[2K\x1b[1Aprogress 100%",
			want:  "progress 100%",
		},
		{
			name:  "osc hyperlink terminated by st",
			input: "see \x1b]8;;https://example.com\x1b\\docs\x1b]8;;\x1b\\ here",
			want:  "see docs here",
		},
		{
			name:  "osc title terminated by bel",
			input: "\x1b]0;title\x07text",
			want:  "text",
		},
		{
			name:  "two character escape",
			input: "\x1bMline",
			want:  "line",
		},
		{
			name:  "utf8 preserved",
			input: "\x1b[32m🚀 déployé\x1b[0m",
			want:  "🚀 déployé",
		},
		{
			name:  "multiline",
			input: "\x1b[31ma\x1b[0m\n\x1b[32mb\x1b[0m\n",
			want:  "a\nb\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, StripANSI(tt.input))
		})
	}
}
