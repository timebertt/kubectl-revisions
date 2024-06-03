package cmd

import (
	"os"
	"regexp"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/client-go/rest"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	utilcomp "k8s.io/kubectl/pkg/util/completion"

	"github.com/timebertt/kubectl-revisions/pkg/cmd/completion"
	"github.com/timebertt/kubectl-revisions/pkg/cmd/diff"
	"github.com/timebertt/kubectl-revisions/pkg/cmd/get"
	"github.com/timebertt/kubectl-revisions/pkg/cmd/util"
	"github.com/timebertt/kubectl-revisions/pkg/cmd/version"
)

type Options struct {
	genericiooptions.IOStreams

	ConfigFlags *genericclioptions.ConfigFlags
}

func NewOptions() *Options {
	return &Options{
		IOStreams:   genericiooptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr},
		ConfigFlags: genericclioptions.NewConfigFlags(true).WithDeprecatedPasswordFlag().WithDiscoveryBurst(300).WithDiscoveryQPS(50.0),
	}
}

func NewCommand() *cobra.Command {
	o := NewOptions()

	cmd := &cobra.Command{
		Use:   "revisions",
		Short: "Time-travel through your workload revision history",

		PersistentPreRunE: func(*cobra.Command, []string) error {
			warningHandler := rest.NewWarningWriter(o.IOStreams.ErrOut, rest.WarningWriterOptions{Deduplicate: true, Color: printers.AllowsColorOutput(o.IOStreams.ErrOut)})
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
	f := util.NewFactory(o.ConfigFlags)

	defaultGroup := &cobra.Group{
		ID:    "default",
		Title: "Available Commands:",
	}
	cmd.AddGroup(defaultGroup)

	for _, subcommand := range []*cobra.Command{
		diff.NewCommand(f, o.IOStreams),
		get.NewCommand(f, o.IOStreams),
	} {
		subcommand.GroupID = defaultGroup.ID
		cmd.AddCommand(subcommand)
	}

	otherGroup := &cobra.Group{
		ID:    "other",
		Title: "Other Commands:",
	}
	cmd.AddGroup(otherGroup)
	cmd.SetHelpCommandGroupID(otherGroup.ID)

	for _, subcommand := range []*cobra.Command{
		completion.NewCommand(o.IOStreams),
		version.NewCommand(o.IOStreams),
	} {
		subcommand.GroupID = otherGroup.ID
		cmd.AddCommand(subcommand)
		hideGlobalFlagsInUsage(cmd)
	}

	customizeUsageTemplate(cmd)

	utilcomp.SetFactoryForCompletion(f)
	registerCompletionFuncForGlobalFlags(cmd, f)

	return cmd
}

// customizeUsageTemplate makes sure the plugin's help output always has the `kubectl ` command prefix before the
// plugin's command path to match the expected command path when it is executed via kubectl.
// This implements https://krew.sigs.k8s.io/docs/developer-guide/develop/best-practices/#help-messages.
// I.e., the default template would output:
//
//	Usage:
//	  revisions [command]
//
// The modified template outputs:
//
//	Usage:
//	  kubectl revisions [command]
//
// Changing cmd.Use to `kubectl revisions` makes cobra remove `revisions` from all command paths and use lines.
func customizeUsageTemplate(cmd *cobra.Command) {
	defaultTmpl := cmd.UsageTemplate()

	r := regexp.MustCompile(`([{ ])(.CommandPath|.UseLine)([} ])`)
	tmpl := r.ReplaceAllString(defaultTmpl, `$1(printf "kubectl %s" $2)$3`)

	cmd.SetUsageTemplate(tmpl)
}

// hideGlobalFlagsInUsage customizes the help output of subcommands to skip the global flags section.
// The function should be called after adding the subcommand to the parent command, otherwise the customization from
// customizeUsageTemplate will be lost.
func hideGlobalFlagsInUsage(cmd *cobra.Command) {
	defaultTmpl := cmd.UsageTemplate()

	r := regexp.MustCompile(`([{ ]).HasAvailableInheritedFlags([} ])`)
	tmpl := r.ReplaceAllString(defaultTmpl, `${1}false${2}`)

	cmd.SetUsageTemplate(tmpl)
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
