package sys

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

func Run(cmd string, args ...string) (err error) {
	c := exec.Command(cmd, args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	if err = c.Run(); err != nil {
		log.Printf("command %s %q failed: %v", cmd, args, err)
	}
	return
}

func MustRun(cmd string, args ...string) {
	if err := Run(cmd, args...); err != nil {
		log.Fatal("FATAL: mandatory command did not succeed")
	}
}

func Mkdir(dir string, mode os.FileMode) {
	if err := os.MkdirAll(dir, mode); err != nil {
		log.Fatalf("FATAL: mkdir %q failed: %v", dir, err)
	}
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("WARNING: failed to stat %q, assuming not exist: %v", path, err)
		}
		return false
	}
	return true
}

func WaitFile(path string, timeout <-chan time.Time) {
	if FileExists(path) {
		return
	}

	watch, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal("FATAL: fsnotify: failed to create: ", err)
	}

	defer watch.Close()

	dir := filepath.Dir(path)

	if err = watch.Add(dir); err != nil {
		log.Fatalf("FATAL: fsnotify: failed to add %s: %v", dir, err)
	}

	go func() {
		for err := range watch.Errors {
			log.Fatal("FATAL: fsnotify: error: ", err)
		}
	}()

	timedOut := false
	for !timedOut {
		select {
		case <-watch.Events:
			// skip

		case <-timeout:
			timedOut = true
		}

		if FileExists(path) {
			return
		}
	}

	log.Fatal("FATAL: timed out waiting for ", path)
}
