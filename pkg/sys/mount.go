package sys

import (
	"log"
	"os"
	"syscall"
)

func Mount(source, target, fstype string, flags uintptr, data string) {
	if _, err := os.Stat(target); os.IsNotExist(err) {
		Mkdir(target, 0755)
	}

	if err := syscall.Mount(source, target, fstype, flags, data); err != nil {
		log.Fatalf("FATAL: mount %q %q -t %q -o %q failed: %v", source, target, fstype, data, err)
	}

	log.Printf("mounted %q", target)
}
