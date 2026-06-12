package cmd_upgrade

import (
	"github.com/aif-go/ag-core/tool/cmd/aggo/internal/base"
	"github.com/aif-go/ag-core/tool/cmd/aggo/internal/rootcmd"
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

var (
	baseMode string
)

func init() {
	rootcmd.RegCommand(CmdUpgrade)

	baseMode = "ag-core"
	CmdUpgrade.Flags().StringVarP(&baseMode, "base-mode", "b", baseMode, "base mode")
}

// CmdUpgrade represents the upgrade command.
var CmdUpgrade = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade the aggo tools",
	Long:  "Upgrade the aggo tools. Example: aggo upgrade",
	Run:   Run,
}

// var basemodulename = "ag-core"

// Run upgrade the aggo tools.
func Run(_ *cobra.Command, _ []string) {
	fmt.Println("aggo upgrade")

	err := base.GoInstall(
		"google.golang.org/protobuf/cmd/protoc-gen-go@latest",
		// "google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest",
		// "github.com/google/gnostic/cmd/protoc-gen-openapi@latest",
		"github.com/google/gnostic/cmd/protoc-gen-openapi@v0.6.9",
		// fmt.Sprintf("%s/tool/cmd/protoc-gen-go-aggenall@latest", baseMode),
		fmt.Sprintf("%s/tool/cmd/protoc-gen-go-agserver@latest", baseMode),
		fmt.Sprintf("%s/tool/cmd/protoc-gen-go-agservice@latest", baseMode),
		fmt.Sprintf("%s/tool/cmd/protoc-gen-go-aghertz@latest", baseMode),
		fmt.Sprintf("%s/tool/cmd/protoc-gen-go-agkitex@latest", baseMode),
		fmt.Sprintf("%s/tool/cmd/protoc-gen-go-agapi@latest", baseMode),
	)
	if err != nil {
		slog.Error("upgrade failed", "err", err)
		os.Exit(1)
	}
}
