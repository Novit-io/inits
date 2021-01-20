package applyconfig

import (
	"flag"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"novit.nc/direktil/inits/pkg/apply"
	"novit.nc/direktil/pkg/config"

	dlog "novit.nc/direktil/pkg/log"
)

var (
	filesFilters string
	log          = dlog.Get("dkl")
)

func Command() (c *cobra.Command) {
	c = &cobra.Command{
		Use:   "apply-config <config.yaml>",
		Short: "apply a config to the current system",
		Args:  cobra.ExactArgs(1),

		Run: run,
	}

	flag.StringVar(&filesFilters, "files-filters", "", "comma-separated filters to select files to apply")

	return c
}

func run(_ *cobra.Command, args []string) {
	configPath := args[0]

	var (
		cfg *config.Config
		err error
	)

	if configPath == "-" {
		log.Print("loading config from stdin")
		cfg, err = config.Read(os.Stdin)

	} else {
		log.Print("loading config from ", configPath)
		cfg, err = config.Load(configPath)
	}

	if err != nil {
		log.Print("failed to load config: ", err)
	}

	filters := []string{}
	if filesFilters != "" {
		filters = strings.Split(filesFilters, ",")
	}
	if err = apply.Files(cfg /*log,*/, filters...); err != nil {
		log.Taint(dlog.Fatal, "failed to apply files: ", err)
		os.Exit(1)
	}
}
