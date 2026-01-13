package test

import (
	"ag-core/contribute/agredis"
	"context"
	"fmt"
	"testing"
)

func testCase1(cli agredis.AgRedisClient, t *testing.T) {
	ctx := context.Background()
	pingcmd := cli.Ping(ctx)
	if pingcmd.Err() != nil {
		t.Fatal(pingcmd.Err())
	}

	echocmd := cli.Echo(ctx, "hello")
	if echocmd.Err() != nil {
		t.Fatal(echocmd.Err())
	}
	fmt.Printf("echo: %s\n", echocmd.Val())

	// Increment the counter
	val, err := cli.Incr(ctx, "a").Result()
	if err != nil {
		panic(err)
	}
	fmt.Printf("a: %d\n", val)

}
