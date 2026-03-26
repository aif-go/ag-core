package simple

import (
	"ag-core/contribute/agonet"
	"ag-core/contribute/agonet/pkg/aerrors"
	"errors"
	"log/slog"
)

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

func (h *SimpleEventHandler) OnBoot(eng agonet.Engine) (action agonet.Action) {
	eng.IsClient()
	// FIXME 检查相关配置
	return
}

func (h *SimpleEventHandler) OnClose(conn agonet.Conn, err error) (action agonet.Action) {
	channel := conn.Context().(Channel)
	channel.Close(err)
	// TODO 连接关闭时触发
	return
}

func (h *SimpleEventHandler) OnOpen(conn agonet.Conn) (out []byte, action agonet.Action) {
	// 创建pipeline
	pipeline := h.pipelineFactory()

	// 创建通道
	// channel := newChannel(conn, pipeline)
	channel := h.channelFactory(conn, pipeline)

	// ChannelInitializer 初始化pipeline
	err := h.channelInitializer(channel)
	if err != nil {
		// 初始化pipeline失败，关闭连接
		channel.Close(err)
		action = agonet.Close
	}

	channel.Pipeline().ServeChannel(channel)

	conn.SetContext(channel)
	return
}

func (h *SimpleEventHandler) OnTraffic(conn agonet.Conn) (action agonet.Action) {
	defer func() {
		if r := recover(); r != nil {
			err, ok := r.(error)
			// if ok && err == aerrors.ErrIncompletePacket {
			if ok && errors.Is(err, aerrors.ErrIncompletePacket) {
				// 数据不完整，等待更多数据
				action = agonet.None
				return
			}
			slog.Error("OnTraffic failed", "err", err)
			action = agonet.Close
		}
	}()

	channel := conn.Context().(Channel) // TODO 从context中获取pipeline

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
