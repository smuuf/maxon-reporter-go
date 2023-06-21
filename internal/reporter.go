package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"sync"
	"time"

	"github.com/mohae/deepcopy"
)

// Recursively iterates over a payload template and expands variables and
// expressions in all of the string values present. The result is then returned.
func buildPayload(
	template PayloadType,
	vars EvalVariables,
) PayloadType {

	for k, v := range template {
		switch v := v.(type) {
		case string:
			_v, err := expandExpressions(v, vars)

			// Replace the value only if there was not an error.
			if err == nil {
				template[k] = _v
			}
		case float64:
			template[k] = v
		case StringKeyMap:
			template[k] = buildPayload(v, vars)
		case []interface{}:
			// Slice/array? Iterate over its items (we assume the items are
			// maps) and process each one of them).
			for i, sub_v := range v {
				v[i] = buildPayload(sub_v.(StringKeyMap), vars)
			}
		default:
			ErrorExit("Building payload", fmt.Sprintf("encountered unexpected payload template value of type '%s'", v))
		}
	}

	return template

}

func executeGatherer(
	wg *sync.WaitGroup,
	channel chan<- *OrderedGathererResult,
	index int,
	gathererPath string,
) {
	defer (*wg).Done()

	println("Executing gatherer:", TryMakingRelativePath(gathererPath))
	result := exec.Command(gathererPath)
	stdout, err := result.Output()

	channel <- &OrderedGathererResult{
		index:     index,
		gatherer:  gathererPath,
		exitError: err,
		data:      readIniValues(stdout),
	}
}

func (r *Reporter) sendPayload(payload PayloadType) {

	jsonPayload, err := json.Marshal(payload)
	FatalExitOnError(err)

	if Settings.VerboseMode {
		pretty, err := json.MarshalIndent(payload, "", "  ")
		FatalExitOnError(err)
		println("Payload:")
		println(string(pretty))
	}

	bytes := bytes.NewBuffer(jsonPayload)
	userAgent := fmt.Sprintf("maxon-reporter[go][%s]", ReporterVersion)

	for _, target := range r.ConfigJson.Target {

		request, err := http.NewRequest("POST", target, bytes)
		FatalExitOnError(err)

		request.Header.Set("User-Agent", userAgent)
		request.Header.Set("Content-Type", "application/json; charset=UTF-8")

		fmt.Printf("Sending payload to: %s ", target)
		response, err := r.HttpClient.Do(request)
		if err != nil {
			fmt.Printf("[Error: %s]\n", err.Error())
		} else {
			fmt.Printf("[%s]\n", response.Status)
			defer response.Body.Close()
		}
	}
}

func (r *Reporter) Single() {
	var wg sync.WaitGroup
	gatherers := r.ConfigJson.Gatherers
	// Prepare empty slice for storing results of gatherers. We do this to keep
	// track of their order (gatherers are executed asynchronously).
	// This slice will then be used to build the final single StringMap
	// containing all results of all gatherers (and because we keep information
	// about the order we get predictable behavior when two gatherers return
	// values under indentital key - the latter will overwrite the former.)
	results := make([]StringMap, len(gatherers))

	channel := make(chan *OrderedGathererResult)

	// Run gatherers asynchronously in goroutines.
	for index, gatherer := range gatherers {
		wg.Add(1)
		go executeGatherer(&wg, channel, index, gatherer)
	}

	// Wait for all goroutines to finish.
	// Why does this have to be in a goroutine? This is the best explanation
	// I've found so far:
	// > wg.Wait will only unblock once all the goroutine return. But as
	// > mentioned, all the goroutines are blocked at channel send.
	// > When you use a separate goroutine to wait, code execution actually
	// > reaches the range loop on queue.
	// See https://stackoverflow.com/a/46560572/1285669
	go func() {
		wg.Wait()
		close(channel)
	}()

	processResults(channel, &results)
	finalResult := MergeResults(results)

	payload := buildPayload(
		deepcopy.Copy(r.ConfigJson.Payload).(PayloadType),
		finalResult,
	)

	r.sendPayload(payload)
}

func (r *Reporter) Run() {
	for {
		// Wait first - when the Maxon Reporter is executed, it's initial
		// "gathering" is called first via Reporter.single(), which happens even
		// before daemonization of the Reporter. Because of that when this
		// Reporter.run() is then called, it's we actually have our first
		// "gathering" done and it makes sense to wait at this point.
		time.Sleep(10 * time.Second)

		// Make empty line between gatherings, for better readability of
		// Reporter's CLI output.
		println()

		r.Single()
	}
}

func processResults(channel chan *OrderedGathererResult, results *[]StringMap) {
	// Gather results of all gatherers (while keeping track of their original
	// order).
	for result := range channel {
		if result.exitError != nil {
			fmt.Printf("Gatherer %s exited with non-zero code: ", TryMakingRelativePath(result.gatherer))
			if exitErr, ok := result.exitError.(*exec.ExitError); ok {
				println(exitErr.ExitCode())
			} else {
				println("Unexpected error when processing exit code:", result.exitError.Error())
			}
		} else {
			(*results)[result.index] = result.data
		}
	}
}
