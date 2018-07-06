package main

import (
	"bufio"
	"io"
	"os"
	"syscall"
)

func listenInitctl() {
	const f = "/run/initctl"

	if err := syscall.Mkfifo(f, 0700); err != nil {
		fatal("can't create "+f+": ", err)
	}

	for {
		func() {
			fifo, err := os.Open(f)
			if err != nil {
				fatal("can't open "+f+": ", err)
			}
			defer fifo.Close()

			r := bufio.NewReader(fifo)

			for {
				s, err := r.ReadString('\n')
				if err == io.EOF {
					break
				}
				if err != nil {
					initLog.Print(f+": read error: ", err)
				}

				switch s {
				case "prepare-shutdown\n":
					prepareShutdown()
				case "poweroff\n", "shutdown\n":
					poweroff()
				case "reboot\n":
					reboot()
				default:
					initLog.Printf(f+": unknown command: %q", s)
				}
			}
		}()
	}
}
