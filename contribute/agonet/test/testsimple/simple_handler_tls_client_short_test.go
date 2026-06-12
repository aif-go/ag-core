package testsimple

import (
	"github.com/aif-go/ag-core/contribute/agonet"
	"github.com/aif-go/ag-core/contribute/agonet/simple"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/petermattis/goid"
)

func TestSimpleTlsClientShortHandler(t *testing.T) {

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

	eventhandler, err := _simpleShortEventHandler()
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

	var do = func() {
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

		// 使用channel发送数据
		gid := goid.Get()
		channel.Write(fmt.Sprintf("ping_%d", gid))
		time.Sleep(time.Millisecond)

		channel.Close(nil)

		channel.IsActive()
	}

	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			// func() {
			// for {
			do()
			// }
			defer wg.Done()
		}()
	}

	wg.Wait()

}

func _simpleShortEventHandler() (agonet.EventHandler, error) {
	var testHand = simple.NewSimpleInboundHandler(func(ctx simple.InboundContext, msg []byte) {
		fmt.Printf("收: %s\n", string(msg))
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
