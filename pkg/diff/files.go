package diff

import (
	"os"
	"path/filepath"

	"github.com/hashicorp/go-multierror"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/printers"

	"github.com/timebertt/kubectl-revisions/pkg/runutil"
)

// Files is a compound handle for multiple directories and files that shall be compared using a diff program.
// This is similar to how `kubectl diff` works. This should behave similarly (e.g., create files in two different
// directories) as some external diff tools might have some heuristic detections in places, e.g., see dyff:
// https://github.com/homeport/dyff/blame/c382d5132c86d2280335f4cb71754ab20776a85a/internal/cmd/root.go#L85-L98
type Files struct {
	From, To *Version
}

// NewFiles creates two Version handles (i.e., two temporary directories).
func NewFiles(from, to string) (f *Files, err error) {
	f = &Files{}

	defer func() {
		// if we weren't able to create both versions, clean up the leftovers before returning the error
		if err != nil {
			runutil.CaptureError(&err, f.TearDown)
		}
	}()

	f.From, err = NewVersion(from)
	if err != nil {
		return nil, err
	}
	f.To, err = NewVersion(to)
	if err != nil {
		return nil, err
	}

	return f, nil
}

// TearDown removes any temporary directories held by this handle.
func (f Files) TearDown() error {
	return multierror.Append(nil, f.From.TearDown(), f.To.TearDown()).ErrorOrNil()
}

// Version is a handle for a directory that can hold multiple files of a single version to compare.
type Version struct {
	Dir string
}

// NewVersion creates a temporary directory.
func NewVersion(name string) (*Version, error) {
	dir, err := os.MkdirTemp("", name+"-")
	if err != nil {
		return nil, err
	}

	return &Version{
		Dir: dir,
	}, nil
}

// Print prints the given object using the given printer to a file with the specified name in Version.Dir.
func (v *Version) Print(name string, obj runtime.Object, printer printers.ResourcePrinter) (err error) {
	// nolint:gosec // no additional permissions given if file name escapes f.Dir
	file, err := os.OpenFile(filepath.Join(v.Dir, name), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}

	defer runutil.CaptureError(&err, file.Close)
	return printer.PrintObj(obj, file)
}

// TearDown removes the temporary directory held by this handle.
func (v *Version) TearDown() error {
	if v == nil {
		return nil
	}

	return os.RemoveAll(v.Dir)
}
