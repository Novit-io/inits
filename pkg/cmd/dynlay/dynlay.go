package dynlay

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/antage/mntent"
	"github.com/spf13/cobra"
)

var urlPrefix = "https://dkl.novit.nc/dist/layers"

func Command() (c *cobra.Command) {
	c = &cobra.Command{
		Use:   "dynlay <layer> <version>",
		Short: "set dynamic layer",
		Args:  cobra.ExactArgs(2),

		Run: run,
	}

	c.Flags().StringVar(&urlPrefix, "url-prefix", urlPrefix, "Layer URL prefix")

	return c
}

func run(_ *cobra.Command, args []string) {
	layer, version := args[0], args[1]

	layDir := filepath.Join("/opt/dynlay", layer)

	err := os.MkdirAll(layDir, 0755)
	fail(err)

	layPath := filepath.Join(layDir, version)

	// fetch if not exist
	_, err = os.Stat(layPath)
	if os.IsNotExist(err) {
		fetch(layPath, urlPrefix+"/"+layer+"/"+version)
	}

	mounted := map[string]bool{}
	{
		mounts, err := mntent.Parse("/etc/mtab")
		fail(err)
		for _, mount := range mounts {
			mounted[mount.Directory] = true
		}
	}

	// mount
	mountPath := filepath.Join("/run/dynlay", layer, version)

	if !mounted[mountPath] {
		err = os.MkdirAll(mountPath, 0755)
		fail(err)

		cmd := exec.Command("mount", "-t", "squashfs", layPath, mountPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			os.Remove(mountPath)
			fail(err)
		}
	}

	// update links
	existing := map[string]bool{}
	{
		paths, err := linkableFrom(mountPath)
		fail(err)

		for _, path := range paths {
			existing[path] = true

			target := "/" + path

			fmt.Println("linking", path)

			os.Remove(target)

			os.MkdirAll(filepath.Dir(target), 0755)

			err := os.Symlink(filepath.Join(mountPath, path), target)
			if err != nil {
				log.Print("warning: symlink of ", path, " failed: ", err)
			}
		}
	}

	// remove expired links
	basePaths, err := filepath.Glob(filepath.Join(filepath.Dir(mountPath), "*"))
	fail(err)

	for _, base := range basePaths {
		if base == mountPath {
			continue
		}

		stat, err := os.Stat(base)
		fail(err)

		if !stat.IsDir() {
			continue
		}

		paths, err := linkableFrom(base)
		fail(err)

		for _, path := range paths {
			if existing[path] {
				continue
			}

			_, err := os.Stat("/" + path)
			if os.IsNotExist(err) {
				continue
			}

			fmt.Print("unlinking ", path, " (from ", filepath.Base(base), ")\n")

			err = os.Remove("/" + path)
			if err != nil {
				log.Print("warning: failed to unlink ", path, ": ", err)
			}
		}

		if mounted[base] {
			cmd := exec.Command("umount", "-l", base)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Run()
		}

		os.Remove(base)
	}
}

func linkableFrom(dir string) (paths []string, err error) {
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		path, err = filepath.Rel(dir, path)
		paths = append(paths, path)
		return nil
	})
	return
}

func fetch(outPath string, url string) {
	f, err := os.Create(outPath + ".part")
	fail(err)

	defer f.Close()

	log.Print("fetching ", url)
	resp, err := http.Get(url)
	fail(err)

	if resp.StatusCode != 200 {
		log.Fatal("GET failed on ", url, ": ", resp.Status)
	}

	_, err = io.Copy(f, resp.Body)
	fail(err)

	f.Close()

	os.Rename(outPath+".part", outPath)
}

func fail(err error) {
	if err != nil {
		log.SetFlags(log.Flags() | log.Lshortfile)
		log.Output(2, err.Error())
		os.Exit(-1)
	}
}
