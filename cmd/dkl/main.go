package main

import (
	"log"

	"github.com/spf13/cobra"

	"novit.nc/direktil/inits/pkg/cmd/applyconfig"
	cmddynlay "novit.nc/direktil/inits/pkg/cmd/dynlay"
	cmdinit "novit.nc/direktil/inits/pkg/cmd/init"
)

func main() {
	root := &cobra.Command{}

	root.AddCommand(cmdinit.Command())
	root.AddCommand(cmddynlay.Command())
	root.AddCommand(applyconfig.Command())

	if err := root.Execute(); err != nil {
		log.Fatal("error: ", err)
	}
}
