package main

import (
	"bufio"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"syscall"
)

const (
	msSysfs = syscall.MS_NODEV | syscall.MS_NOEXEC | syscall.MS_NOSUID
)

func init() {
	services.Register(
		&CommandService{
			Name: "kmod-static-nodes",
			PreExec: func() error {
				mkdir("/run/tmpfiles.d", 0755)
				return nil
			},
			Command: []string{"kmod", "static-nodes", "--format=tmpfiles",
				"--output=/run/tmpfiles.d/kmod.conf"},
		},
		&Oneshot{
			Name: "devfs",
			Func: func() error {
				for _, mount := range []struct {
					fstype string
					target string
					mode   os.FileMode
					flags  uintptr
					data   string
					source string
				}{
					{"mqueue", "/dev/mqueue", 01777, syscall.MS_NODEV, "", "mqueue"},
					{"devpts", "/dev/pts", 0755, 0, "gid=5", "devpts"},
					{"tmpfs", "/dev/shm", 01777, syscall.MS_NODEV, "mode=1777", "shm"},
					{"tmpfs", "/sys/fs/cgroup", 0755, 0, "mode=755,size=10m", "cgroup"},
				} {
					initLog.Print("mounting ", mount.target)

					flags := syscall.MS_NOEXEC | syscall.MS_NOSUID | mount.flags

					mkdir(mount.target, mount.mode)
					err := syscall.Mount(mount.source, mount.target, mount.fstype, flags, mount.data)
					if err != nil {
						fatalf("mount failed: %v", err)
					}
				}

				// mount cgroup controllers
				for line := range readLines("/proc/cgroups") {
					parts := strings.Split(line, "\t")
					name, enabled := parts[0], parts[3]

					if enabled != "1" {
						continue
					}

					initLog.Print("mounting cgroup fs for controller ", name)

					mp := "/sys/fs/cgroup/" + name
					mkdir(mp, 0755)
					mount(name, mp, "cgroup", msSysfs, name)
				}

				if err := ioutil.WriteFile("/sys/fs/cgroup/memory/memory.use_hierarchy",
					[]byte{'1'}, 0644); err != nil {
					initLog.Print("failed to enable use_hierarchy in memory cgroup: ", err)
				}

				return nil
			},
		},
		&CommandService{
			Name:    "dmesg",
			Command: []string{"dmesg", "-n", "warn"},
		},
	)
}

func readLines(path string) chan string {
	f, err := os.Open(path)
	if err != nil {
		fatalf("failed to open %s: %v", path, err)
	}

	bf := bufio.NewReader(f)

	ch := make(chan string, 1)

	go func() {
		defer f.Close()
		defer close(ch)

		for {
			line, err := bf.ReadString('\n')
			if err == io.EOF {
				break
			} else if err != nil {
				fatalf("error while reading %s: %v", path, err)
			}
			ch <- line[:len(line)-1]
		}
	}()

	return ch
}
