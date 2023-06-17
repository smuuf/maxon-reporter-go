package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/sevlyar/go-daemon.v0"
)

// Daemonizes reporter process.
// Returns true in the child (forked) process and false in the parent process.
func Daemonize() (bool, *daemon.Context) {
	ctx := &daemon.Context{
		PidFileName: filepath.Join(Settings.SelfDir, ".reporter.pid"),
		PidFilePerm: 0644,
		LogFileName: "ahoj",
	}

	runningProcess, _ := ctx.Search()
	if runningProcess != nil {
		fmt.Printf("Killing running reporter with PID %d ...\n", (*runningProcess).Pid)
		(*runningProcess).Signal(os.Interrupt)

		// This delay seems to fix some weird behavior of Reborn() if it's
		// called when the already running daemon process might not be
		// terminated yet.
		time.Sleep(time.Millisecond * 500)
	}

	child, _ := ctx.Reborn()
	return child == nil, ctx
}
