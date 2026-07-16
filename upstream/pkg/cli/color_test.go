package cli

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestColorScheme(t *testing.T) {
	// Enable colors for testing
	t.Setenv("CLICOLOR_FORCE", "1")

	cs := NewColorScheme(true, true)

	tests := []struct {
		name     string
		function func(string) string
		input    string
		expected string
	}{
		{"Red", cs.Red, "test", red("test")},
		{"Green", cs.Green, "test", green("test")},
		{"Blue", cs.Blue, "test", blue("test")},
		{"Yellow", cs.Yellow, "test", yellow("test")},
		{"Magenta", cs.Magenta, "test", magenta("test")},
		{"Cyan", cs.Cyan, "test", cyan("test")},
		{"Gray", cs.Gray, "test", gray256("test")},
		{"Bold", cs.Bold, "test", bold("test")},
		{"Underline", cs.Underline, "test", underline("test")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.function(tt.input)
			assert.Equal(t, got, tt.expected, "ColorScheme.%s() = %v, want %v", tt.name, got, tt.expected)
		})
	}
}

func TestColorStatus(t *testing.T) {
	cs := NewColorScheme(true, true)

	tests := []struct {
		status   string
		expected string
	}{
		{"succeeded", cs.Green("succeeded")},
		{"failed", cs.Red("failed")},
		{"pipelineruntimeout", cs.Yellow("Timeout")},
		{"norun", cs.Dimmed("norun")},
		{"running", cs.Blue("running")},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			got := cs.ColorStatus(tt.status)
			assert.Equal(t, got, tt.expected, "ColorScheme.ColorStatus() = %v, want %v", got, tt.expected)
		})
	}
}

func TestEnvColorDisabled(t *testing.T) {
	tests := []struct {
		name     string
		noColor  string
		cliColor string
		expected bool
	}{
		{"NO_COLOR set", "1", "", true},
		{"CLICOLOR=0", "", "0", true},
		{"none set", "", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("NO_COLOR", tt.noColor)
			t.Setenv("CLICOLOR", tt.cliColor)
			assert.Equal(t, EnvColorDisabled(), tt.expected)
		})
	}
}

func TestEnvColorForced(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		{"unset", "", false},
		{"zero", "0", false},
		{"set", "1", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("CLICOLOR_FORCE", tt.value)
			assert.Equal(t, EnvColorForced(), tt.expected)
		})
	}
}

func TestIs256ColorSupported(t *testing.T) {
	tests := []struct {
		name      string
		term      string
		colorterm string
		expected  bool
	}{
		{"term 256color", "xterm-256color", "", true},
		{"term truecolor", "xterm", "truecolor", true},
		{"term 24bit", "", "24bit", true},
		{"plain term", "xterm", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("TERM", tt.term)
			t.Setenv("COLORTERM", tt.colorterm)
			assert.Equal(t, Is256ColorSupported(), tt.expected)
		})
	}
}

func TestColorSchemeMore(t *testing.T) {
	t.Setenv("CLICOLOR_FORCE", "1")
	cs := NewColorScheme(true, true)
	disabled := NewColorScheme(false, false)

	assert.Equal(t, cs.Orange("test"), orangeBold("test"))
	assert.Equal(t, disabled.Orange("test"), "test")

	assert.Equal(t, cs.Boldf("hello %s", "world"), bold("hello world"))
	assert.Equal(t, disabled.Bold("test"), "test")

	assert.Equal(t, cs.RedBold("test"), redBold("test"))
	assert.Equal(t, disabled.RedBold("test"), "test")
	assert.Equal(t, cs.Redf("hello %s", "world"), red("hello world"))

	assert.Equal(t, cs.Bullet(), "∙ ")
	assert.Equal(t, disabled.Bullet(), "")
	assert.Equal(t, cs.BulletSpace(), "  ")
	assert.Equal(t, disabled.BulletSpace(), "")

	assert.Equal(t, cs.Yellowf("hello %s", "world"), yellow("hello world"))
	assert.Equal(t, cs.Greenf("hello %s", "world"), green("hello world"))
	assert.Equal(t, disabled.Underline("test"), "test")

	assert.Equal(t, cs.Grayf("hello %s", "world"), gray256("hello world"))
	nonAnsi256 := NewColorScheme(true, false)
	assert.Equal(t, nonAnsi256.Gray("test"), gray("test"))
	assert.Equal(t, disabled.Gray("test"), "test")

	assert.Equal(t, cs.Magentaf("hello %s", "world"), magenta("hello world"))
	assert.Equal(t, disabled.Magenta("test"), "test")

	assert.Equal(t, cs.Cyanf("hello %s", "world"), cyan("hello world"))
	assert.Equal(t, disabled.Cyan("test"), "test")
	assert.Equal(t, cs.CyanBold("test"), cyanBold("test"))
	assert.Equal(t, disabled.CyanBold("test"), "test")

	assert.Equal(t, disabled.Blue("test"), "test")
	assert.Equal(t, cs.BlueBold("test"), blueBold("test"))
	assert.Equal(t, disabled.BlueBold("test"), "test")
	assert.Equal(t, cs.Bluef("hello %s", "world"), blue("hello world"))

	assert.Equal(t, cs.SuccessIcon(), green("✓"))
	assert.Equal(t, cs.InfoIcon(), blueBold("ℹ"))
	assert.Equal(t, cs.WarningIcon(), yellow("!"))
	assert.Equal(t, cs.FailureIcon(), red("X"))

	assert.Equal(t, cs.GreenBold("test"), greenBold("test"))
	assert.Equal(t, disabled.GreenBold("test"), "test")

	assert.Equal(t, cs.HyperLink("title", "href"), hyperLink("title", "href"))
	assert.Equal(t, disabled.HyperLink("title", "href"), "title")
}

func TestColorFromString(t *testing.T) {
	cs := NewColorScheme(true, true)
	tests := []struct {
		name  string
		input string
	}{
		{"bold", "Bold"},
		{"red", "RED"},
		{"yellow", "yellow"},
		{"green", "green"},
		{"gray", "gray"},
		{"magenta", "magenta"},
		{"cyan", "cyan"},
		{"blue", "blue"},
		{"unknown", "notacolor"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := cs.ColorFromString(tt.input)
			assert.Assert(t, fn != nil)
			assert.Assert(t, fn("test") != "")
		})
	}
}
