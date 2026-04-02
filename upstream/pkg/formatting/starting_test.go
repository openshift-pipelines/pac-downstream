package formatting

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestMessageTemplateMakeTemplate(t *testing.T) {
	mt := MessageTemplate{
		PipelineRunName: "test-pipeline",
		Namespace:       "test-namespace",
		ConsoleName:     "test-console",
		ConsoleURL:      "https://test-console-url.com",
		TknBinary:       "test-tkn",
		TknBinaryURL:    "https://test-tkn-url.com",
		FailureSnippet:  "such a failure",
	}

	tests := []struct {
		name    string
		mt      MessageTemplate
		msg     string
		want    string
		wantErr bool
	}{
		{
			name: "Test MakeTemplate",
			mt:   mt,
			msg:  "Starting Pipelinerun {{.Mt.PipelineRunName}} in namespace {{.Mt.Namespace}}",
			want: "Starting Pipelinerun test-pipeline in namespace test-namespace",
		},
		{
			name:    "Error MakeTemplate",
			mt:      mt,
			msg:     "Starting Pipelinerun {{.Mt.PipelineRunName}} in namespace {{.FOOOBAR }}",
			wantErr: true,
		},
		{
			name: "Failure template",
			mt:   mt,
			msg:  "I am {{ .Mt.FailureSnippet }}",
			want: "I am such a failure",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.mt.MakeTemplate(tt.msg)
			if tt.wantErr {
				assert.Assert(t, err != nil)
				return
			}
			assert.NilError(t, err)
			assert.Equal(t, got, tt.want)
		})
	}
}
