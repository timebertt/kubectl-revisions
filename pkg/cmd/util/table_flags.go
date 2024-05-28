package util

import (
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/utils/pointer"

	"github.com/timebertt/kubectl-revisions/pkg/printer"
)

type TablePrintFlags struct {
	NoHeaders    *bool
	ShowLabels   *bool
	ColumnLabels []string
	ShowKind     *bool

	Kind schema.GroupKind
}

func NewTablePrintFlags() *TablePrintFlags {
	return &TablePrintFlags{
		NoHeaders:    pointer.Bool(false),
		ShowLabels:   pointer.Bool(false),
		ColumnLabels: []string{},
		ShowKind:     pointer.Bool(false),
	}
}

func (f *TablePrintFlags) AllowedFormats() []string {
	return []string{"wide"}
}

// SetKind sets the Kind option
func (f *TablePrintFlags) SetKind(kind schema.GroupKind) {
	f.Kind = kind
}

func (f *TablePrintFlags) ToPrinter(outputFormat string) (printers.ResourcePrinter, error) {
	if f == nil || (len(outputFormat) > 0 && outputFormat != "wide") {
		return nil, genericclioptions.NoCompatiblePrinterError{Options: f, AllowedFormats: f.AllowedFormats()}
	}

	p := printers.NewTablePrinter(printers.PrintOptions{
		Kind:         f.Kind,
		NoHeaders:    pointer.BoolDeref(f.NoHeaders, false),
		ShowLabels:   pointer.BoolDeref(f.ShowLabels, false),
		ColumnLabels: f.ColumnLabels,
		WithKind:     pointer.BoolDeref(f.ShowKind, false),
		Wide:         outputFormat == "wide",
	})

	return printer.RevisionsToTablePrinter{
		Delegate: p,
		Columns:  printer.DefaultTableColumns,
	}, nil
}

func (f *TablePrintFlags) AddFlags(cmd *cobra.Command) {
	if f == nil {
		return
	}

	if f.NoHeaders != nil {
		cmd.Flags().BoolVar(f.NoHeaders, "no-headers", *f.NoHeaders, "When using the default output format, don't print headers (default print headers).")
	}
	if f.ShowLabels != nil {
		cmd.Flags().BoolVar(f.ShowLabels, "show-labels", *f.ShowLabels, "When printing, show all labels as the last column (default hide labels column)")
	}
	if f.ColumnLabels != nil {
		cmd.Flags().StringSliceVarP(&f.ColumnLabels, "label-columns", "L", f.ColumnLabels, "Accepts a comma separated list of labels that are going to be presented as columns. Names are case-sensitive. You can also use multiple flag options like -L label1 -L label2...")
	}
	if f.ShowKind != nil {
		cmd.Flags().BoolVar(f.ShowKind, "show-kind", *f.ShowKind, "If present, list the resource type for the requested object(s).")
	}
}
