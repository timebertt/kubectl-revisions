package main

import (
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/component-base/cli"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"

	"github.com/timebertt/kubectl-revisions/pkg/cmd"
)

func main() {
	if err := cli.RunNoErrOutput(cmd.NewCommand()); err != nil {
		cmdutil.CheckErr(err)
	}
}
