package files

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	pconfig "novit.nc/direktil/pkg/config"
)

func listCommand() (cmd *cobra.Command) {
	return &cobra.Command{
		Use: "list",
		Run: list,
	}
}

func list(_ *cobra.Command, args []string) {
	for _, file := range filteredFiles(args) {
		fmt.Println(file.Path)
	}
}

func filteredFiles(filters []string) (ret []pconfig.FileDef) {
	ret = make([]pconfig.FileDef, 0, len(config.Files))

	for _, file := range config.Files {
		if len(filters) != 0 {
			match := false
			for _, filter := range filters {
				if ok, _ := filepath.Match(filter, file.Path); ok {
					match = true
					break
				}
			}

			if !match {
				continue
			}
		}

		ret = append(ret, file)
	}

	return
}
