package completion

import (
	_ "embed"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"

	"github.com/timebertt/kubectl-revisions/pkg/cmd/help"
)

//go:embed kubectl_complete-revisions
var completionScript []byte

type Options struct {
	genericiooptions.IOStreams
}

func NewOptions(streams genericiooptions.IOStreams) *Options {
	return &Options{
		IOStreams: streams,
	}
}

func NewCommand(streams genericiooptions.IOStreams) *cobra.Command {
	o := NewOptions(streams)

	cmd := &cobra.Command{
		Use:                   "completion",
		DisableFlagsInUseLine: true,

		Annotations: map[string]string{
			help.AnnotationHideGlobalFlagsInUsage: "true",
		},

		Short: "Setup shell completion",
		Long: `The completion command outputs a script which makes the revisions plugin's completion available to kubectl's completion
(supported in kubectl v1.26+), see https://github.com/kubernetes/kubernetes/pull/105867 and
https://github.com/kubernetes/sample-cli-plugin#shell-completion.

This script needs to be installed as an executable file in PATH named kubectl_complete-revisions. E.g., you could
install it in krew's binary directory. This is not supported natively yet, but can be done manually as follows
(see https://github.com/kubernetes-sigs/krew/issues/812):
` + "```" + `
SCRIPT="${KREW_ROOT:-$HOME/.krew}/bin/kubectl_complete-revisions"; kubectl revisions completion > "$SCRIPT" && chmod +x "$SCRIPT"
` + "```" + `

If you don't use krew, you can install the script next to the binary itself as follows:
` + "```" + `
SCRIPT="$(dirname "$(which kubectl-revisions)")/kubectl_complete-revisions"; kubectl revisions completion > "$SCRIPT" && chmod +x "$SCRIPT"
` + "```" + `

Alternatively, you can also use https://github.com/marckhouzam/kubectl-plugin_completion to generate completion
scripts for this plugin along with other kubectl plugins that support it.
`,

		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Run())
		},
	}

	return cmd
}

// Run outputs the completion script.
func (o *Options) Run() error {
	_, err := o.Out.Write(completionScript)
	return err
}
