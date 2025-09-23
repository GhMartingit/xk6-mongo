package goproxy

import (
	"io/fs"
	"os"
	"path/filepath"
)

// ReadDir reads the content of the files in a directory into a map.
// The maps has the path, relative to the root dir, of each file
func ReadDir(rootDir string) (map[string][]byte, error) {
	files := map[string][]byte{}

	err := filepath.Walk(rootDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		content, err := os.ReadFile(path) //nolint:forbidigo,gosec
		if err != nil {
			return err
		}

		fileName, _ := filepath.Rel(rootDir, path)
		files[fileName] = content

		return nil
	})

	return files, err
}
