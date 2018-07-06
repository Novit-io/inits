package main

import (
	"os/exec"
	"syscall"
	"time"

	"novit.nc/direktil/pkg/log"
)

var (
	// StdRestart is a wait duration between restarts of a service, if you have no inspiration.
	StdRestart = 1 * time.Second

	killDelay = 30 * time.Second
)

// Service represents a service to run as part of the init process.
type Service interface {
	GetName() string
	CanStart() bool
	Run(notify func()) error
	Stop()
}

type CommandService struct {
	Name     string
	Command  []string
	Restart  time.Duration
	Needs    []string
	Provides []string
	PreExec  func() error

	log     *log.Log
	stop    bool
	command *exec.Cmd
}

var _ Service = &CommandService{}

func (s *CommandService) GetName() string {
	return s.Name
}

// CanStart is part of the Service interface
func (s *CommandService) CanStart() bool {
	return services.HasFlag(s.Needs...)
}

func (s *CommandService) Stop() {
	stopped := false

	s.stop = true

	c := s.command
	if c == nil {
		return
	}

	c.Process.Signal(syscall.SIGTERM)

	go func() {
		time.Sleep(killDelay)
		if !stopped {
			c.Process.Signal(syscall.SIGKILL)
		}
	}()

	c.Wait()
	stopped = true
}

func (s *CommandService) Run(notify func()) error {
	s.stop = false

	if s.log == nil {
		s.log = log.Get(s.Name)
	}

	isOneshot := s.Restart == time.Duration(0)

	myNotify := func() {
		for _, provide := range s.Provides {
			services.SetFlag(provide)
		}

		notify()
	}

	if s.PreExec != nil {
		if err := s.PreExec(); err != nil {
			return err
		}
	}

	// Starting
	var err error

retry:
	if isOneshot {
		// oneshot services are only active after exit
		err = s.exec(func() {})
	} else {
		err = s.exec(myNotify)
	}

	if s.stop {
		return err
	}

	if isOneshot {
		myNotify()

	} else {
		// auto-restart service
		services.Set(s.Name, Failed)
		time.Sleep(s.Restart)

		s.log.Taintf(log.Warning, "-- restarting --")
		services.Set(s.Name, Starting)
		goto retry
	}

	return err
}

func (s *CommandService) exec(notify func()) error {
	c := exec.Command(s.Command[0], s.Command[1:]...)

	s.command = c
	defer func() {
		s.command = nil
	}()

	c.Stdout = s.log
	c.Stderr = s.log

	if err := c.Start(); err != nil {
		s.log.Taintf(log.Error, "failed to start: %v", err)
		return err
	}

	notify()

	if err := c.Wait(); err != nil {
		s.log.Taintf(log.Error, "failed: %v", err)
		return err
	}

	return nil
}
