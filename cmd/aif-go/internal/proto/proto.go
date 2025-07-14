package proto

import (
	"github.com/frochyzhang/ag-core/cmd/aif-go/internal/proto/service"
	"github.com/spf13/cobra"

	"github.com/frochyzhang/ag-core/cmd/aif-go/internal/proto/add"
	"github.com/frochyzhang/ag-core/cmd/aif-go/internal/proto/client"
	"github.com/frochyzhang/ag-core/cmd/aif-go/internal/proto/server"
)

// CmdProto represents the proto command.
var CmdProto = &cobra.Command{
	Use:   "proto",
	Short: "Generate the proto files",
	Long:  "Generate the proto files.",
}

func init() {
	CmdProto.AddCommand(add.CmdAdd)
	CmdProto.AddCommand(client.CmdClient)
	CmdProto.AddCommand(server.CmdServer)
	CmdProto.AddCommand(service.CmdService)
}
