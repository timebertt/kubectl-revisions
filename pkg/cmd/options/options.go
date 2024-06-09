package options

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/util/templates"
)

func NewCommand(streams genericiooptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "options",
		DisableFlagsInUseLine: true,

		Short: "Print the list of flags inherited by all commands",
		Long:  "Print the list of flags inherited by all commands",

		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(cmd.Usage())
		},
	}

	// The `options` command needs write its output to the `out` stream
	// (typically stdout). Without calling SetOutput here, the Usage()
	// function call will fall back to stderr.
	//
	// See https://github.com/kubernetes/kubernetes/pull/46394 for details.
	cmd.SetOut(streams.Out)
	cmd.SetErr(streams.Out)

	templates.UseOptionsTemplates(cmd)
	return cmd
}
