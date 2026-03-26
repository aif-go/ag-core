package testtls

import (
	"ag-core/contribute/agonet"
	"ag-core/contribute/agonet/simple"
	"ag-core/contribute/agonet/simple/codec"
	"encoding/hex"
	"fmt"
	_ "net/http/pprof"
	"testing"
)

func TestSimpleTlsServerHandler(t *testing.T) {

	evenhand, err := _simpleServerEventHandler()
	if err != nil {
		t.Fatalf("NewSimpleEventHandlerWithOptions failed: %v", err)
	}

	binder, err := buildCfgBinder(_agonet_tls_ser)
	if err != nil {
		t.Fatalf("buildCfgBinder failed: %v", err)
	}
	serverCfg, err := agonet.NewServerConfig(binder)
	if err != nil {
		t.Fatalf("NewServerConfig failed: %v", err)
	}

	server, err := agonet.NewServer(evenhand, serverCfg)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	err = server.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
}

func _simpleServerEventHandler() (agonet.EventHandler, error) {
	// var testHand = &simplehandler.SimpleHandler[[]byte]{
	var testHand = simple.NewSimpleInboundHandler(func(ctx simple.InboundContext, msg []byte) {
		hexStr := hex.EncodeToString(msg)
		fmt.Printf("msg: %s, len: %d, hexStr: %s\n", string(msg), len(msg), hexStr)

		replymsg := fmt.Sprintf("reply:%s", msg)
		// replymsg := msg

		ctx.Channel().Write(replymsg)

		// ctx.Write([]byte(replymsg))
	})

	// lengthDecod := lengthDecod.NewLengthFieldCodec(binary.BigEndian, 1024, 0, 4, 0, 0)
	lengthDecod := codec.NewLengthFieldDecoder(nil, 1024, 0, 2, 0, 2)
	// lengthDecod := codec.NewLengthFieldDecoder(nil, 1024, 0, 2, 0, 0)
	lengthEncod := codec.NewLengthFieldEncoder(nil, 2, 0, false)

	custCodec := codec.NewSimpleCodec(
		"custCodec",
		func(msg []byte) ([][]byte, error) {
			fmt.Println("custdecode msg:", string(msg))
			return [][]byte{msg}, nil
		},
		func(msg []byte) ([]byte, error) {
			fmt.Println("custencode msg:", string(msg))
			return msg, nil
		},
	)

	custCodec2 := &codec.SimpleCodec[[]byte, []byte]{
		CodeName: "custCodec2",
		Encode: func(msg []byte) ([]byte, error) {
			return msg, nil
		},
		Decode: func(msg []byte) ([][]byte, error) {
			return [][]byte{msg}, nil
		},
	}

	pipelineInitializer := func(c simple.Channel) error {
		c.Pipeline().
			// AddLast(&echoHandler{}).
			AddLast(
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
		return nil, err
	}

	return evenhand, nil
}
