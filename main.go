package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"reporter/internal"
	_ "testing"

	"github.com/akamensky/argparse"
)

func init() {
	// Remove prefix from log message output.
	log.SetFlags(log.Lmsgprefix)

	// Print Maxon Reporter text header.
	internal.PrintHeader()
}

// Separate from init() as we don't want this code (i.e. parsing CLI arguments)
// to be executed when initializing testing/benchmarks and when this logic was
// placed inside Go's native init(), the benchmarks wouldn't run when executed
// via "go test --bench ." for some reason...
func initialize() {
	// Parse CLI arguments.
	parser := argparse.NewParser("reporter", "Maxon Reporter")
	configJsonPath := parser.String("c", "config", &argparse.Options{Required: false, Help: "Path to config.json"})
	verboseMode := parser.Flag("v", "verbose", &argparse.Options{Required: false, Help: "Enable verbose mode", Default: false})
	justTry := parser.Flag("t", "try", &argparse.Options{Required: false, Help: "Do a single run and exit", Default: false})
	foregroundMode := parser.Flag("f", "foreground", &argparse.Options{Required: false, Help: "Run in foreground without daemonization", Default: false})
	// ... and actually parse the arguments.
	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Print(parser.Usage(err))
		os.Exit(1)
	}

	settings := &internal.Settings

	// Determine absolute path to directory of the binary.
	selfPath, err := os.Executable()
	internal.FatalExitOnError(err)
	selfDir, err := filepath.Abs(path.Dir(selfPath))
	internal.FatalExitOnError(err)
	settings.SelfDir = selfDir

	settings.DaemonMode = true
	settings.VerboseMode = *verboseMode
	settings.ForegroundMode = *foregroundMode
	settings.ConfigJsonPath = *configJsonPath
	settings.JustTry = *justTry

	// Force some settings when In "just try" mode.
	if settings.JustTry {
		settings.DaemonMode = false
		settings.VerboseMode = true
	}

	// Force some settings when In "just try" mode.
	if settings.ForegroundMode {
		settings.DaemonMode = false
	}

	// Force some of the settings for dev builds.
	if internal.ReporterDevFlag == "1" {
		// Because the built dev binary is in different directory than
		// example config, we'll force the example config to be used (unless
		// overridden with CLI argument). This makes manual testing during
		// development easier.
		settings.ConfigJsonPath = filepath.Join(selfDir, "../example/config.json")
	}
}

func main() {
	initialize()

	settings := &internal.Settings

	// If the config JSON file path was not specified via CLI argument, we'll
	// try to find the config file ourselves.
	if settings.ConfigJsonPath == "" {
		settings.ConfigJsonPath = internal.FindConfig()
	}

	// Do this before daemonization, so that errors in config are visible soon.
	loadedConfig := internal.LoadConfig(settings.ConfigJsonPath)

	reporter := internal.Reporter{
		ConfigJson: loadedConfig,
		HttpClient: &http.Client{},
	}

	reporter.Single()

	if settings.JustTry {
		return
	}

	if settings.DaemonMode {
		weAreTheDaemon, ctx := internal.Daemonize()

		if weAreTheDaemon {
			// Do our best to remove the PID file when terminating the child
			// process.
			defer func() {
				if err := ctx.Release(); err != nil {
					log.Printf("Unable to release PID file: %s", err.Error())
				}
			}()
		} else {
			// We're the parent process - we'll tell the client the reporter
			// is being daemonized and then exit.
			fmt.Println("Daemonizing ...")
			os.Exit(0)
		}
	}

	reporter.Run()
}
