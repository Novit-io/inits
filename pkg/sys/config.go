package sys

import (
	"log"
	"sync"

	"novit.nc/direktil/pkg/config"
)

const cfgPath = "/config.yaml"

var (
	cfg     *config.Config
	cfgLock sync.Mutex
)

func Config() *config.Config {
	if cfg != nil {
		return cfg
	}

	cfgLock.Lock()
	defer cfgLock.Unlock()

	if cfg != nil {
		return cfg
	}

	c, err := config.Load(cfgPath)
	if err != nil {
		log.Fatal("FATAL: failed to load config: ", err)
	}

	cfg = c
	return cfg
}
