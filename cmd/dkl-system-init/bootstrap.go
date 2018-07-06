package main

import (
	"log"
	"os"
	"syscall"
)

func bootstrap() {
	mount("proc", "/proc", "proc", 0, "")
	mount("sys", "/sys", "sysfs", 0, "")
	mount("dev", "/dev", "devtmpfs", syscall.MS_NOSUID, "mode=0755,size=10M")
	mount("run", "/run", "tmpfs", 0, "")

	mount("/run", "/var/run", "", syscall.MS_BIND, "")

	mkdir("/run/lock", 0775)
	log.Print("/run/lock: correcting owner")
	if err := os.Chown("/run/lock", 0, 14); err != nil {
		fatal(err)
	}
}

func mount(source, target, fstype string, flags uintptr, data string) {
	if _, err := os.Stat(target); os.IsNotExist(err) {
		mkdir(target, 0755)
	}

	if err := syscall.Mount(source, target, fstype, flags, data); err != nil {
		fatalf("mount %q %q -t %q -o %q failed: %v", source, target, fstype, data, err)
	}
	log.Printf("mounted %q", target)
}
