package diff

import (
	"k8s.io/cli-runtime/pkg/genericclioptions"
	kubectldiff "k8s.io/kubectl/pkg/cmd/diff"
	"k8s.io/utils/exec"
)

// Program is a diff programm that compares two files.
type Program interface {
	// Run executes the diff programm to compare the given files.
	Run(a, b string) error
}

// NewProgram returns kubectl's default Program implementation that respects the KUBECTL_EXTERNAL_DIFF environment
// variable. It falls back to `diff -u -N` if the env var is unset.
func NewProgram(streams genericclioptions.IOStreams) Program {
	return &kubectldiff.DiffProgram{
		Exec:      exec.New(),
		IOStreams: streams,
	}
}
