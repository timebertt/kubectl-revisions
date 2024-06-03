package diff

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	utilcomp "k8s.io/kubectl/pkg/util/completion"
	"k8s.io/utils/exec"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/timebertt/kubectl-revisions/pkg/cmd/util"
	"github.com/timebertt/kubectl-revisions/pkg/diff"
	"github.com/timebertt/kubectl-revisions/pkg/history"
	"github.com/timebertt/kubectl-revisions/pkg/runutil"
)

type Options struct {
	genericiooptions.IOStreams

	Namespace  string
	Revisions  []int64
	PrintFlags *util.PrintFlags

	Diff diff.Program
}

func NewOptions(streams genericiooptions.IOStreams) *Options {
	printFlags := util.NewPrintFlags()
	printFlags.WithDefaultOutput("yaml")
	printFlags.TemplateOnly = true
	// disable table and name output formats
	printFlags.TableFlags = nil
	printFlags.CustomColumnsFlags = nil
	printFlags.NamePrintFlags = nil

	return &Options{
		IOStreams:  streams,
		PrintFlags: printFlags,
		Diff:       diff.NewProgram(streams),
	}
}

func NewCommand(f util.Factory, streams genericiooptions.IOStreams) *cobra.Command {
	o := NewOptions(streams)

	cmd := &cobra.Command{
		Use:     "diff (TYPE[.VERSION][.GROUP] NAME | TYPE[.VERSION][.GROUP]/NAME)",
		Aliases: []string{"why"},

		Short: "Compare multiple revisions of a workload resource",
		Long: `Compare multiple revisions of a workload resource (Deployment, StatefulSet, or DaemonSet).
  A.k.a., "Why was my Deployment rolled?"

 The history is based on the ReplicaSets/ControllerRevisions still in the system. I.e., the history is limited by the
configured revisionHistoryLimit.

 By default, the latest two revisions are compared. The --revision flag allows selecting the revisions to compare.

 KUBECTL_EXTERNAL_DIFF environment variable can be used to select your own diff command. Users can use external commands
with params too, example: KUBECTL_EXTERNAL_DIFF="colordiff -N -u"

 By default, the "diff" command available in your path will be run with the "-u" (unified diff) and "-N" (treat absent
files as empty) options.`,
		Example: `  # Find out why the nginx Deployment was rolled: compare the latest two revisions
  kubectl revisions diff deploy nginx
  
  # Compare the first and third revision
  kubectl revisions diff deploy nginx --revision=1,3
  
  # Compare the previous revision and the revision before that
  kubectl revisions diff deploy nginx --revision=-2
  
  # Use a colored external diff program
  KUBECTL_EXTERNAL_DIFF="colordiff -u" kubectl revisions diff deploy nginx
  
  # Use dyff as a rich diff program
  KUBECTL_EXTERNAL_DIFF="dyff between --omit-header" kubectl revisions diff deploy nginx
  
  # Show diff in VS Code
  KUBECTL_EXTERNAL_DIFF="code --diff --wait" kubectl revisions diff deploy nginx
`,

		ValidArgsFunction: utilcomp.SpecifiedResourceTypeAndNameNoRepeatCompletionFunc(f, util.Map(history.SupportedKinds, strings.ToLower)),
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(o.Complete(f))
			cmdutil.CheckErr(o.Validate())
			cmdutil.CheckErr(o.Run(cmd.Context(), f, args))
		},
	}

	o.PrintFlags.AddFlags(cmd)

	cmd.Flags().Int64SliceVarP(&o.Revisions, "revision", "r", nil, "Compare the specified revision with its predecessor. "+
		"Specify -1 for the latest revision, -2 for the one before the latest, etc.\n"+
		"If given twice, compare the specified two revisions. If not given, compare the latest two revisions.")

	return cmd
}

// Complete takes the command arguments and factory and infers any remaining options.
func (o *Options) Complete(f util.Factory) error {
	var err error
	o.Namespace, _, err = f.ToRawKubeConfigLoader().Namespace()

	// default to the latest revision if none is given
	if len(o.Revisions) == 0 {
		o.Revisions = []int64{-1}
	}

	return err
}

// Validate checks the set of flags provided by the user.
func (o *Options) Validate() error {
	if len(o.Revisions) > 2 {
		return fmt.Errorf("expected at maximum 2 revisions, but got %d", len(o.Revisions))
	}

	for _, revision := range o.Revisions {
		if revision == 0 {
			return fmt.Errorf("invalid revision 0")
		}
	}

	return nil
}

// Run performs the diff operation.
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
	if len(revs) == 1 {
		return fmt.Errorf("only 1 revision found for %s/%s", kindString, info.Name)
	}

	// get selected revisions
	var a, b history.Revision
	if a, err = revs.ByNumber(o.Revisions[0]); err != nil {
		return err
	}

	if len(o.Revisions) > 1 {
		if b, err = revs.ByNumber(o.Revisions[1]); err != nil {
			return err
		}
	} else {
		// if only one revision is given, compare it with its predecessor
		if b, err = revs.Predecessor(o.Revisions[0]); err != nil {
			return err
		}
	}

	// a should be older than b
	if a.Number() > b.Number() {
		a, b = b, a
	}

	_, err = fmt.Fprintf(o.ErrOut, "comparing revisions %d and %d of %s/%s\n", a.Number(), b.Number(), kindString, info.Name)
	if err != nil {
		return err
	}

	// prepare files for diff program
	fileName := kindString + "." + info.Namespace + "." + info.Name
	files, err := diff.NewFiles(ToDirName(a), ToDirName(b))
	if err != nil {
		return err
	}
	defer runutil.CaptureError(&err, files.TearDown)

	p, err := o.PrintFlags.ToPrinter()
	if err != nil {
		return err
	}

	// the yaml printer adds a `---` separator starting from the second call to PrintObj
	// call it once to /dev/null to have the separator in both files to compare
	if err = p.PrintObj(&corev1.Namespace{}, io.Discard); err != nil {
		return err
	}

	if err := files.From.Print(fileName, a, p); err != nil {
		return err
	}
	if err := files.To.Print(fileName, b, p); err != nil {
		return err
	}

	// run diff program against prepared files
	if err := o.Diff.Run(files.From.Dir, files.To.Dir); err != nil {
		// don't propagate exit status 1 (signaling a diff) upwards and exit cleanly instead
		// there will always be a diff between revisions, there is no point in checking that
		var exitError exec.ExitError
		if errors.As(err, &exitError) && exitError.ExitStatus() <= 1 {
			return nil
		}
		return err
	}

	return nil
}

// ToDirName returns a name for a directory which the given revision should be written to.
func ToDirName(rev history.Revision) string {
	return fmt.Sprintf("%d-%s", rev.Number(), rev.Name())
}
