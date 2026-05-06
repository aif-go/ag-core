package main

import (
	"ag-core/tool/cmd/aggo/internal/rootcmd"
	"log"

	_ "ag-core/tool/cmd/aggo/internal/init"
)

func main() {

	// rootcmd.SetVersion(release)

	err := rootcmd.Run()
	if err != nil {
		log.Fatal(err)
	}

}
