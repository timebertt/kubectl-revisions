package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra/doc"

	"github.com/timebertt/kubectl-revisions/pkg/cmd"
)

func main() {
	if len(os.Args) < 2 {
		panic("must provide target directory as argument")
	}
	dir := os.Args[1]

	if err := doc.GenMarkdownTree(cmd.NewCommand(), dir); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
