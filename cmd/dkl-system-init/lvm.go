package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"novit.nc/direktil/pkg/config"
	"novit.nc/direktil/pkg/log"
)

const (
	pDevName = "DEVNAME="
)

func init() {
	go services.WaitPath("/run/lvm/lvmetad.socket")

	services.Register(
		&CommandService{
			Name:    "lvmetad",
			Restart: StdRestart,
			Needs:   []string{"service:devfs"},
			Command: []string{"lvmetad", "-f"},
			PreExec: func() error {
				mkdir("/run/lvm", 0700)
				mkdir("/run/lock/lvm", 0700)

				if !dmInProc() {
					run("modprobe", "dm-mod")
				}

				return nil
			},
		},
		&CommandService{
			Name:  "lvm",
			Needs: []string{"file:/run/lvm/lvmetad.socket"},
			Command: []string{"/bin/sh", "-c", `set -ex
/sbin/lvm pvscan
/sbin/lvm vgscan --mknodes
/sbin/lvm vgchange --sysinit -a ly
`},
		},
	)
}

func isDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		fatal("failed to query ", path, ": ", err)
	}

	return s.IsDir()
}

func dmInProc() bool {
	for _, f := range []string{"devices", "misc"} {
		c, err := ioutil.ReadFile("/proc/" + f)
		if err != nil {
			fatalf("failed to read %s: %v", f, err)
		}
		if !bytes.Contains(c, []byte("device-mapper")) {
			return false
		}
	}
	return true
}

func setupVG(udevMatch string) {
	dev := ""
	try := 0

retry:
	paths, err := filepath.Glob("/sys/class/block/*")
	if err != nil {
		fatal("failed to list block devices: ", err)
	}

	for _, path := range paths {
		// ignore loop devices
		if strings.HasPrefix("loop", filepath.Base(path)) {
			continue
		}

		// fetch udev informations
		out, err := exec.Command("udevadm", "info", "-q", "property", path).CombinedOutput()
		if err != nil {
			initLog.Taintf(log.Warning, "udev query of %q failed: %v\n%s", path, err, string(out))
			continue
		}

		propertyLines := strings.Split(strings.TrimSpace(string(out)), "\n")

		devPath := ""
		matches := false

		for _, line := range propertyLines {
			if strings.HasPrefix(line, pDevName) {
				devPath = line[len(pDevName):]
			}

			if line == udevMatch {
				matches = true
			}

			if devPath != "" && matches {
				break
			}
		}

		if devPath != "" && matches {
			dev = devPath
			break
		}
	}

	if dev == "" {
		time.Sleep(1 * time.Second)
		try++
		if try > 30 {
			fatal("storage device not found after 30s, failing.")
		}
		goto retry
	}

	initLog.Taint(log.Info, "found storage device at ", dev)

	run("pvcreate", dev)
	run("vgcreate", "storage", dev)
}

func setupLV(volume config.VolumeDef) {
	if volume.Extents != "" {
		run("lvcreate", "-l", volume.Extents, "-n", volume.Name, "storage")
	} else {
		run("lvcreate", "-L", volume.Size, "-n", volume.Name, "storage")
	}

	// wait the device link
	devPath := "/dev/storage/" + volume.Name
	for i := 0; i < 300; i++ {
		_, err := os.Stat(devPath)
		if err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	args := make([]string, 0)

	switch volume.FS {
	case "btrfs":
		args = append(args, "-f")
	case "ext4":
		args = append(args, "-F")
	}

	run("mkfs."+volume.FS, append(args, devPath)...)
}
