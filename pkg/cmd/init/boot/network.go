package initboot

import (
	"bytes"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	ping "github.com/sparrc/go-ping"
	"novit.nc/direktil/pkg/config"

	"novit.nc/direktil/inits/pkg/sys"
	"novit.nc/direktil/inits/pkg/vars"
)

var networkStarted = map[string]bool{}

func setupNetworking() {
	cfg := sys.Config()
	for idx, network := range cfg.Networks {
		setupNetwork(idx, network)
	}
}

func setupNetwork(idx int, network config.NetworkDef) {
	tries := 0
retry:
	ifaces, err := net.Interfaces()
	if err != nil {
		log.Fatalf("FATAL: failed to get network interfaces: %v", err)
	}

	match := false
	for _, iface := range ifaces {
		if networkStarted[iface.Name] {
			continue
		}

		if network.Match.Name != "" {
			if ok, err := filepath.Match(network.Match.Name, iface.Name); err != nil {
				log.Fatalf("FATAL: network[%d] name match error: %v", idx, err)
			} else if !ok {
				continue
			}
		}

		if network.Match.Ping != nil {
			log.Printf("network[%d] ping check on %s", idx, iface.Name)

			if ok, err := networkPingCheck(iface.Name, network); err != nil {
				log.Printf("ERROR: network[%d] ping check failed: %v", idx, err)

			} else if !ok {
				continue
			}
		}

		log.Printf("network[%d] matches interface %s", idx, iface.Name)
		match = true

		startNetwork(iface.Name, idx, network)

		if !network.Match.All {
			return
		}
	}

	if !match {
		log.Printf("WARNING: network[%d] did not match any interface", idx)

		tries++
		if network.Optional && tries > 3 {
			return
		}

		time.Sleep(1 * time.Second)
		log.Printf("WARNING: network[%d] retrying (try: %d)", idx, tries)
		goto retry
	}
}

func startNetwork(ifaceName string, idx int, network config.NetworkDef) {
	cfg := sys.Config()

	log.Printf("starting network[%d]", idx)

	script := vars.Substitute([]byte(network.Script), cfg)

	c := exec.Command("/bin/sh")
	c.Stdin = bytes.NewBuffer(script)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	// TODO doc
	c.Env = append(append(make([]string, 0), os.Environ()...), "IFNAME="+ifaceName)

	if err := c.Run(); err != nil {
		links, _ := exec.Command("ip", "link", "ls").CombinedOutput()
		log.Fatalf("FATAL: network setup failed (link list below): %v\n%s", err, string(links))
	}

	networkStarted[ifaceName] = true
}

func networkPingCheck(ifName string, network config.NetworkDef) (b bool, err error) {
	check := network.Match.Ping

	source := string(vars.Substitute([]byte(check.Source), sys.Config()))

	if err = sys.Run("ip", "addr", "add", source, "dev", ifName); err != nil {
		return
	}
	if err = sys.Run("ip", "link", "set", ifName, "up"); err != nil {
		return
	}

	defer func() {
		sys.MustRun("ip", "link", "set", ifName, "down")
		sys.MustRun("ip", "addr", "del", source, "dev", ifName)
	}()

	count := 3
	if check.Count != 0 {
		count = check.Count
	}

	for n := 0; n < count; n++ {
		// TODO probably better to use golang.org/x/net/icmp directly
		pinger, e := ping.NewPinger(network.Match.Ping.Target)
		if e != nil {
			err = e
			return
		}

		pinger.Count = 1

		pinger.Timeout = 1 * time.Second
		if check.Timeout > 0 {
			pinger.Timeout = time.Duration(check.Timeout) * time.Second
		}

		pinger.SetPrivileged(true)
		pinger.Run()

		if pinger.Statistics().PacketsRecv > 0 {
			b = true
			return
		}
	}

	return
}
