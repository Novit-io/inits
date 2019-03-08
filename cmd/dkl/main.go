package main

import (
	"log"

	"github.com/spf13/cobra"

	cmdinit "novit.nc/direktil/inits/pkg/cmd/init"
)

func main() {
	root := &cobra.Command{}

	root.AddCommand(cmdinit.Command())

	if err := root.Execute(); err != nil {
		log.Fatal("error: ", err)
	}
}
