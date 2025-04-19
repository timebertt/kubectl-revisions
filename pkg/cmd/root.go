package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/client-go/rest"
	"k8s.io/component-base/logs"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	utilcomp "k8s.io/kubectl/pkg/util/completion"

	"github.com/timebertt/kubectl-revisions/pkg/cmd/completion"
	"github.com/timebertt/kubectl-revisions/pkg/cmd/diff"
	"github.com/timebertt/kubectl-revisions/pkg/cmd/get"
	"github.com/timebertt/kubectl-revisions/pkg/cmd/help"
	"github.com/timebertt/kubectl-revisions/pkg/cmd/options"
	"github.com/timebertt/kubectl-revisions/pkg/cmd/util"
	"github.com/timebertt/kubectl-revisions/pkg/cmd/version"
)

type Options struct {
	genericiooptions.IOStreams

	ConfigFlags *genericclioptions.ConfigFlags
}

func NewOptions() *Options {
	ioStreams := genericiooptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}
	return &Options{
		IOStreams:   ioStreams,
		ConfigFlags: genericclioptions.NewConfigFlags(true).WithDeprecatedPasswordFlag().WithDiscoveryBurst(300).WithDiscoveryQPS(50.0).WithWarningPrinter(ioStreams),
	}
}

func NewCommand() *cobra.Command {
	o := NewOptions()

	cmd := &cobra.Command{
		Use:   "revisions",
		Short: "Time-travel through your workload revision history",

		Annotations: map[string]string{
			// Setting the following annotation makes sure the plugin's help output always has the `kubectl ` command prefix
			// before the plugin's command path to match the expected command path when it is executed via kubectl.
			// This implements https://krew.sigs.k8s.io/docs/developer-guide/develop/best-practices/#help-messages.
			// Changing cmd.Use to `kubectl revisions` makes cobra remove `revisions` from all command paths and use lines.
			cobra.CommandDisplayNameAnnotation: "kubectl revisions",
			help.AnnotationHideFlagsInUsage:    "true",
		},

		PersistentPreRunE: func(*cobra.Command, []string) error {
			warningHandler := rest.NewWarningWriter(o.ErrOut, rest.WarningWriterOptions{Deduplicate: true, Color: printers.AllowsColorOutput(o.ErrOut)})
			rest.SetDefaultWarningHandler(warningHandler)
			return nil
		},

		CompletionOptions: cobra.CompletionOptions{
			// Supporting shell completion for a kubectl plugin requires a dedicated completion executable.
			// Disable cobra's completion command in favor of a custom completion command that explains how to set up
			// completion.
			DisableDefaultCmd: true,
		},
	}

	flags := cmd.PersistentFlags()
	o.ConfigFlags.AddFlags(flags)
	logs.AddFlags(flags)
	f := util.NewFactory(o.ConfigFlags)

	cobra.EnableCommandSorting = false

	// default group
	defaultGroup := &cobra.Group{
		ID:    "default",
		Title: "Available Commands:",
	}
	cmd.AddGroup(defaultGroup)

	for _, subcommand := range []*cobra.Command{
		get.NewCommand(f, o.IOStreams),
		diff.NewCommand(f, o.IOStreams),
	} {
		subcommand.GroupID = defaultGroup.ID
		cmd.AddCommand(subcommand)
	}

	// other group
	otherGroup := &cobra.Group{
		ID:    "other",
		Title: "Other Commands:",
	}
	cmd.AddGroup(otherGroup)

	for _, subcommand := range []*cobra.Command{
		completion.NewCommand(o.IOStreams),
		version.NewCommand(o.IOStreams),
	} {
		subcommand.GroupID = otherGroup.ID
		cmd.AddCommand(subcommand)
	}

	// help group
	helpGroup := &cobra.Group{
		ID:    "help",
		Title: "Help Commands:",
	}
	cmd.AddGroup(helpGroup)
	cmd.SetHelpCommandGroupID(helpGroup.ID)

	optionsCommand := options.NewCommand(o.IOStreams)
	optionsCommand.GroupID = helpGroup.ID
	cmd.AddCommand(optionsCommand)

	help.CustomizeTemplates(cmd)

	// completion
	utilcomp.SetFactoryForCompletion(f)
	registerCompletionFuncForGlobalFlags(cmd, f)

	return cmd
}

func registerCompletionFuncForGlobalFlags(cmd *cobra.Command, f cmdutil.Factory) {
	cmdutil.CheckErr(cmd.RegisterFlagCompletionFunc(
		"namespace",
		func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return utilcomp.CompGetResource(f, "namespace", toComplete), cobra.ShellCompDirectiveNoFileComp
		}))
	cmdutil.CheckErr(cmd.RegisterFlagCompletionFunc(
		"context",
		func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return utilcomp.ListContextsInConfig(toComplete), cobra.ShellCompDirectiveNoFileComp
		}))
	cmdutil.CheckErr(cmd.RegisterFlagCompletionFunc(
		"cluster",
		func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return utilcomp.ListClustersInConfig(toComplete), cobra.ShellCompDirectiveNoFileComp
		}))
	cmdutil.CheckErr(cmd.RegisterFlagCompletionFunc(
		"user",
		func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return utilcomp.ListUsersInConfig(toComplete), cobra.ShellCompDirectiveNoFileComp
		}))
}
