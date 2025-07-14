package main

import (
	"github.com/frochyzhang/ag-core/cmd/aif-go/internal/db"
	"log"

	"github.com/spf13/cobra"

	"github.com/frochyzhang/ag-core/cmd/aif-go/internal/change"
	"github.com/frochyzhang/ag-core/cmd/aif-go/internal/project"
	"github.com/frochyzhang/ag-core/cmd/aif-go/internal/proto"
	"github.com/frochyzhang/ag-core/cmd/aif-go/internal/run"
	"github.com/frochyzhang/ag-core/cmd/aif-go/internal/upgrade"
)

var rootCmd = &cobra.Command{
	Use:     "aif-go",
	Short:   "aif-go: An elegant toolkit for Go microservices.",
	Long:    `aif-go: An elegant toolkit for Go microservices.`,
	Version: release,
}

func init() {
	rootCmd.AddCommand(project.CmdNew)
	rootCmd.AddCommand(proto.CmdProto)
	rootCmd.AddCommand(db.CmdDb)
	rootCmd.AddCommand(upgrade.CmdUpgrade)
	rootCmd.AddCommand(change.CmdChange)
	rootCmd.AddCommand(run.CmdRun)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
