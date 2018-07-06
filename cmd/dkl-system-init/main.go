package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"novit.nc/direktil/pkg/color"
	"novit.nc/direktil/pkg/config"
	"novit.nc/direktil/pkg/log"
)

const cfgPath = "/config.yaml"

var (
	bootVarPrefix = []byte("direktil.var.")
	cfg           *config.Config

	initLog = log.Get("init")
)

func main() {
	switch filepath.Base(os.Args[0]) {
	case "poweroff":
		initCommand("poweroff\n")
	case "reboot":
		initCommand("reboot\n")
	}

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "0":
			initCommand("poweroff\n")
		case "6":
			initCommand("reboot\n")
		default:
			fmt.Fprintf(os.Stderr, "unknown args: %v\n", os.Args)
			os.Exit(1)
		}
	}

	if os.Getpid() != 1 {
		fmt.Println("not PID 1")
		os.Exit(1)
	}

	color.Write(os.Stderr, color.Cyan, "Direktil system starting\n")
	initLog.SetConsole(os.Stderr)

	go handleChildren()
	go handleSignals()

	// handle abnormal ends
	defer func() {
		if err := recover(); err != nil {
			fatal("FATAL: panic in main: ", err)
		} else {
			fatal("FATAL: exited from main")
		}
	}()

	// set a reasonable path
	os.Setenv("PATH", strings.Join([]string{
		"/usr/local/bin:/usr/local/sbin",
		"/usr/bin:/usr/sbin",
		"/bin:/sbin",
	}, ":"))

	// load the configuration
	{
		c, err := config.Load(cfgPath)
		if err != nil {
			fatal("failed to load config: ", err)
		}

		if err := os.Remove(cfgPath); err != nil {
			initLog.Taint(log.Warning, "failed to remove config: ", err)
		}

		cfg = c
	}

	// bootstrap the basic things
	bootstrap()

	go listenInitctl()

	// start the services
	services.Start()

	// Wait for configuration, but timeout to always give a login
	ch := make(chan int, 1)
	go func() {
		services.Wait(func() bool {
			return services.HasFlag("configured") ||
				services.HasFlag("boot-failed")
		})
		close(ch)
	}()

	select {
	case <-time.After(1 * time.Minute):
		initLog.Taint(log.Warning, "configuration took too long, allowing login anyway.")
	case <-ch:
	}

	// Handle CAD command (ctrl+alt+del)
	intCh := make(chan os.Signal, 1)
	signal.Notify(intCh, syscall.SIGINT)

	syscall.Reboot(syscall.LINUX_REBOOT_CMD_CAD_ON)
	go func() {
		<-intCh
		initLog.Taint(log.Warning, "received ctrl+alt+del, rebooting...")
		reboot()
	}()

	// Allow login now
	go allowLogin()

	// Wait all services
	services.WaitAll()
	initLog.Taint(log.OK, "all services are started")

	endOfProcess()
}

func initCommand(c string) {
	err := ioutil.WriteFile("/run/initctl", []byte(c), 0600)
	if err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

func mkdir(dir string, mode os.FileMode) {
	if err := os.MkdirAll(dir, mode); err != nil {
		fatalf("mkdir %q failed: %v", dir, err)
	}
}

func run(cmd string, args ...string) {
	c := exec.Command(cmd, args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	if err := c.Run(); err != nil {
		fatalf("command %s %q failed: %v", cmd, args, err)
	}
}

func touch(path string) {
	run("touch", path)
}

func poweroff() {
	prepareShutdown()

	initLog.Print("final sync")
	syscall.Sync()

	syscall.Reboot(syscall.LINUX_REBOOT_CMD_POWER_OFF)
}

func reboot() {
	prepareShutdown()

	initLog.Print("final sync")
	syscall.Sync()

	syscall.Reboot(syscall.LINUX_REBOOT_CMD_RESTART)
}

func prepareShutdown() {
	services.Stop()
	initLog.Taint(log.Info, "services stopped")

	log.DisableFiles()

	for try := 0; try < 5; try++ {
		initLog.Taint(log.Info, "unmounting filesystems")

		// FIXME: filesystem list should be build from non "nodev" lines in /proc/filesystems
		c := exec.Command("umount", "-a", "-t", "ext2,ext3,ext4,vfat,msdos,xfs,btrfs")
		c.Stdout = initLog
		c.Stderr = initLog

		if err := c.Run(); err != nil {
			initLog.Taint(log.Warning, "umounting failed: ", err)
			time.Sleep(time.Duration(2*try) * time.Second)
			continue
		}

		break
	}

	initLog.Taint(log.Info, "sync'ing")
	exec.Command("sync").Run()
}

func allowLogin() {
	b := make([]byte, 1)
	for {
		os.Stdout.Write([]byte("\n" + color.Yellow + "[press enter to login]" + color.Reset + "\n\n"))
		for {
			os.Stdin.Read(b)
			if b[0] == '\n' {
				break
			}
		}

		c := exec.Command("/sbin/agetty", "--noclear", "--noissue", "console", "linux")
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr

		c.Run()
	}
}
