package log

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/ulikunitz/xz"
)

func compress(path string) {
	in, err := os.Open(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "compress %s: failed to open: %v", path, err)
		return
	}

	defer in.Close()

	outPath := filepath.Join(filepath.Dir(path), filepath.Base(path)+".xz")

	out, err := os.Create(outPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "compress %s: failed to create target: %v", path, err)
		return
	}

	defer out.Close()

	w, err := xz.NewWriter(out)
	if err != nil {
		fmt.Fprintf(os.Stderr, "compress %s: failed to create writer: %v", path, err)
		return
	}

	if _, err := io.Copy(w, in); err != nil {
		fmt.Fprintf(os.Stderr, "compress %s: write failed: %v", path, err)
		return
	}

	if err := w.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "compress %s: close failed: %v", path, err)
		return
	}

	if err := os.Remove(path); err != nil {
		fmt.Fprintf(os.Stderr, "compress %s: failed to remove source: %v", path, err)
	}
}
