package testsimple

import (
	"ag-core/contribute/agonet"
	"ag-core/contribute/agonet/simple"
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/petermattis/goid"
)

func TestSimpleTlsClientShortSyncHandler(t *testing.T) {

	addr := "localhost:8443"

	binder, err := buildCfgBinder(_agonet_tls_cli)
	if err != nil {
		t.Fatalf("buildCfgBinder failed: %v", err)
	}

	cliCfg, err := agonet.NewClientConfig(binder)
	if err != nil {
		t.Fatalf("NewClientConfig failed: %v", err)
	}

	eventhandler, err := _simpleShortSyncEventHandler2()
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

	var doSend = func(msg string) {
		ctx := context.Background()

		// replyChan := make(chan any)
		// ctx = context.WithValue(ctx, "replyChan", replyChan)

		// 创建链接
		tcon, err := client.DialContext("tcp", addr, ctx)
		if err != nil {
			t.Fatalf("Dial failed: %v", err)
		}

		// 从链接中获取channel
		channel, err := simple.ChannelFromConn(tcon)
		if err != nil {
			t.Fatalf("ChannelForConn failed: %v", err)
		}

		promise := simple.NewPromise()

		testHand := simple.NewSimpleInboundHandler(func(ctx simple.InboundContext, msg []byte) {
			// 消息正常回来时返回消息
			promise.Resolve(msg)
		})
		channel.Pipeline().AddLast(testHand)

		inactiveHand := simple.InactiveHandlerFunc(func(ctx simple.InactiveContext, ex error) {
			// 通道非激活时拒绝消息
			promise.Reject(ex)
		})
		channel.Pipeline().AddLast(inactiveHand)

		// 使用channel发送数据
		if msg == "" {
			gid := goid.Get()
			msg = fmt.Sprintf("ping_%d", gid)
		}
		channel.Write(msg)

		// reply, err := promise.AwaitTimeout(time.Millisecond * 500)
		reply, err := promise.Await()
		if err != nil {
			fmt.Printf("Await failed: %v\n", err)
		}
		//u: 等待回复
		fmt.Printf("收: %s\n", reply) // 等待回复

		channel.Close(nil)
		time.Sleep(time.Millisecond * 100)
	}
	// doSend("testerr")

	wg := sync.WaitGroup{}
	for i := 0; i < 5; i++ {
		wg.Add(1)
		// go func() {
		func() {
			// for {
			doSend("")
			// }
			defer wg.Done()
		}()
	}

	// wg.Wait()

}

func TestSimpleTlsClientShortClientSyncHandler(t *testing.T) {

	addr := "localhost:8443"

	binder, err := buildCfgBinder(_agonet_tls_cli)
	if err != nil {
		t.Fatalf("buildCfgBinder failed: %v", err)
	}

	cliCfg, err := agonet.NewClientConfig(binder)
	if err != nil {
		t.Fatalf("NewClientConfig failed: %v", err)
	}

	eventhandler, err := _simpleShortSyncEventHandler2()
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

	shortCli, err := simple.NewSimpleShortClient(
		client,
		func(opts *simple.ShortClientOptions) {
			opts.Timeout = time.Second * 2
		},
	)
	if err != nil {
		t.Fatalf("NewSimpleShortClient failed: %v", err)
	}
	var doSend = func(msg string) {
		ctx := context.Background()

		// 使用channel发送数据
		if msg == "" {
			gid := goid.Get()
			msg = fmt.Sprintf("ping_%d", gid)
		}
		fmt.Printf("发: %s\n", msg)
		reply, err := shortCli.RequestSync(ctx, addr, msg)

		if err != nil {
			fmt.Printf("Await failed: %v\n", err)
		}
		//u: 等待回复
		fmt.Printf("收: %s\n", reply) // 等待回复
	}

	wg := sync.WaitGroup{}
	for i := 0; i < 3; i++ {
		wg.Add(1)
		// go func() {
		func() {
			// for {
			doSend("")
			// }
			defer wg.Done()
		}()
	}
	wg.Wait()

	doSend("testerr")
	// shortCli.RequestSync(context.Background(), addr, "testerr")
	time.Sleep(time.Millisecond * 100)

}

func _simpleShortSyncEventHandler2() (agonet.EventHandler, error) {
	// var testHand = simple.NewSimpleInboundHandler(func(ctx simple.InboundContext, msg []byte) {
	// 	promise.Resolve(msg)
	// })

	lengthDecod := simple.NewLengthFieldDecoder(nil, 1024, 0, 2, 0, 2)
	lengthEncod := simple.NewLengthFieldEncoder(nil, 2, 0, false)

	pipelineInitializer := func(c simple.Channel) error {
		c.Pipeline().
			AddLast(
				lengthDecod,
				lengthEncod,
				// testHand,
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
