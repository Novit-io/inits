package main

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	yaml "gopkg.in/yaml.v2"

	"novit.nc/direktil/pkg/log"
)

var (
	reYamlStart = regexp.MustCompile("^#\\s+---\\s*$")
)

type UserService struct {
	Restart  int
	Needs    []string
	Provides []string
}

func loadUserServices() {
retry:
	files, err := filepath.Glob("/etc/direktil/services/*")
	if err != nil {
		initLog.Taint(log.Error, "failed to load user services: ", err)
		time.Sleep(10 * time.Second)
		goto retry
	}

	for _, path := range files {
		path := path
		go func() {
			for {
				if err := loadUserService(path); err != nil {
					initLog.Taintf(log.Error, "failed to load %s: %v", path, err)
					time.Sleep(10 * time.Second)
					continue
				}
				break
			}
		}()
	}
}

func loadUserService(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	defer f.Close()

	r := bufio.NewReader(f)

	yamlBuf := &bytes.Buffer{}
	inYaml := false

	for {
		line, err := r.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		if inYaml {
			if !strings.HasPrefix(line, "# ") {
				break
			}

			yamlBuf.WriteString(line[2:])

		} else if reYamlStart.MatchString(line) {
			inYaml = true
		}
	}

	spec := &UserService{}

	if inYaml {
		if err := yaml.Unmarshal(yamlBuf.Bytes(), &spec); err != nil {
			return err
		}
	}

	svc := &CommandService{
		Name:     filepath.Base(path),
		Restart:  time.Duration(spec.Restart) * time.Second,
		Needs:    spec.Needs,
		Provides: spec.Provides,
		Command:  []string{path},
	}

	initLog.Taintf(log.OK, "user service: %s", path)
	services.Register(svc)

	return nil
}
