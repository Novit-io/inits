package initboot

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"novit.nc/direktil/pkg/config"

	"novit.nc/direktil/inits/pkg/sys"
	"novit.nc/direktil/inits/pkg/vars"
)

func setupLVM() {
	if !dmInProc() {
		sys.MustRun("modprobe", "dm-mod")
	}

	// start lvmetad
	sys.Mkdir("/run/lvm", 0700)
	sys.Mkdir("/run/lock/lvm", 0700)
	sys.Run("lvmetad")

	sys.WaitFile("/run/lvm/lvmetad.socket", time.After(30*time.Second))

	// scan devices
	sys.Run("lvm", "pvscan")
	sys.Run("lvm", "vgscan", "--mknodes")
	sys.Run("lvm", "vgchange", "--sysinit", "-a", "ly")

	cfg := sys.Config()

	// setup storage
	log.Print("checking storage")
	if err := exec.Command("vgdisplay", "storage").Run(); err != nil {
		log.Print("- creating VG storage")
		setupVG(vars.BootArgValue("storage", cfg.Storage.UdevMatch))
	}

	for _, name := range cfg.Storage.RemoveVolumes {
		dev := "/dev/storage/" + name

		if _, err := os.Stat(dev); os.IsNotExist(err) {
			continue

		} else if err != nil {
			log.Fatal("failed to stat ", dev, ": ", err)
		}

		log.Print("- removing LV ", name)
		cmd := exec.Command("lvremove", "-f", "storage/"+name)
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			log.Fatal("failed to remove LV ", name)
		}
	}

	// setup volumes
	for _, volume := range cfg.Storage.Volumes {
		if err := exec.Command("lvdisplay", "storage/"+volume.Name).Run(); err != nil {
			log.Print("- creating LV ", volume.Name)
			setupLV(volume)
		}

		dev := "/dev/storage/" + volume.Name

		sys.WaitFile(dev, time.After(30*time.Second))

		log.Printf("checking filesystem on %s", dev)
		sys.MustRun("fsck", "-p", dev)

		sys.Mount(dev, volume.Mount.Path, volume.FS,
			syscall.MS_NOATIME|syscall.MS_RELATIME,
			volume.Mount.Options)
	}
}

func dmInProc() bool {
	for _, f := range []string{"devices", "misc"} {
		c, err := ioutil.ReadFile("/proc/" + f)
		if err != nil {
			log.Fatalf("failed to read %s: %v", f, err)
		}
		if !bytes.Contains(c, []byte("device-mapper")) {
			return false
		}
	}
	return true
}

func setupVG(udevMatch string) {
	const pDevName = "DEVNAME="

	dev := ""
	try := 0

retry:
	paths, err := filepath.Glob("/sys/class/block/*")
	if err != nil {
		log.Fatal("failed to list block devices: ", err)
	}

	for _, path := range paths {
		// ignore loop devices
		if strings.HasPrefix("loop", filepath.Base(path)) {
			continue
		}

		// fetch udev informations
		out, err := exec.Command("udevadm", "info", "-q", "property", path).CombinedOutput()
		if err != nil {
			log.Printf("WARNING: udev query of %q failed: %v\n%s", path, err, string(out))
			continue
		}

		propertyLines := strings.Split(strings.TrimSpace(string(out)), "\n")

		devPath := ""
		matches := false

		for _, line := range propertyLines {
			if strings.HasPrefix(line, pDevName) {
				devPath = line[len(pDevName):]
			}

			if matched, err := filepath.Match(udevMatch, line); err != nil {
				log.Fatalf("FATAL: invalid match: %q: %v", udevMatch, err)

			} else if matched {
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
			log.Fatal("FATAL: storage device not found after 30s: ", udevMatch)
		}
		goto retry
	}

	log.Print("found storage device at ", dev)

	sys.MustRun("pvcreate", dev)
	sys.MustRun("vgcreate", "storage", dev)
}

func setupLV(volume config.VolumeDef) {
	if volume.Extents != "" {
		sys.MustRun("lvcreate", "-l", volume.Extents, "-n", volume.Name, "storage")
	} else {
		sys.MustRun("lvcreate", "-L", volume.Size, "-n", volume.Name, "storage")
	}

	// wait the device link
	devPath := "/dev/storage/" + volume.Name
	sys.WaitFile(devPath, time.After(30*time.Second))

	args := make([]string, 0)

	switch volume.FS {
	case "btrfs":
		args = append(args, "-f")
	case "ext4":
		args = append(args, "-F")
	}

	sys.MustRun("mkfs."+volume.FS, append(args, devPath)...)
}
