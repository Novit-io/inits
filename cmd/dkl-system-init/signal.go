package main

import (
	"os"
	"os/signal"
	"syscall"
)

func handleSignals() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGPWR)

	for sig := range c {
		switch sig {
		case syscall.SIGPWR:
			poweroff()
		}
	}
}
