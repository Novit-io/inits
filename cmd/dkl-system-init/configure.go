package main

import (
	"bytes"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/sparrc/go-ping"

	"novit.nc/direktil/inits/pkg/apply"
	"novit.nc/direktil/inits/pkg/vars"
	"novit.nc/direktil/pkg/config"
	"novit.nc/direktil/pkg/log"
)

func init() {
	services.Register(configure{})
}

type configure struct{}

func (_ configure) GetName() string {
	return "configure"
}

func (_ configure) CanStart() bool {
	return services.HasFlag("service:lvm", "service:udev trigger")
}

func (_ configure) Run(_ func()) error {
	// make root rshared (default in systemd, required by Kubernetes 1.10+)
	// equivalent to "mount --make-rshared /"
	// see kernel's Documentation/sharedsubtree.txt (search rshared)
	if err := syscall.Mount("", "/", "", syscall.MS_SHARED|syscall.MS_REC, ""); err != nil {
		fatalf("mount --make-rshared / failed: %v", err)
	}

	// - setup root user
	if passwordHash := cfg.RootUser.PasswordHash; passwordHash == "" {
		run("/usr/bin/passwd", "-d", "root")
	} else {
		run("/bin/sh", "-c", "chpasswd --encrypted <<EOF\nroot:"+passwordHash+"\nEOF")
	}

	// - groups
	for _, group := range cfg.Groups {
		opts := make([]string, 0)
		opts = append(opts, "-r")
		if group.Gid != 0 {
			opts = append(opts, "-g", strconv.Itoa(group.Gid))
		}
		opts = append(opts, group.Name)

		run("groupadd", opts...)
	}

	// - user
	for _, user := range cfg.Users {
		opts := make([]string, 0)
		opts = append(opts, "-r")
		if user.Gid != 0 {
			opts = append(opts, "-g", strconv.Itoa(user.Gid))
		}
		if user.Uid != 0 {
			opts = append(opts, "-u", strconv.Itoa(user.Uid))
		}
		opts = append(opts, user.Name)

		run("useradd", opts...)
	}

	// - files
	if err := apply.Files(cfg, initLog); err != nil {
		fatal(err)
	}
	services.SetFlag("files-written")

	// - hostname
	initLog.Taint(log.Info, "setting hostname")
	run("hostname", "-F", "/etc/hostname")

	// - modules
	for _, module := range cfg.Modules {
		initLog.Taint(log.Info, "loading module ", module)
		run("modprobe", module)
	}

	// - networks
	for idx, network := range cfg.Networks {
		setupNetwork(idx, network)
	}

	services.SetFlag("network-up")

	// - setup storage
	initLog.Print("checking storage")
	if err := exec.Command("vgdisplay", "storage").Run(); err != nil {
		initLog.Print("creating VG storage")
		setupVG(vars.BootArgValue("storage", cfg.Storage.UdevMatch))
	}

	for _, name := range cfg.Storage.RemoveVolumes {
		dev := "/dev/storage/" + name

		if _, err := os.Stat(dev); os.IsNotExist(err) {
			continue

		} else if err != nil {
			fatal("failed to stat ", dev, ": ", err)
		}

		initLog.Print("removing LV ", name)
		cmd := exec.Command("lvremove", "-f", "storage/"+name)
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fatal("failed to remove LV ", name)
		}
	}

	for _, volume := range cfg.Storage.Volumes {
		if err := exec.Command("lvdisplay", "storage/"+volume.Name).Run(); err != nil {
			initLog.Print("creating LV ", volume.Name)
			setupLV(volume)
		}

		dev := "/dev/storage/" + volume.Name

		initLog.Printf("checking filesystem on %s", dev)
		run("fsck", "-p", dev)

		mount(dev, volume.Mount.Path, volume.FS,
			syscall.MS_NOATIME|syscall.MS_RELATIME,
			volume.Mount.Options)
	}

	// finished configuring :-)
	log.EnableFiles()
	services.SetFlag("configured")

	// load user services
	go loadUserServices()

	return nil
}

func (_ configure) Stop() {
	// no-op
}

var networkStarted = map[string]bool{}

func setupNetwork(idx int, network config.NetworkDef) {
	tries := 0
retry:
	ifaces, err := net.Interfaces()
	if err != nil {
		fatalf("failed to get network interfaces: %v", err)
	}

	match := false
	for _, iface := range ifaces {
		if networkStarted[iface.Name] {
			continue
		}

		if network.Match.Name != "" {
			if ok, err := filepath.Match(network.Match.Name, iface.Name); err != nil {
				fatalf("network[%d] name match error: %v", idx, err)
			} else if !ok {
				continue
			}
		}

		if network.Match.Ping != nil {
			initLog.Printf("network[%d] ping check on %s", idx, iface.Name)

			if ok, err := networkPingCheck(iface.Name, network); err != nil {
				initLog.Taintf(log.Error, "network[%d] ping check failed: %v",
					idx, err)
			} else if !ok {
				continue
			}
		}

		initLog.Printf("network[%d] matches interface %s", idx, iface.Name)
		match = true

		startNetwork(iface.Name, idx, network)

		if !network.Match.All {
			return
		}
	}

	if !match {
		initLog.Taintf(log.Warning, "network[%d] did not match any interface", idx)

		tries++
		if network.Optional && tries > 3 {
			return
		}

		time.Sleep(1 * time.Second)
		initLog.Taintf(log.Warning, "network[%d] retrying (try: %d)", idx, tries)
		goto retry
	}
}

func startNetwork(ifaceName string, idx int, network config.NetworkDef) {
	initLog.Taintf(log.Info, "starting network[%d]", idx)

	script := vars.Substitute([]byte(network.Script), cfg)

	c := exec.Command("/bin/sh")
	c.Stdin = bytes.NewBuffer(script)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	// TODO doc
	c.Env = append(append(make([]string, 0), os.Environ()...), "IFNAME="+ifaceName)

	if err := c.Run(); err != nil {
		links, _ := exec.Command("ip", "link", "ls").CombinedOutput()
		fatalf("network setup failed (link list below): %v\n%s", err, string(links))
	}

	networkStarted[ifaceName] = true
}

func networkPingCheck(ifName string, network config.NetworkDef) (bool, error) {
	check := network.Match.Ping

	source := string(vars.Substitute([]byte(check.Source), cfg))

	run("ip", "addr", "add", source, "dev", ifName)
	run("ip", "link", "set", ifName, "up")

	defer func() {
		run("ip", "link", "set", ifName, "down")
		run("ip", "addr", "del", source, "dev", ifName)
	}()

	pinger, err := ping.NewPinger(network.Match.Ping.Target)
	if err != nil {
		return false, err
	}

	pinger.Count = 3
	if check.Count > 0 {
		pinger.Count = check.Count
	}

	pinger.Timeout = 1 * time.Second
	if check.Timeout > 0 {
		pinger.Timeout = time.Duration(check.Timeout) * time.Second
	}

	pinger.SetPrivileged(true)
	pinger.Run()

	return pinger.Statistics().PacketsRecv > 0, nil
}
