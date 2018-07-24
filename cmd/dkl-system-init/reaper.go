package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"golang.org/x/sys/unix"
	"novit.nc/direktil/pkg/log"
)

var reapLock = sync.RWMutex{}

func handleChildren() {
	sigchld := make(chan os.Signal, 2048)
	signal.Notify(sigchld, syscall.SIGCHLD)

	// set us as a sub-reaper
	if err := unix.Prctl(unix.PR_SET_CHILD_SUBREAPER, 1, 0, 0, 0); err != nil {
		initLog.Taintf(log.Error, "reaper: failed to set myself a child sub-reaper: %v", err)
	}

	for range sigchld {
		reapChildren()
	}
}

func reapChildren() {
	reapLock.Lock()
	defer reapLock.Unlock()
	for {
		pid, err := syscall.Wait4(-1, nil, syscall.WNOHANG, nil)
		if err != nil && err != syscall.ECHILD {
			initLog.Taintf(log.Warning, "reaper: wait4 failed: %v", err)
			fmt.Printf("reaper: wait4 failed: %v\n", err)
			break
		}
		if pid <= 0 {
			break
		}
	}
}
