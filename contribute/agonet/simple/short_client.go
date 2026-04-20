package simple

import (
	"ag-core/contribute/agonet"
	"context"
	"time"
)

var (
	_ SimpleShortClient = (*simpleShortClient)(nil)
)

type ResponseCallback func(response any, err error)

type SimpleShortClient interface {
	// 同步调用：建立连接→发送请求→接收响应→关闭连接
	RequestSync(ctx context.Context, addr string, request any) (any, error)

	// 异步调用：建立连接→发送请求→通过回调处理响应
	RequestAsync(ctx context.Context, addr string, request any, callback ResponseCallback) error

	// 关闭客户端，释放资源
	Close() error
}

type simpleShortClient struct {
	// 底层 agonet 客户端
	client agonet.Client

	opts *ShortClientOptions

	// // 同步信号量控制并发
	// semaphore chan struct{}
}

type ShortClientOption func(*ShortClientOptions)

type ShortClientOptions struct {
	Timeout time.Duration
}

func NewSimpleShortClient(client agonet.Client, opt ...ShortClientOption) (SimpleShortClient, error) {
	opts := &ShortClientOptions{
		Timeout: time.Second * 30, // 默认超时时间30秒
	}

	for _, o := range opt {
		o(opts)
	}

	cli := &simpleShortClient{
		client: client,
		opts:   opts,
	}
	return cli, nil
}

func (c *simpleShortClient) RequestSync(ctx context.Context, addr string, request any) (any, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	// 创建链接
	tcon, err := c.client.DialContext("tcp", addr, ctx)
	if err != nil {
		return nil, err
	}
	defer tcon.Close()

	// 从链接中获取channel
	channel, err := ChannelFromConn(tcon)
	if err != nil {
		return nil, err
	}

	promise := NewPromise()

	promiseHandler := NewSimpleInboundHandler(func(ctx InboundContext, msg []byte) {
		// 消息正常回来时返回消息
		promise.Resolve(msg)
	})
	channel.Pipeline().AddLast(promiseHandler)

	inactiveHand := InactiveHandlerFunc(func(ctx InactiveContext, ex error) {
		// 通道非激活时拒绝消息
		promise.Reject(ex)
	})
	channel.Pipeline().AddLast(inactiveHand)

	err = channel.Write(request)
	if err != nil {
		return nil, err
	}

	// reply, err := promise.AwaitTimeout(time.Millisecond * 500)
	timeout := c.opts.Timeout // 默认超时时间30秒
	if timeout <= 0 {
		timeout = time.Second * 30
	}
	// reply, err := promise.Await()
	reply, err := promise.AwaitTimeout(timeout)
	if err != nil {
		// fmt.Printf("Await failed: %v\n", err)
		return nil, err
	}

	return reply, nil
}

func (c *simpleShortClient) RequestAsync(ctx context.Context, addr string, request any, callback ResponseCallback) error {
	go func() {
		response, err := c.RequestSync(ctx, addr, request)
		callback(response, err)
	}()
	return nil
}

func (c *simpleShortClient) Close() error {
	return c.client.Stop()
}
