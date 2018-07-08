package apply

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"novit.nc/direktil/inits/pkg/vars"
	"novit.nc/direktil/pkg/config"
	dlog "novit.nc/direktil/pkg/log"
)

// Files writes the files from the given config
func Files(cfg *config.Config, log *dlog.Log) (err error) {
	if cfg.RootUser.AuthorizedKeys != nil {
		err = writeFile(
			"/root/.ssh/authorized_keys",
			[]byte(strings.Join(cfg.RootUser.AuthorizedKeys, "\n")),
			0600, 0700, cfg, log,
		)

		if err != nil {
			return
		}
	}

	for _, file := range cfg.Files {
		mode := file.Mode
		if mode == 0 {
			mode = 0644
		}

		content := []byte(file.Content)

		err = writeFile(
			file.Path,
			content,
			mode,
			0755,
			cfg,
			log,
		)

		if err != nil {
			return
		}
	}

	return
}

func writeFile(path string, content []byte, fileMode, dirMode os.FileMode,
	cfg *config.Config, log *dlog.Log) (err error) {

	if err = os.MkdirAll(filepath.Dir(path), dirMode); err != nil {
		return
	}

	content = vars.Substitute(content, cfg)

	log.Printf("writing %q, mode %04o, %d bytes", path, fileMode, len(content))
	if err = ioutil.WriteFile(path, content, fileMode); err != nil {
		err = fmt.Errorf("failed to write %s: %v", path, err)
	}

	return
}
