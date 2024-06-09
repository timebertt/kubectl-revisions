package help

import (
	"bytes"

	sprig "github.com/go-task/slim-sprig/v3"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/kubectl/pkg/util/templates"
	"k8s.io/kubectl/pkg/util/term"
)

const (
	// AnnotationHideFlagsInUsage is an annotation on a cobra command that causes flags to be hidden in the command help.
	AnnotationHideFlagsInUsage = "annotation_command_hide_global_flags"
	// AnnotationHideGlobalFlagsInUsage is an annotation on a cobra command that causes inherited flags to be hidden in
	// the command help.
	AnnotationHideGlobalFlagsInUsage = "annotation_command_hide_global_flags"
)

var (
	customHelpTemplate = `{{ with (or .Long .Short) -}}
Synopsis:
{{ . | trimTrailingWhitespaces | indent 2 }}
{{- end -}}
{{- if or .Runnable .HasSubCommands -}}
{{ .UsageString | trimTrailingWhitespaces}}
{{- end }}
`

	customUsageTemplate = `{{if .HasExample }}

Examples:
{{.Example | trim | indent 2}}
{{- end}}

{{- if .HasAvailableSubCommands}}
{{- $cmds := .Commands}}
{{- if eq (len .Groups) 0}}
Available Commands:
{{- range $cmds}}
	{{- if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}
	{{- end}}
{{- end}}
{{- else}}
{{- range $group := .Groups}}

{{.Title}}{{range $cmds}}{{if (and (eq .GroupID $group.ID) (or .IsAvailableCommand (eq .Name "help")))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if not .AllChildCommandsHaveGroup}}

Additional Commands:{{range $cmds}}{{if (and (eq .GroupID "") (or .IsAvailableCommand (eq .Name "help")))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{end}}{{end}}

{{- if and .HasAvailableLocalFlags (not .Annotations.` + AnnotationHideFlagsInUsage + `)}}

Flags:
{{ flagsUsages .LocalFlags | trimTrailingWhitespaces}}
{{- end}}

{{- if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}
{{- end}}

{{- if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}
{{- end}}

Usage:
{{- if .Runnable}}
  {{.UseLine}}
{{ end}}
{{- if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]
{{ end}}

{{- if .HasAvailableSubCommands}}
Use "{{.CommandPath}} [command] --help" for more information about a command.
{{- end}}
{{- if and .HasAvailableInheritedFlags (not .Annotations.` + AnnotationHideGlobalFlagsInUsage + `)}}
Use "kubectl revisions options" for a list of global command-line options (applies to all commands).
{{- end}}
`
)

// CustomizeTemplates customizes the help and usage templates of the command.
func CustomizeTemplates(cmd *cobra.Command) {
	// add more helpful template funcs like `indent`
	cobra.AddTemplateFuncs(sprig.HermeticTxtFuncMap())
	cobra.AddTemplateFunc("flagsUsages", flagsUsages)

	cmd.SetHelpTemplate(customHelpTemplate)
	cmd.SetUsageTemplate(customUsageTemplate)
}

// flagsUsages will print out the kubectl help flags
func flagsUsages(fs *pflag.FlagSet) (string, error) {
	flagBuf := new(bytes.Buffer)
	wrapLimit, err := term.GetWordWrapperLimit()
	if err != nil {
		wrapLimit = 0
	}
	printer := templates.NewHelpFlagPrinter(flagBuf, wrapLimit)

	fs.VisitAll(func(flag *pflag.Flag) {
		if flag.Hidden {
			return
		}
		printer.PrintHelpFlag(flag)
	})

	return flagBuf.String(), nil
}
