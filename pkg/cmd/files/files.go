package files

import (
	"log"

	"github.com/spf13/cobra"
	pconfig "novit.nc/direktil/pkg/config"
)

var (
	configPath string
	config     *pconfig.Config
)

func Command() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:  "files",
		Args: cobra.NoArgs,

		PersistentPreRun: loadConfig,
	}

	cmd.PersistentFlags().StringVar(&configPath, "config", "/boot/config.yaml", "path to the boot config")

	cmd.AddCommand(
		listCommand(),
	)

	return
}

func loadConfig(_ *cobra.Command, _ []string) {
	c, err := pconfig.Load(configPath)

	if err != nil {
		log.Fatal("failed to load config: ", err)
	}

	config = c
}
