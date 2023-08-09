package get

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/printers"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/timebertt/kubectl-history/pkg/cmd/util"
	"github.com/timebertt/kubectl-history/pkg/history"
	"github.com/timebertt/kubectl-history/pkg/printer"
	"github.com/timebertt/kubectl-history/pkg/runutil"
)

type Options struct {
	genericclioptions.IOStreams

	Namespace  string
	Revision   int64
	PrintFlags *util.PrintFlags
}

func NewOptions(streams genericclioptions.IOStreams) *Options {
	return &Options{
		IOStreams:  streams,
		PrintFlags: util.NewPrintFlags(),
	}
}

func NewCommand(f util.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	o := NewOptions(streams)

	cmd := &cobra.Command{
		Use:   "get (TYPE[.VERSION][.GROUP] NAME | TYPE[.VERSION][.GROUP]/NAME)",
		Short: "Get the history of a workload resource",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			cmdutil.CheckErr(o.Complete(f))
			cmdutil.CheckErr(o.Validate())
			cmdutil.CheckErr(o.Run(ctx, f, args))

			return nil
		},
	}

	o.PrintFlags.AddFlags(cmd.Flags(), "Print")

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

	if o.Revision != 0 {
		// select a single revision
		rev, err := revs.ByNumber(o.Revision)
		if err != nil {
			return err
		}
		revs = history.Revisions{rev}
	}

	// print revisions table
	w := printers.GetNewTabWriter(o.Out)
	defer runutil.CaptureError(&err, w.Flush)

	p := printer.TablePrinter{
		Columns: printer.DefaultTableColumns,
	}
	if err := p.PrintHeaders(w); err != nil {
		return err
	}

	for _, r := range revs {
		if err := p.Print(r, w); err != nil {
			return err
		}
	}

	return nil
}
