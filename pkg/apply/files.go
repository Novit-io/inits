package apply

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"novit.nc/direktil/inits/pkg/vars"
	"novit.nc/direktil/pkg/config"
)

const (
	authorizedKeysPath = "/root/.ssh/authorized_keys"
)

// Files writes the files from the given config
func Files(cfg *config.Config, filters ...string) (err error) {
	accept := func(n string) bool { return true }

	if len(filters) > 0 {
		accept = func(n string) bool {
			for _, filter := range filters {
				if matched, err := filepath.Match(filter, n); err != nil {
					log.Printf("ERROR: bad filter ignored: %q: %v", filter, err)
				} else if matched {
					return true
				}
			}
			return false
		}
	}

	if cfg.RootUser.AuthorizedKeys != nil && accept(authorizedKeysPath) {
		err = writeFile(
			authorizedKeysPath,
			[]byte(strings.Join(cfg.RootUser.AuthorizedKeys, "\n")),
			0600, 0700, cfg,
		)

		if err != nil {
			return
		}
	}

	for _, file := range cfg.Files {
		if !accept(file.Path) {
			continue
		}

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
		)

		if err != nil {
			log.Print("failed to write file ", file.Path, ": ", err)
			continue
		}
	}

	return
}

func writeFile(path string, content []byte, fileMode, dirMode os.FileMode, cfg *config.Config) (err error) {
	if err = os.MkdirAll(filepath.Dir(path), dirMode); err != nil {
		return
	}

	content = vars.Substitute(content, cfg)

	log.Printf("writing %q, mode %04o, %d bytes", path, fileMode, len(content))
	if err = ioutil.WriteFile(path, content, fileMode); err != nil {
		err = fmt.Errorf("failed to write %s: %v", path, err)
	}

	if chmodErr := os.Chmod(path, fileMode); chmodErr != nil {
		log.Print("- failed chmod: ", chmodErr)
	}

	return
}
