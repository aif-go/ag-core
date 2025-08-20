package main

import (
	"ag-core/cmd/aif-go/internal/db"
	"log"

	"github.com/spf13/cobra"

	"ag-core/cmd/aif-go/internal/change"
	"ag-core/cmd/aif-go/internal/project"
	"ag-core/cmd/aif-go/internal/proto"
	"ag-core/cmd/aif-go/internal/run"
	"ag-core/cmd/aif-go/internal/upgrade"
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
