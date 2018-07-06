package main

import (
	"os"
	"sync"
	"time"

	"novit.nc/direktil/pkg/log"
)

var (
	services = ServiceRegistry{
		rw:       &sync.RWMutex{},
		cond:     sync.NewCond(&sync.Mutex{}),
		services: make(map[string]Service),
		statuses: make(map[string]ServiceStatus),
		flags:    make(map[string]bool),
	}
)

type ServiceStatus int

const (
	Pending ServiceStatus = iota
	Starting
	Running
	Failed
	Exited
)

type ServiceRegistry struct {
	rw       *sync.RWMutex
	cond     *sync.Cond
	services map[string]Service
	statuses map[string]ServiceStatus
	flags    map[string]bool
	started  bool
}

func (sr *ServiceRegistry) Register(services ...Service) {
	for _, service := range services {
		name := service.GetName()

		if _, ok := sr.services[name]; ok {
			fatalf("duplicated service name: %s", name)
		}

		sr.services[name] = service

		if sr.started {
			go sr.startService(name, service)
		}
	}
}

func (sr *ServiceRegistry) Start() {
	initLog.Taintf(log.Info, "starting service registry")
	sr.started = true
	for name, service := range sr.services {
		go sr.startService(name, service)
	}
}

func (sr *ServiceRegistry) startService(name string, service Service) {
	sr.Set(name, Pending)
	sr.Wait(service.CanStart)

	sr.Set(name, Starting)
	initLog.Taintf(log.Info, "starting service %s", name)

	if err := service.Run(func() {
		sr.Set(name, Running)
		sr.SetFlag("service:" + name)
	}); err == nil {
		initLog.Taintf(log.OK, "service %s finished.", name)
		sr.Set(name, Exited)
		sr.SetFlag("service:" + name)

	} else {
		initLog.Taintf(log.Error, "service %s failed: %v", name, err)
		sr.Set(name, Failed)
	}
}

func (sr *ServiceRegistry) Stop() {
	wg := sync.WaitGroup{}
	wg.Add(len(sr.services))

	for name, service := range sr.services {
		name, service := name, service
		go func() {
			initLog.Taintf(log.Info, "stopping %s", name)
			service.Stop()
			wg.Done()
		}()
	}

	wg.Wait()
}

func (sr *ServiceRegistry) Wait(check func() bool) {
	sr.cond.L.Lock()
	defer sr.cond.L.Unlock()

	for {
		if check() {
			return
		}

		sr.cond.Wait()
	}
}

func (sr *ServiceRegistry) WaitAll() {
	flags := make([]string, 0, len(sr.services))
	for name, _ := range sr.services {
		flags = append(flags, "service:"+name)
	}
	sr.Wait(sr.HasFlagCond(flags...))
}

func (sr *ServiceRegistry) Set(serviceName string, status ServiceStatus) {
	sr.cond.L.Lock()
	sr.rw.Lock()
	defer func() {
		sr.rw.Unlock()
		sr.cond.L.Unlock()
		sr.cond.Broadcast()
	}()

	sr.statuses[serviceName] = status
}

func (sr *ServiceRegistry) HasStatus(status ServiceStatus, serviceNames ...string) bool {
	sr.rw.RLock()
	defer sr.rw.RUnlock()

	for _, name := range serviceNames {
		if sr.statuses[name] != status {
			return false
		}
	}
	return true
}

func (sr *ServiceRegistry) SetFlag(flag string) {
	sr.cond.L.Lock()
	sr.rw.Lock()
	defer func() {
		sr.rw.Unlock()
		sr.cond.L.Unlock()
		sr.cond.Broadcast()
	}()

	sr.flags[flag] = true
}

func (sr *ServiceRegistry) HasFlag(flags ...string) bool {
	sr.rw.RLock()
	defer sr.rw.RUnlock()

	for _, flag := range flags {
		if !sr.flags[flag] {
			return false
		}
	}
	return true
}

func (sr *ServiceRegistry) HasFlagCond(flags ...string) func() bool {
	return func() bool {
		return sr.HasFlag(flags...)
	}
}

func (sr *ServiceRegistry) WaitPath(path string) {
	for {
		if _, err := os.Stat(path); err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	sr.SetFlag("file:" + path)
}
