package get

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	utilcomp "k8s.io/kubectl/pkg/util/completion"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/timebertt/kubectl-revisions/pkg/cmd/util"
	"github.com/timebertt/kubectl-revisions/pkg/history"
)

type Options struct {
	genericiooptions.IOStreams

	Namespace  string
	Revision   int64
	PrintFlags *util.PrintFlags
}

func NewOptions(streams genericiooptions.IOStreams) *Options {
	printFlags := util.NewPrintFlags()

	return &Options{
		IOStreams:  streams,
		PrintFlags: printFlags,
	}
}

func NewCommand(f util.Factory, streams genericiooptions.IOStreams) *cobra.Command {
	o := NewOptions(streams)

	cmd := &cobra.Command{
		Use:     "get (TYPE[.VERSION][.GROUP] NAME | TYPE[.VERSION][.GROUP]/NAME)",
		Aliases: []string{"list", "ls"},

		Short: "Get the revision history of a workload resource",
		Long: `Get the revision history of a workload resource (Deployment, StatefulSet, or DaemonSet).

The history is based on the ReplicaSets/ControllerRevisions still in the system. I.e., the history is limited by the
configured revisionHistoryLimit.

By default, all revisions are printed as a list. If the --revision flag is given, the selected revision is printed
instead.
`,

		Example: `# Get all revisions of the nginx Deployment
kubectl revisions get deploy nginx

# Print additional columns like the revisions' images
kubectl revisions get deploy nginx -o wide

# Get the latest revision in YAML
kubectl revisions get deploy nginx --revision=-1 -o yaml
`,

		ValidArgsFunction: utilcomp.SpecifiedResourceTypeAndNameNoRepeatCompletionFunc(f, util.Map(history.SupportedKinds, strings.ToLower)),
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(f))
			cmdutil.CheckErr(o.Validate())
			cmdutil.CheckErr(o.Run(cmd.Context(), f, args))
		},
	}

	o.PrintFlags.AddFlags(cmd)

	cmd.Flags().Int64VarP(&o.Revision, "revision", "r", 0, "Print the specified revision instead of getting the entire history. "+
		"Specify -1 for the latest revision, -2 for the one before the latest, etc.")

	return cmd
}

// Complete takes the command arguments and factory and infers any remaining options.
func (o *Options) Complete(f util.Factory) error {
	var err error
	o.Namespace, _, err = f.ToRawKubeConfigLoader().Namespace()
	return err
}

// Validate checks the set of flags provided by the user.
func (o *Options) Validate() error {
	return nil
}

// Run performs the get operation.
func (o *Options) Run(ctx context.Context, f util.Factory, args []string) (err error) {
	r := f.NewBuilder().
		Unstructured().
		NamespaceParam(o.Namespace).DefaultNamespace().
		ResourceTypeOrNameArgs(true, args...).
		SingleResourceType().
		Do()

	if err := r.Err(); err != nil {
		return err
	}

	c, err := f.Client()
	if err != nil {
		return err
	}

	infos, err := r.Infos()
	if err != nil {
		return err
	}
	info := infos[0]
	groupKind := info.Mapping.GroupVersionKind.GroupKind()
	kindString := fmt.Sprintf("%s.%s", strings.ToLower(groupKind.Kind), groupKind.Group)

	// get all revisions for the given object
	hist, err := history.ForGroupKind(c, groupKind)
	if err != nil {
		return err
	}

	revs, err := hist.ListRevisions(ctx, client.ObjectKey{Namespace: info.Namespace, Name: info.Name})
	if err != nil {
		return err
	}
	if len(revs) == 0 {
		return fmt.Errorf("no revisions found for %s/%s", kindString, info.Name)
	}

	o.PrintFlags.SetKind(revs[0].GetObjectKind().GroupVersionKind().GroupKind())
	p, err := o.PrintFlags.ToPrinter()
	if err != nil {
		return err
	}

	if o.Revision != 0 {
		// select a single revision
		rev, err := revs.ByNumber(o.Revision)
		if err != nil {
			return err
		}

		return p.PrintObj(rev, o.Out)
	}

	return p.PrintObj(revs, o.Out)
}
