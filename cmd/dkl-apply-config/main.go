package main

import (
	"flag"
	"os"
	"strings"

	"novit.nc/direktil/inits/pkg/apply"
	"novit.nc/direktil/pkg/config"
	dlog "novit.nc/direktil/pkg/log"
)

var (
	log = dlog.Get("dkl-apply-config")
)

func main() {
	configPath := flag.String("config", "config.yaml", "config to load (\"-\" for stdin)")
	doFiles := flag.Bool("files", false, "apply files")
	filesFilters := flag.String("files-filters", "", "comma-separated filters to select files to apply")
	flag.Parse()

	log.SetConsole(os.Stderr)

	var (
		cfg *config.Config
		err error
	)

	if *configPath == "-" {
		log.Print("loading config from stdin")
		cfg, err = config.Read(os.Stdin)

	} else {
		log.Print("loading config from ", *configPath)
		cfg, err = config.Load(*configPath)
	}

	if err != nil {
		log.Print("failed to load config: ", err)
	}

	if *doFiles {
		filters := []string{}
		if *filesFilters != "" {
			filters = strings.Split(*filesFilters, ",")
		}
		if err = apply.Files(cfg, log, filters...); err != nil {
			log.Taint(dlog.Fatal, "failed to apply files: ", err)
			os.Exit(1)
		}
	}
}
