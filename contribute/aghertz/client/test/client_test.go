package test

import (
	ahclient "ag-core/contribute/aghertz/client"
	"fmt"
	"log/slog"
	"testing"
)

func TestAgHertzClientHello(t *testing.T) {
	slog.Info("hertz client test")
	param := &ahclient.HertzClientParams{}

	client, err := ahclient.NewHertzClient(param)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("%v", client)

	// usage
	// client.Do(context.Background(), nil, nil)
	// client.Get(context.Background(), nil, "")
	// client.Post(context.Background(), nil, "", nil)

}
