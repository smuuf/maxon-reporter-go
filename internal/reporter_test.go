package internal

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
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

func TestSendPayloadSendsBodyToAllTargets(t *testing.T) {
	var bodies [2][]byte

	newCapturingServer := func(slot int) *httptest.Server {
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatal(err)
			}
			bodies[slot] = body
			w.WriteHeader(http.StatusOK)
		}))
	}

	server1 := newCapturingServer(0)
	defer server1.Close()
	server2 := newCapturingServer(1)
	defer server2.Close()

	reporter := Reporter{
		ConfigJson: Config{
			Target: []string{server1.URL, server2.URL},
		},
		HttpClient: &http.Client{},
	}

	reporter.sendPayload(PayloadType{"foo": "bar"})

	if len(bodies[0]) == 0 {
		t.Fatal("expected first target to receive a non-empty body")
	}
	if len(bodies[1]) == 0 {
		t.Fatal("expected second target to receive a non-empty body")
	}
	if string(bodies[0]) != string(bodies[1]) {
		t.Fatalf("expected both targets to receive the identical body, got %q and %q", bodies[0], bodies[1])
	}
}

func TestExecuteGathererTimesOutInsteadOfHanging(t *testing.T) {
	originalTimeout := gathererTimeout
	gathererTimeout = 100 * time.Millisecond
	defer func() { gathererTimeout = originalTimeout }()

	scriptPath := filepath.Join(t.TempDir(), "sleepy.sh")
	// "sleep" is not the last statement, so the shell must fork a real child
	// process to run it rather than exec-ing directly into it. This exercises
	// the case where killing the shell alone leaves the sleep grandchild
	// running - the gatherer timeout must kill the whole process group.
	script := "#!/bin/sh\nsleep 5\necho done=yes\n"
	if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	channel := make(chan *OrderedGathererResult, 1)
	wg.Add(1)

	start := time.Now()
	done := make(chan struct{})
	go func() {
		executeGatherer(&wg, channel, 0, scriptPath, map[string]string{})
		close(done)
	}()

	const maxWait = 1 * time.Second // well under the script's 5s sleep

	select {
	case <-done:
		if elapsed := time.Since(start); elapsed > maxWait {
			t.Fatalf("executeGatherer took %s to return; gatherer timeout was not enforced", elapsed)
		}
	case <-time.After(maxWait):
		t.Fatal("executeGatherer did not return within expected time; gatherer timeout was not enforced")
	}

	result := <-channel
	if result.exitError == nil {
		t.Fatal("expected exitError due to timeout, got nil")
	}
}
