package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	Settings.SelfDir = "../example"
	config := LoadConfig(FindConfig())

	assert.Len(t, config.Target, 2)
	assert.Equal(t, "https://httpbingo.org/post", config.Target[0])
	assert.Equal(t, "https://localhost/report", config.Target[1])

	assert.Len(t, config.Env, 2)
	assert.Equal(t, "This env var is available in gatherers", config.Env["SOME_ENV_VAR_XYZ"])
	assert.Equal(t, "And this one too...", config.Env["ANOTHER_ENV_VAR_ABC"])

	assert.Len(t, config.Gatherers, 2)
	assert.Contains(t, config.Gatherers[0], "/gatherers/machine.sh")
	assert.Contains(t, config.Gatherers[1], "/gatherers/some_script.py")

	assert.NotEmpty(t, config.Payload)
}
