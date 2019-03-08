package initdefault

import "github.com/spf13/cobra"

func Command() (c *cobra.Command) {
	c = &cobra.Command{
		Use:   "default",
		Short: "default stage",
		Run:   run,
	}

	return
}

func run(c *cobra.Command, args []string) {
}
