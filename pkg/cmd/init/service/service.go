package initservices

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	plog "novit.nc/direktil/pkg/log"
)

var (
	delays = []time.Duration{
		1 * time.Second,
		2 * time.Second,
		4 * time.Second,
		8 * time.Second,
	}

	crashForgiveDelay = 10 * time.Minute
)

func Command() (c *cobra.Command) {
	c = &cobra.Command{
		Use:   "services",
		Short: "run user services",
		Run:   run,
	}

	return
}

func run(c *cobra.Command, args []string) {
	paths, err := filepath.Glob("/etc/direktil/services/*")

	if err != nil && !os.IsNotExist(err) {
		log.Fatal("failed to list services: ", err)
	}

	for _, path := range paths {
		stat, err := os.Stat(path)
		if err != nil {
			log.Fatalf("failed to stat %s: %v", path, err)
		}

		if stat.Mode()&0100 == 0 {
			// not executable
			continue
		}

		go runService(path)
	}

	select {}
}

func runService(svcPath string) {
	svc := filepath.Base(svcPath)

	logger := plog.Get(svc)
	plog.EnableFiles()

	n := 0
	for {
		lastStart := time.Now()

		cmd := exec.Command(svcPath)
		cmd.Stdout = logger
		cmd.Stderr = logger
		err := cmd.Run()

		if time.Since(lastStart) > crashForgiveDelay {
			n = 0
		}

		if err == nil {
			logger.Taintf(plog.Error, "service exited (%v), waiting %v", err, delays[n])
		} else {
			logger.Taintf(plog.Error, "service exited on error (%v), waiting %v", err, delays[n])
		}

		time.Sleep(delays[n])

		if n+1 < len(delays) {
			n++
		}
	}
}
