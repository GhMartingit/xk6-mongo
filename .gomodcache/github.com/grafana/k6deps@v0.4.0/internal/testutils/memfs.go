package testutils

import (
	"testing"
	"testing/fstest"

	"github.com/grafana/k6deps/internal/rootfs"
)

// Filemap defines a map of files, given their paths an content
type Filemap map[string][]byte

// NewMapFS return an fs.Fs from a Filemap. It creates the root directory and
// adds all files are under this directory. If the root directory is not absolute
// it is made absolute with respect of an OS specific root dir.
func NewMapFS(t *testing.T, root string, files Filemap) rootfs.FS {
	t.Helper()

	memFS := fstest.MapFS{}
	for filePath, content := range files {
		memFS[filePath] = &fstest.MapFile{
			Data: content,
		}
	}

	return rootfs.NewFromFS(root, memFS)
}
