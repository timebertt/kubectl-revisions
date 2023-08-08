package util

import (
	"fmt"

	"github.com/spf13/pflag"

	"github.com/timebertt/kubectl-history/pkg/printer"
)

// PrintFlags composes common output-related flags used in multiple commands.
type PrintFlags struct {
	TemplateOnly      bool
	ShowManagedFields bool
}

func NewPrintFlags() *PrintFlags {
	return &PrintFlags{
		TemplateOnly:      true,
		ShowManagedFields: false,
	}
}

func (f *PrintFlags) AddFlags(flags *pflag.FlagSet, operation string) {
	flags.BoolVar(&f.TemplateOnly, "template-only", f.TemplateOnly, operation+" only the revision's pod template instead of the full revision object.")
	flags.BoolVar(&f.ShowManagedFields, "show-managed-fields", f.ShowManagedFields, operation+" also the revision object's managedFields.")
}

func (f *PrintFlags) Validate() error {
	if f.ShowManagedFields && f.TemplateOnly {
		return fmt.Errorf("--show-managed-fields option can only be used with --template-only=false")
	}

	return nil
}

func (f *PrintFlags) ToPrinter() printer.Printer {
	return printer.ObjectPrinter{
		Encoder:           printer.YAMLEncoder{},
		TemplateOnly:      f.TemplateOnly,
		ShowManagedFields: f.ShowManagedFields,
	}
}
