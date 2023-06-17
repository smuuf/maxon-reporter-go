package internal

import (
	"net/http"
	"testing"
)

func TestReporter(t *testing.T) {

	Settings.SelfDir = "../example"
	config := LoadConfig(FindConfig())

	reporter := Reporter{
		ConfigJson: config,
		HttpClient: &http.Client{},
	}

	reporter.Single()
}
