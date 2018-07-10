package main

import (
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sys/unix"
	"novit.nc/direktil/pkg/log"
)

func handleChildren() {
	// set us as a sub-reaper
	if err := unix.Prctl(unix.PR_SET_CHILD_SUBREAPER, 1, 0, 0, 0); err != nil {
		initLog.Taintf(log.Error, "reaper: failed to set myself a child sub-reaper: %v", err)
	}

	sigchld := make(chan os.Signal, 2048)
	signal.Notify(sigchld, syscall.SIGCHLD)

	for range sigchld {
		reapChildren()
	}
}

func reapChildren() {
	for {
		pid, err := syscall.Wait4(-1, nil, syscall.WNOHANG, nil)
		if err != nil {
			if err == unix.ECHILD {
				break
			}
			initLog.Taintf(log.Warning, "reaper: wait4 failed: %v", err)
		}
		if pid <= 0 {
			break
		}
	}
}
