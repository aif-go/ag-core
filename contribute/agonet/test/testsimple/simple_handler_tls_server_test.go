package testsimple

import (
	"ag-core/contribute/agonet"
	"ag-core/contribute/agonet/simple"
	"ag-core/contribute/agonet/simple/utils"
	"fmt"
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	"strings"
	"testing"
	"time"

	"github.com/petermattis/goid"
)

func TestSimpleTlsServerHandler(t *testing.T) {
	// slog.SetLogLoggerLevel(slog.LevelDebug)

	// 启动 http pprof
	go func() {
		http.ListenAndServe(":6060", nil)
	}()

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
		// hexStr := hex.EncodeToString(msg)
		gid := goid.Get()
		// fmt.Printf("msg: %s, len: %d, gid: %d\n", string(msg), len(msg), gid)

		replymsg := fmt.Sprintf("R_%s_s:%d", msg, gid)
		fmt.Println(replymsg)

		// 用于测试异常处理
		if strings.HasPrefix(string(msg), "testerr") {
			utils.Assert(fmt.Errorf("test error"))
			return
		}
		// replymsg := msg
		// time.Sleep(time.Second)

		ctx.Channel().Write(replymsg)

		// ctx.Write([]byte(replymsg))
	})

	// lengthDecod := lengthDecod.NewLengthFieldCodec(binary.BigEndian, 1024, 0, 4, 0, 0)
	// lengthDecod := simple.NewLengthFieldDecoder(nil, 10, 0, 2, 0, 2)

	lengthDecod := simple.NewLengthFieldDecoder(nil, 1024, 0, 2, 0, 2)
	lengthEncod := simple.NewLengthFieldEncoder(nil, 2, 0, false)

	// lengthDecod := simple.NewLengthFieldStrDecoder(1024, 0, 6, 0, 6)
	// lengthEncod := simple.NewLengthFieldStrEncoder(6, 0, false)

	// lengthDecod := simple.NewLengthFieldDecoder(nil, 1024, 0, 6, 0, 6)
	// lengthEncod := simple.NewLengthFieldEncoder(nil, 6, 0, false)

	custCodec := simple.NewSimpleCodec(
		"custCodec",
		func(msg []byte) (out []any, err error) {
			// fmt.Println("custdecode msg:", string(msg))
			out = append(out, msg)
			// out = append(out, string(msg))
			return out, nil

		},
		func(msg string) ([]any, error) {
			// fmt.Println("custencode msg:", string(msg))
			return []any{msg}, nil
		},
	)

	custCodec2 := &simple.SimpleCodec[string, string]{
		CodeName: "custCodec2",
		Decode: func(msg string) ([]any, error) {
			// fmt.Println("custdecode2 msg:", msg)
			return []any{msg}, nil
		},
		Encode: func(msg string) ([]any, error) {
			// fmt.Println("custencode2 msg:", msg)
			return []any{msg}, nil
		},
	}

	// 通道激活事件
	activeHand := simple.ActiveHandlerFunc(func(ctx simple.ActiveContext) {
		fmt.Printf("test active, remote addr: %s\n", ctx.Channel().RemoteAddr())
		ctx.FireActive()
	})

	// 通道非激活事件
	inactiveHand := simple.InactiveHandlerFunc(func(ctx simple.InactiveContext, ex error) {
		fmt.Printf("test inactive, remote addr: %s, reason: %v\n", ctx.Channel().RemoteAddr(), ex)
		ctx.FireInactive(ex)
	})

	// idleHandler := simple.IdleStateHandler(3, 0, 0, time.Second)
	idleHandler := simple.IdleStateHandler(0, 0, 3, time.Second)

	eventHandler := simple.EventHandlerFunc(func(ctx simple.EventContext, event any) {
		if ie, ok := event.(simple.IdleStateEvent); ok {
			slog.Info(fmt.Sprintf("idle state:%s, first:%v", ie.State, ie.First))
		}
		ctx.FireEvent(event)
	})
	eventHandler2 := simple.EventHandlerFunc(func(ctx simple.EventContext, event any) {
		if ie, ok := event.(simple.IdleStateEvent); ok {
			slog.Info(fmt.Sprintf("idle2 state:%s, first:%v", ie.State, ie.First))
		}
		ctx.FireEvent(event)
	})

	pipelineInitializer := func(c simple.Channel) error {
		c.Pipeline().
			// AddLast(&echoHandler{}).
			AddLast(
				lengthDecod,
				lengthEncod,
				custCodec,
				custCodec2,
				testHand,
				activeHand,
				inactiveHand,
				idleHandler,
				eventHandler,
				eventHandler2,
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
