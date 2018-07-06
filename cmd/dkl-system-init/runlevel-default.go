package main

func init() {
	services.Register(
		&CommandService{
			Name:    "sysctl",
			Needs:   []string{"configured"},
			Command: []string{"/usr/sbin/sysctl", "--system"},
		},
		&CommandService{
			Name:    "ssh-keygen",
			Needs:   []string{"files-written"},
			Command: []string{"/usr/bin/ssh-keygen", "-A"},
		},
		&CommandService{
			Name:    "sshd",
			Restart: StdRestart,
			Needs:   []string{"service:ssh-keygen"},
			Command: []string{"/usr/sbin/sshd", "-D"},
		},
		&CommandService{
			Name:    "chrony",
			Restart: StdRestart,
			Needs:   []string{"configured"},
			Command: []string{"chronyd", "-d"},
		},
	)
}
