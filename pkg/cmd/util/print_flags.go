package util

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/printers"
	kubectlget "k8s.io/kubectl/pkg/cmd/get"
	"k8s.io/kubectl/pkg/scheme"

	"github.com/timebertt/kubectl-history/pkg/printer"
)

// PrintFlags composes common output-related flags used in multiple commands.
type PrintFlags struct {
	*genericclioptions.PrintFlags
	CustomColumnsFlags *kubectlget.CustomColumnsPrintFlags
	TableFlags         *TablePrintFlags

	TemplateOnly bool
}

func NewPrintFlags() *PrintFlags {
	return &PrintFlags{
		PrintFlags:         genericclioptions.NewPrintFlags("").WithTypeSetter(scheme.Scheme),
		CustomColumnsFlags: kubectlget.NewCustomColumnsPrintFlags(),
		TableFlags:         NewTablePrintFlags(),
	}
}

func (f *PrintFlags) AllowedFormats() []string {
	formats := f.PrintFlags.AllowedFormats()
	if f.CustomColumnsFlags != nil {
		formats = append(formats, f.CustomColumnsFlags.AllowedFormats()...)
	}
	if f.TableFlags != nil {
		formats = append(formats, f.TableFlags.AllowedFormats()...)
	}
	return formats
}

func (f *PrintFlags) AddFlags(cmd *cobra.Command) {
	f.PrintFlags.AddFlags(cmd)
	f.TableFlags.AddFlags(cmd)
	f.CustomColumnsFlags.AddFlags(cmd)

	cmd.Flags().Lookup("output").Usage = f.OutputUsage()

	cmd.Flags().BoolVar(&f.TemplateOnly, "template-only", f.TemplateOnly, "If false, print the full revision object (e.g., ReplicaSet) instead of only the pod template.")
}

// OutputUsage returns the descriptions for the --output flag based on what is available for the active command.
func (f *PrintFlags) OutputUsage() string {
	usage := fmt.Sprintf("Output format. One of: (%s).", strings.Join(f.AllowedFormats(), ", "))

	// Usage hint for additional formats
	usage += " See "
	if f.CustomColumnsFlags != nil {
		usage += "custom columns [https://kubernetes.io/docs/reference/kubectl/#custom-columns], "
	}
	usage += "golang template [http://golang.org/pkg/text/template/#pkg-overview] and jsonpath template [https://kubernetes.io/docs/reference/kubectl/jsonpath/]."

	return usage
}

// SetKind sets the Kind option of humanreadable flags
func (f *PrintFlags) SetKind(kind schema.GroupKind) {
	f.TableFlags.SetKind(kind)
}

// ToPrinter returns a printer capable of handling the specified output format.
// History objects (e.g., Revision and Revisions) can be directly passed to the returned printer.
func (f *PrintFlags) ToPrinter() (printers.ResourcePrinter, error) {
	outputFormat := ""
	if f.OutputFormat != nil {
		outputFormat = *f.OutputFormat
	}

	if p, err := f.TableFlags.ToPrinter(outputFormat); !genericclioptions.IsNoCompatiblePrinterError(err) {
		return p, err
	}

	// For the remaining output formats, we need to create a delegating printer that extracts a single runtime.Object or
	// a list from the given revision object/list. This way, the printers from cli-runtime can be reused for history
	// objects.
	revisionPrinter := printer.RevisionPrinter{
		TemplateOnly: f.TemplateOnly,
	}

	if f.CustomColumnsFlags != nil {
		// copy over values of commonly-used flags to custom columns
		if f.TableFlags.NoHeaders != nil {
			f.CustomColumnsFlags.NoHeaders = *f.TableFlags.NoHeaders
		}
		if f.TemplatePrinterFlags.TemplateArgument != nil {
			f.CustomColumnsFlags.TemplateArgument = *f.TemplatePrinterFlags.TemplateArgument
		}

		if p, err := f.CustomColumnsFlags.ToPrinter(outputFormat); !genericclioptions.IsNoCompatiblePrinterError(err) {
			revisionPrinter.Delegate = p
			return revisionPrinter, err
		}
	}

	if p, err := f.PrintFlags.ToPrinter(); !genericclioptions.IsNoCompatiblePrinterError(err) {
		revisionPrinter.Delegate = p
		return revisionPrinter, err
	}

	return nil, genericclioptions.NoCompatiblePrinterError{OutputFormat: &outputFormat, AllowedFormats: f.AllowedFormats()}
}
