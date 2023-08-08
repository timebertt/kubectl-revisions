package diff

import (
	"os"
	"path/filepath"

	"github.com/timebertt/kubectl-history/pkg/runutil"
)

// Files is a compound handle for multiple files in a directory that shall be compared using a diff programm.
type Files struct {
	Dir  string
	A, B *os.File
}

// NewFiles creates a temporary directory and two files in it, and returns a Files handle.
func NewFiles(dirNamePrefix, fileNameA, fileNameB string) (f *Files, err error) {
	f = &Files{}

	f.Dir, err = os.MkdirTemp("", dirNamePrefix+"-")
	if err != nil {
		return nil, err
	}

	defer func() {
		// if we weren't able to create both files, clean up the directory before returning the error
		if err != nil {
			runutil.CaptureError(&err, f.TearDown)
		}
	}()

	flags := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	if f.A, err = os.OpenFile(filepath.Join(f.Dir, fileNameA), flags, 0600); err != nil {
		return nil, err
	}

	if f.B, err = os.OpenFile(filepath.Join(f.Dir, fileNameB), flags, 0600); err != nil {
		return nil, err
	}

	return f, nil
}

// TearDown removes the temporary directory held by this handle.
func (f Files) TearDown() error {
	return os.RemoveAll(f.Dir)
}
