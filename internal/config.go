package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tidwall/jsonc"
)

var configSearchPaths = []string{
	"/config.json",
	"/config/config.json",
}

// Build Config struct from JSON data passed as bytes.
// The relative
func buildConfigFromJson(jsonBytes []byte, baseDir string) Config {
	c := Config{}

	if err := json.Unmarshal(jsonc.ToJSON(jsonBytes), &c); err != nil {
		ErrorExit("Parsing config", err.Error())
		return c
	}

	var err error
	var absPath string
	for i, gathererPath := range c.Gatherers {
		// Absolute gatherer paths will be kept as-is, while relative gatherer
		// paths will be made absolute by basing them upon provided baseDir.
		if filepath.IsAbs(gathererPath) {
			absPath = gathererPath
		} else {
			absPath, err = filepath.Abs(filepath.Join(baseDir, gathererPath))
			FatalExitOnError(err)
		}

		c.Gatherers[i] = absPath
	}

	return c
}

func validateConfig(c Config) error {
	var err error

	for _, target := range c.Target {
		if !strings.HasPrefix(target, "https://") && !strings.HasPrefix(target, "http://") {
			return fmt.Errorf("target URL '%s' is not an acceptable URL", target)
		}
	}

	for _, gatherer := range c.Gatherers {
		if !IsExistingFile(gatherer) {
			return fmt.Errorf("gatherer '%s' not found", gatherer)
		}
	}

	return err
}

// Finds some config file relative to main executable.
func FindConfig() string {
	var tried []string

	for _, path := range configSearchPaths {
		path, err := filepath.Abs(Settings.SelfDir + path)
		FatalExitOnError(err)

		tried = append(tried, path)
		if IsExistingFile(path) {
			println("Found config:", path)
			return path
		}
	}

	ErrorExit(fmt.Sprintf(
		"cannot find 'config.json' file, tried:\n%s",
		strings.Join(tried, "\n"),
	))
	return ""
}

func LoadConfig(configPath string) Config {
	// Make the config path absolute.
	configPath, err := filepath.Abs(configPath)
	FatalExitOnError(err)
	jsonBytes, err := os.ReadFile(configPath)
	FatalExitOnError(err)

	config := buildConfigFromJson(jsonBytes, filepath.Dir(configPath))
	if err := validateConfig(config); err != nil {
		ErrorExit("Config validation", err.Error())
	}

	return config
}
