package testsimple

import (
	"ag-core/contribute/agonet"
	"ag-core/contribute/agonet/simple"
	"encoding/hex"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"testing"
)

func TestServerHandler(t *testing.T) {
	addr := "tcp://:9000"

	// 启动 http pprof
	go func() {
		http.ListenAndServe(":6060", nil)
	}()

	// var testHand = &simplehandler.SimpleHandler[[]byte]{
	var testHand = simple.NewSimpleInboundHandler(func(ctx simple.InboundContext, msg []byte) {
		hexStr := hex.EncodeToString(msg)
		fmt.Printf("msg: %s, len: %d, hexStr: %s\n", string(msg), len(msg), hexStr)

		replymsg := fmt.Sprintf("reply:%s", msg)
		// replymsg := msg

		ctx.Channel().Write([]byte(replymsg))

		// ctx.Write([]byte(replymsg))
	})

	// lengthDecod := lengthDecod.NewLengthFieldCodec(binary.BigEndian, 1024, 0, 4, 0, 0)
	lengthDecod := simple.NewLengthFieldDecoder(nil, 1024, 0, 2, 0, 2)
	// lengthDecod := simple.NewLengthFieldDecoder(nil, 1024, 0, 2, 0, 0)
	lengthEncod := simple.NewLengthFieldEncoder(nil, 2, 0, false)

	custCodec := simple.NewSimpleCodec(
		"custCodec",
		func(msg []byte) (out []any, err error) {
			fmt.Println("custdecode msg:", string(msg))
			return []any{msg}, nil
		},
		func(msg []byte) ([]any, error) {
			fmt.Println("custencode msg:", string(msg))
			return []any{msg}, nil
		},
	)

	custCodec2 := &simple.SimpleCodec[[]byte, []byte]{
		CodeName: "custCodec2",
		Encode: func(msg []byte) ([]any, error) {
			return []any{msg}, nil
		},
		Decode: func(msg []byte) ([]any, error) {
			return []any{msg}, nil
		},
	}

	// 通道激活事件
	activeHand := simple.ActiveHandlerFunc(func(ctx simple.ActiveContext) {
		fmt.Printf("test active, remote addr: %s\n", ctx.Channel().RemoteAddr())
	})

	// 通道非激活事件
	inactiveHand := simple.InactiveHandlerFunc(func(ctx simple.InactiveContext, ex error) {
		fmt.Printf("test inactive, remote addr: %s, reason: %v\n", ctx.Channel().RemoteAddr(), ex)
	})

	pipelineInitializer := func(c simple.Channel) error {
		c.Pipeline().
			// AddLast(&echoHandler{}).
			AddLast(
				activeHand,
				inactiveHand,
				lengthDecod,
				lengthEncod,
				custCodec,
				custCodec2,
				testHand,
			)

		// c.Pipeline().AddLast(testHand)
		return nil
	}

	evenhand, err := simple.NewSimpleEventHandlerWithOptions(
		simple.WithChannelInitializer(pipelineInitializer),
	)
	if err != nil {
		t.Fatalf("NewSimpleEventHandlerWithOptions failed: %v", err)
	}

	server, err := agonet.NewServer(evenhand, &agonet.ServerConfig{
		Addr: addr,
		Config: agonet.OptionsConfig{
			Engine: agonet.EngineConfig{
				NumEventLoop: 1,
			},
		},
	})
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	err = server.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
}
