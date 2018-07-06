package main

import (
	"os"

	"novit.nc/direktil/pkg/color"
	"novit.nc/direktil/pkg/log"
)

const (
	endOfInitMessage = `
.---- END OF INIT -----.
| init process failed. |
 ----------------------
`
)

func fatal(v ...interface{}) {
	initLog.Taint(log.Fatal, v...)
	os.Stderr.Write([]byte(color.Red + endOfInitMessage + color.Reset))

	services.SetFlag("boot-failed")
	endOfProcess()
}

func fatalf(pattern string, v ...interface{}) {
	initLog.Taintf(log.Fatal, pattern, v...)
	os.Stderr.Write([]byte(color.Red + endOfInitMessage + color.Reset))

	services.SetFlag("boot-failed")
	endOfProcess()
}

func endOfProcess() {
	select {}
}
