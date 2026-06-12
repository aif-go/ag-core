package main

import (
	"github.com/aif-go/ag-core/tool/cmd/aggo/internal/rootcmd"
	"log"

	_ "github.com/aif-go/ag-core/tool/cmd/aggo/internal/init"
)

func main() {

	// rootcmd.SetVersion(release)

	err := rootcmd.Run()
	if err != nil {
		log.Fatal(err)
	}

}
