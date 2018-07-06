package main

import (
	"io/ioutil"
	"os"
)

func init() {
	services.Register(
		&CommandService{
			Name:    "udev",
			Restart: StdRestart,
			Needs:   []string{"service:devfs", "service:dmesg", "service:lvm"},
			PreExec: func() error {
				if _, err := os.Stat("/proc/net/unix"); os.IsNotExist(err) {
					run("modprobe", "unix")
				}
				if _, err := os.Stat("/proc/sys/kernel/hotplug"); err == nil {
					ioutil.WriteFile("/proc/sys/kernel/hotplug", []byte{}, 0644)
				}
				return nil
			},
			Command: []string{"/lib/systemd/systemd-udevd"},
		},
		&CommandService{
			Name:    "udev trigger",
			Needs:   []string{"service:udev"},
			Command: []string{"udevadm", "trigger"},
		},
		&CommandService{
			Name:    "udev settle",
			Needs:   []string{"service:udev trigger"},
			Command: []string{"udevadm", "settle"},
		},
	)
}
