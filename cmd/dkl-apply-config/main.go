package main

import (
	"flag"
	"os"

	"novit.nc/direktil/inits/pkg/apply"
	"novit.nc/direktil/pkg/config"
	dlog "novit.nc/direktil/pkg/log"
)

var (
	log = dlog.Get("dkl-apply-config")
)

func main() {
	configPath := flag.String("config", "config.yaml", "config to load")
	doFiles := flag.Bool("files", false, "apply files")
	flag.Parse()

	log.SetConsole(os.Stderr)

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Print("failed to load config: ", err)
	}

	if *doFiles {
		apply.Files(cfg, log)
	}
}
