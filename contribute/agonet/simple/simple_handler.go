package simple

func NewSimpleInboundHandler[T any](onMsg func(ctx InboundContext, message T)) Handler {
	return &SimpleInboundHandler[T]{
		OnMsg: onMsg,
	}
}

type SimpleInboundHandler[T any] struct {
	OnMsg func(ctx InboundContext, message T)
}

func (e *SimpleInboundHandler[T]) HandleRead(ctx InboundContext, message Message) {
	if e.OnMsg == nil {
		ctx.FireRead(message)
		return
	}

	t, ok := message.(T)
	if ok {
		func() {
			// TODO 处理panic的情况
			if r := recover(); r != nil {
				ctx.Channel().Pipeline().FireChannelException(AsException(r))
				// TODO 此处异常怎么处理
			}

			e.OnMsg(ctx, t)
		}()
	} else {
		ctx.FireRead(message)
	}
}
