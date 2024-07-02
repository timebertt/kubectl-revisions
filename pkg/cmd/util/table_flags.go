package util

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/utils/pointer"

	"github.com/timebertt/kubectl-revisions/pkg/printer"
)

type TablePrintFlags struct {
	NoHeaders    *bool
	ShowLabels   *bool
	ColumnLabels []string

	WithNamespace bool
}

func NewTablePrintFlags() *TablePrintFlags {
	return &TablePrintFlags{
		NoHeaders:    pointer.Bool(false),
		ShowLabels:   pointer.Bool(false),
		ColumnLabels: []string{},
	}
}

func (f *TablePrintFlags) AllowedFormats() []string {
	return []string{"wide"}
}

func (f *TablePrintFlags) SetWithNamespace() {
	f.WithNamespace = true
}

func (f *TablePrintFlags) ToPrinter(outputFormat string) (printers.ResourcePrinter, error) {
	if f == nil || (len(outputFormat) > 0 && outputFormat != "wide") {
		return nil, genericclioptions.NoCompatiblePrinterError{Options: f, AllowedFormats: f.AllowedFormats()}
	}

	p := printers.NewTablePrinter(printers.PrintOptions{
		NoHeaders:     pointer.BoolDeref(f.NoHeaders, false),
		ShowLabels:    pointer.BoolDeref(f.ShowLabels, false),
		ColumnLabels:  f.ColumnLabels,
		Wide:          outputFormat == "wide",
		WithNamespace: f.WithNamespace,
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
}
