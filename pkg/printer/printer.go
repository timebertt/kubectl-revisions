package printer

import (
	"io"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/yaml"

	"github.com/timebertt/kubectl-history/pkg/history"
)

// Printer can print revisions.
type Printer interface {
	// Print prints a revision to the given writer in the printer's format.
	Print(rev history.Revision, w io.Writer) error
}

// Encoder can encode arbitrary (partial) objects.
type Encoder interface {
	// Encode encodes the given (partial) object to the given writer.
	Encode(obj runtime.Object, w io.Writer) error
}

var _ Encoder = YAMLEncoder{}

// YAMLEncoder encodes objects using yaml.Marshal.
type YAMLEncoder struct{}

func (y YAMLEncoder) Encode(obj runtime.Object, w io.Writer) error {
	if obj == nil {
		return nil
	}

	data, err := yaml.Marshal(obj)
	if err != nil {
		return err
	}

	_, err = w.Write(data)
	return err
}
