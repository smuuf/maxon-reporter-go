package internal

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
	"gopkg.in/sevlyar/go-daemon.v0"
)

// Daemonizes reporter process.
// Returns true in the child (forked) process and false in the parent process.
func Daemonize() (bool, *daemon.Context) {
	ctx := &daemon.Context{
		PidFileName: filepath.Join(Settings.SelfDir, ".reporter.pid"),
		PidFilePerm: 0644,
		LogFileName: filepath.Join(Settings.SelfDir, "reporter.log"),
		LogFilePerm: 0644,
	}

	// Set up log rotation
	log.SetOutput(&lumberjack.Logger{
		Filename:   ctx.LogFileName,
		MaxSize:    20, // MB
		MaxBackups: 1,
		MaxAge:     30, // days
		Compress:   false,
	})
	log.Println("Log rotation enabled.")

	runningProcess, _ := ctx.Search()
	if runningProcess != nil {
		fmt.Printf("Killing running reporter with PID %d ...\n", (*runningProcess).Pid)
		(*runningProcess).Signal(os.Interrupt)

		// This delay seems to fix some weird behavior of Reborn() if it's
		// called when the already running daemon process might not be
		// terminated yet.
		time.Sleep(time.Millisecond * 1000)
	}

	child, _ := ctx.Reborn()
	return child == nil, ctx
}
