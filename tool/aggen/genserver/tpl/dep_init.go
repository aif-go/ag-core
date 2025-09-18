package tpl

import (
	"ag-core/tool/aggen/types"
)

func init() {
	ok := types.AddGlobalDependencys(
	// "kitex", "github.com/cloudwego/kitex",
	// "client", "github.com/cloudwego/kitex/client",
	// "server", "github.com/cloudwego/kitex/server",
	// "callopt", "github.com/cloudwego/kitex/client/callopt",
	// "frugal", "github.com/cloudwego/frugal",
	// "fieldmask", "github.com/cloudwego/thriftgo/fieldmask",
	)
	if !ok {
		panic("AddGlobalDependencys failed")
	}
}
