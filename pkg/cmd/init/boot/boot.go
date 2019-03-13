package initboot

import (
	"log"

	"github.com/spf13/cobra"

	"novit.nc/direktil/inits/pkg/sys"
)

var (
	doNetwork bool
)

func Command() (c *cobra.Command) {
	c = &cobra.Command{
		Use:   "boot",
		Short: "boot stage",
		Run:   run,
	}

	return
}

func run(c *cobra.Command, args []string) {
	step("files", setupFiles)
	step("modules", setupModules)
	step("network", setupNetworking)
	step("lvm", setupLVM)
}

func setupModules() {
	for _, mod := range sys.Config().Modules {
		log.Print("loading module ", mod)
		sys.Run("modprobe", mod)
	}
}
