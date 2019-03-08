package cmdinit

import (
	"os"
	"strings"

	"github.com/spf13/cobra"

	initboot "novit.nc/direktil/inits/pkg/cmd/init/boot"
	initdefault "novit.nc/direktil/inits/pkg/cmd/init/default"
	initservice "novit.nc/direktil/inits/pkg/cmd/init/service"
)

func Command() (c *cobra.Command) {
	c = &cobra.Command{
		Use:   "init",
		Short: "init stages",

		PersistentPreRun: func(_ *cobra.Command, _ []string) {
			// set a reasonable path
			os.Setenv("PATH", strings.Join([]string{
				"/usr/local/bin:/usr/local/sbin",
				"/usr/bin:/usr/sbin",
				"/bin:/sbin",
			}, ":"))
		},
	}

	c.AddCommand(initboot.Command())
	c.AddCommand(initdefault.Command())
	c.AddCommand(initservice.Command())

	return
}
