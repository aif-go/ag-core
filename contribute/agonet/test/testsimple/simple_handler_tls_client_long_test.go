package testsimple

import (
	"ag-core/contribute/agonet"
	"ag-core/contribute/agonet/simple"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/petermattis/goid"
)

func TestSimpleTlsClientLongHandler(t *testing.T) {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	addr := "localhost:8443"

	// 启动 http pprof
	go func() {
		http.ListenAndServe(":6061", nil)
	}()

	binder, err := buildCfgBinder(_agonet_tls_cli)
	if err != nil {
		t.Fatalf("buildCfgBinder failed: %v", err)
	}

	cliCfg, err := agonet.NewClientConfig(binder)
	if err != nil {
		t.Fatalf("NewClientConfig failed: %v", err)
	}

	eventhandler, err := _simpleLongEventHandler()
	if err != nil {
		t.Fatalf("LoadAuthCert failed: %v", err)
	}

	client, err := agonet.NewClient(eventhandler, cliCfg)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	err = client.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	var do = func(channel simple.Channel, i int64) {
		// 使用channel发送数据
		gid := goid.Get()

		// 获取channel的指针地址

		msg := fmt.Sprintf("ping_i:%d_c:%d", i, gid)
		fmt.Printf("[%p]📡发: %s\n", channel, msg)

		channel.Write(msg)
		time.Sleep(time.Second)

	}

	wg := sync.WaitGroup{}
	for i := 0; i < 2; i++ {
		wg.Add(1)

		// 创建链接
		tcon, err := client.Dial("tcp", addr)
		if err != nil {
			t.Fatalf("Dial failed: %v", err)
		}

		// 从链接中获取channel
		channel, err := simple.ChannelFromConn(tcon)
		if err != nil {
			t.Fatalf("ChannelForConn failed: %v", err)
		}

		go func() {
			// func() {
			i := int64(0)

			for {
				i++
				do(channel, i)
			}

			time.Sleep(time.Second)
			channel.Close(nil)
			defer wg.Done()
		}()
	}

	wg.Wait()

}

func _simpleLongEventHandler() (agonet.EventHandler, error) {
	var testHand = simple.NewSimpleInboundHandler(func(ctx simple.InboundContext, msg []byte) {
		channel := ctx.Channel()
		fmt.Printf("[%p]📥收: %s\n", channel, string(msg))
		ctx.FireRead(msg)
	})
	lengthDecod := simple.NewLengthFieldDecoder(nil, 1024, 0, 2, 0, 2)
	lengthEncod := simple.NewLengthFieldEncoder(nil, 2, 0, false)
	pipelineInitializer := func(c simple.Channel) error {
		c.Pipeline().
			AddLast(
				lengthDecod,
				lengthEncod,
				testHand,
			)
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
