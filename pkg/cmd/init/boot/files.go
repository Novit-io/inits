package initboot

import (
	"log"
	"strconv"
	"syscall"

	"novit.nc/direktil/inits/pkg/apply"
	"novit.nc/direktil/inits/pkg/sys"
)

func setupFiles() {
	cfg := sys.Config()

	// make root rshared (default in systemd, required by Kubernetes 1.10+)
	// equivalent to "mount --make-rshared /"
	// see kernel's Documentation/sharedsubtree.txt (search rshared)
	if err := syscall.Mount("", "/", "", syscall.MS_SHARED|syscall.MS_REC, ""); err != nil {
		log.Fatalf("FATAL: mount --make-rshared / failed: %v", err)
	}

	// - setup root user
	if passwordHash := cfg.RootUser.PasswordHash; passwordHash == "" {
		sys.MustRun("/usr/bin/passwd", "-d", "root")
	} else {
		sys.MustRun("/bin/sh", "-c", "chpasswd --encrypted <<EOF\nroot:"+passwordHash+"\nEOF")
	}

	// - groups
	for _, group := range cfg.Groups {
		opts := make([]string, 0)
		opts = append(opts, "-r")
		if group.Gid != 0 {
			opts = append(opts, "-g", strconv.Itoa(group.Gid))
		}
		opts = append(opts, group.Name)

		sys.MustRun("groupadd", opts...)
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

		sys.MustRun("useradd", opts...)
	}

	// - files
	if err := apply.Files(cfg); err != nil {
		log.Fatal("FATAL: ", err)
	}
}
