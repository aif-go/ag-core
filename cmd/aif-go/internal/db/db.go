package db

import (
	"github.com/frochyzhang/ag-core/cmd/aif-go/internal/db/dao"
	"github.com/spf13/cobra"
)

// CmdDb represents the db command.
var CmdDb = &cobra.Command{
	Use:   "db",
	Short: "Generate the db files",
	Long:  "Generate the db files.",
}

func init() {
	CmdDb.AddCommand(dao.CmdDao)
}
