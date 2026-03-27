package simple

import (
	"ag-core/contribute/agonet"
	"ag-core/contribute/agonet/pkg/aerrors"
	"context"
	"errors"
	"log/slog"
)

type context_channel_key struct{}

type SimpleEventHandler struct {
	agonet.BuiltinEventEngine
	*eventHandlerOptions
}

func NewSimpleEventHandlerWithOptions(option ...Option) (agonet.EventHandler, error) {
	options := &eventHandlerOptions{
		channelIDFactory: SequenceID(),
		pipelineFactory:  NewPipeline,
		channelFactory:   newChannel,
	}
	for _, opt := range option {
		opt(options)
	}

	err := options.check()
	if err != nil {
		return nil, err
	}

	h := &SimpleEventHandler{
		eventHandlerOptions: options,
	}

	return h, nil

}

// OnBoot
func (h *SimpleEventHandler) OnBoot(eng agonet.Engine) (action agonet.Action) {
	eng.IsClient()
	// FIXME 检查相关配置
	return
}

// OnClose
func (h *SimpleEventHandler) OnClose(conn agonet.Conn, err error) (action agonet.Action) {
	// 从context中获取pipeline
	channel, err2 := getChannel(conn)
	if err2 != nil {
		return agonet.None
	}

	// TODO wait async send finished

	// 从通道中获取pipeline
	channel.Pipeline().FireChannelInactive(err)

	channel.Close(err)

	return
}

// OnOpen
func (h *SimpleEventHandler) OnOpen(conn agonet.Conn) (out []byte, action agonet.Action) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("OnOpen failed", "err", r)
			action = agonet.Close
		}
	}()

	// 创建conn context
	cctx := conn.Context()
	var ctx context.Context
	if cctx != nil {
		var ok bool
		ctx, ok = cctx.(context.Context)
		if !ok {
			return nil, agonet.Close // conn 存在ctx,且不是context
		}
	}
	if ctx == nil {
		ctx = context.Background()
	}

	// 创建pipeline
	pipeline := h.pipelineFactory()

	// 创建channel
	channel := h.channelFactory(conn, pipeline)

	// 初始化pipeline
	err := h.channelInitializer(channel)
	if err != nil {
		// 初始化pipeline失败，关闭连接
		channel.Close(err)
		action = agonet.Close
	}

	// 为pipeline绑定channel
	channel.Pipeline().ServeChannel(channel)

	// 将channel绑定到conn context中
	ctx = context.WithValue(ctx, context_channel_key{}, channel)

	// 触发通道激活事件
	channel.Pipeline().FireChannelActive()

	// 设置conn context
	conn.SetContext(ctx)
	return
}

func (h *SimpleEventHandler) OnTraffic(conn agonet.Conn) (action agonet.Action) {
	defer func() {
		if r := recover(); r != nil {
			err, ok := r.(error)
			if ok && errors.Is(err, aerrors.ErrIncompletePacket) {
				// 数据不完整，等待更多数据
				action = agonet.None
				return
			}
			slog.Error("OnTraffic failed", "err", err)
			action = agonet.Close
		}
	}()

	// channel := conn.Context().(Channel) // TODO 从context中获取pipeline
	channel, err := getChannel(conn)
	if err != nil {
		return agonet.Close
	}

	// 从连接中获取读取器
	reader := conn.(agonet.Reader)

	// 从通道中获取pipeline
	pipeline := channel.Pipeline()

	for conn.ReadableBytes() > 0 { // TODO 此处理应该提交到eventLoop中处理
		// 触发通道读取事件
		pipeline.FireChannelRead(reader)
	}

	// TODO 异常处理 -> 断连接 还是 shutdown

	return
}

// func (h *SimpleEventHandler) channelInitializer(channel Channel) error {
// 	pipeline := channel.Pipeline()

// 	pipeline.AddLast(h.handlers...)
// 	return nil
// }

// 从conn中获取channel
func getChannel(conn agonet.Conn) (Channel, error) {
	cctx := conn.Context()
	if cctx == nil {
		return nil, errors.New("conn.Context is nil")
	}

	if ctx, ok := cctx.(context.Context); ok {
		channel, ok := ctx.Value(context_channel_key{}).(Channel)
		if !ok {
			return nil, errors.New("conn.Context is not context.Context")
		}
		return channel, nil
	} else {
		return nil, errors.New("conn.Context is not context.Context")
	}
}
