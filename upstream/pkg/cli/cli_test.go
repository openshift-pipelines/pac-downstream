package cli

import (
	"testing"

	"github.com/AlecAivazis/survey/v2"
	"gotest.tools/v3/assert"
)

func TestNewAskopts(t *testing.T) {
	opt := &survey.AskOptions{}
	err := NewAskopts(opt)
	assert.NilError(t, err)
	assert.Assert(t, opt.Stdio.In != nil)
	assert.Assert(t, opt.Stdio.Out != nil)
	assert.Assert(t, opt.Stdio.Err != nil)
}

func TestNewCliOptions(t *testing.T) {
	opts := NewCliOptions()
	assert.Assert(t, opts != nil)
	assert.Assert(t, opts.AskOpts != nil)
}
