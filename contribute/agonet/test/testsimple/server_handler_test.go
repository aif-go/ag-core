package testsimple

import (
	"ag-core/contribute/agonet"
	"ag-core/contribute/agonet/simple"
	"ag-core/contribute/agonet/simple/codec"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	"testing"
	"time"
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

type echoHandler struct {
}

func (e echoHandler) HandleRead(ctx simple.InboundContext, message simple.Message) {
	fmt.Println("==== echo HandleRead ====", time.Now().UnixNano())
	reader, ok := message.(agonet.Reader)
	msg := ""
	if ok {
		// go func() {
		buf, err := reader.Next(-1)
		if err != nil {
			slog.Error("Read failed", "err", err)
		}
		msg = string(buf)
		// fmt.Println(string(msg))
		slog.Info("read", "msg", msg)
	}
	// fmt.Println("read: ", ctx.Channel().ID(), message, " isActive: ", ctx.Channel().IsActive())

	ctx.FireRead(msg)
}

func (e echoHandler) HandleWrite(ctx simple.OutboundContext, message simple.Message) {
	fmt.Println("==== HandleWrite ====")
	// fmt.Println("write: ", ctx.Channel().ID(), message, " isActive: ", ctx.Channel().IsActive())
	ctx.FireWrite(message)
}

func (e echoHandler) HandleException(ctx simple.ExceptionContext, ex simple.Exception) {
	fmt.Println("exception: ", ctx.Channel().ID(), ex, " isActive: ", ctx.Channel().IsActive())
	ctx.Channel().Close(ex)
}
