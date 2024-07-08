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

	Namespace     string
	AllNamespaces bool

	ChunkSize     int64
	LabelSelector string

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
		Use:     "get (TYPE[.VERSION][.GROUP] [NAME | -l label] | TYPE[.VERSION][.GROUP]/NAME ...)",
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

	cmd.Flags().BoolVarP(&o.AllNamespaces, "all-namespaces", "A", o.AllNamespaces, "If present, list the requested object(s) across all namespaces. Namespace in current context is ignored even if specified with --namespace.")
	cmdutil.AddChunkSizeFlag(cmd, &o.ChunkSize)
	cmdutil.AddLabelSelectorFlagVar(cmd, &o.LabelSelector)

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
		WithScheme(history.Scheme, history.DecodingVersions...).
		NamespaceParam(o.Namespace).DefaultNamespace().AllNamespaces(o.AllNamespaces).
		LabelSelectorParam(o.LabelSelector).
		RequestChunksOf(o.ChunkSize).
		ResourceTypeOrNameArgs(true, args...).
		SingleResourceType().
		Flatten().
		Do()

	if err := r.Err(); err != nil {
		return err
	}

	var singleItemImplied bool
	r.IntoSingleItemImplied(&singleItemImplied)
	if o.Revision != 0 && !singleItemImplied {
		return fmt.Errorf("a revision can only be selected when targeting a single resource")
	}

	c, err := f.Client()
	if err != nil {
		return err
	}

	infos, err := r.Infos()
	if err != nil {
		return err
	}

	if len(infos) == 0 {
		if o.AllNamespaces {
			_, _ = fmt.Fprintf(o.ErrOut, "No resources found.\n")
		} else {
			_, _ = fmt.Fprintf(o.ErrOut, "No resources found in %s namespace.\n", o.Namespace)
		}
		return
	}

	groupKind := infos[0].Mapping.GroupVersionKind.GroupKind()
	kindString := fmt.Sprintf("%s.%s", strings.ToLower(groupKind.Kind), groupKind.Group)

	hist, err := history.ForGroupKind(c, groupKind)
	if err != nil {
		return err
	}

	if o.AllNamespaces {
		o.PrintFlags.SetWithNamespace()
	}
	p, err := o.PrintFlags.ToPrinter()
	if err != nil {
		return err
	}

	var allRevisions history.Revisions
	for _, info := range infos {
		// get all revisions for the given object
		revs, err := hist.ListRevisions(ctx, info.Object.(client.Object))
		if err != nil {
			return err
		}
		if len(revs) == 0 {
			return fmt.Errorf("no revisions found for %s/%s", kindString, info.Name)
		}

		if o.Revision != 0 {
			// select a single revision
			rev, err := revs.ByNumber(o.Revision)
			if err != nil {
				return fmt.Errorf("error for %s/%s: %w", kindString, info.Name, err)
			}

			return p.PrintObj(rev, o.Out)
		} else {
			allRevisions = append(allRevisions, revs...)
		}
	}

	return p.PrintObj(allRevisions, o.Out)
}
