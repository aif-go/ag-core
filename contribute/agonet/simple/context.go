package simple

var (
	_ HandlerContext = (*handlerContext)(nil)
	_ InboundContext = (*handlerContext)(nil)
)

// handlerContext impl HandlerContext
type handlerContext struct {
	pipeline Pipeline
	prev     *handlerContext
	next     *handlerContext

	handler        Handler
	cast2Inbound   InboundHandler
	cast2Outbound  OutboundHandler
	cast2Exception ExceptionHandler
}

func newHandlerContext(p Pipeline, handler Handler, prev, next *handlerContext) *handlerContext {
	hc := &handlerContext{
		pipeline: p,
		handler:  handler,
		prev:     prev,
		next:     next,
	}

	hc.cast2Inbound, _ = handler.(InboundHandler)
	hc.cast2Outbound, _ = handler.(OutboundHandler)
	hc.cast2Exception, _ = handler.(ExceptionHandler)
	return hc
}

func (hc *handlerContext) prevContext() *handlerContext {
	return hc.prev
}

func (hc *handlerContext) nextContext() *handlerContext {
	return hc.next
}

func (hc *handlerContext) Channel() Channel {
	return hc.pipeline.Channel()
}

func (hc *handlerContext) Handler() Handler {
	return hc.handler
}

func (hc *handlerContext) Write(message Message) {
	defer func() {
		if err := recover(); nil != err {
			hc.pipeline.FireChannelException(err.(Exception))
		}
	}()

	var next = hc

	for {
		if next = next.prevContext(); nil == next {
			break
		}

		if handler := next.cast2Outbound; nil != handler {
			handler.HandleWrite(next, message)
			break
		}
	}
}

func (hc *handlerContext) FireRead(message Message) {
	var next = hc

	for {
		if next = next.nextContext(); nil == next {
			break
		}

		if handler := next.cast2Inbound; nil != handler {
			handler.HandleRead(next, message)
			break
		}
	}
}

// FireWrite OutboundContext
func (hc *handlerContext) FireWrite(message Message) {
	var prev = hc

	for {
		if prev = prev.prevContext(); nil == prev {
			break
		}

		if handler := prev.cast2Outbound; nil != handler {
			handler.HandleWrite(prev, message)
			break
		}
	}
}

func (hc *handlerContext) FireExceptionCaught(ex Exception) {
	var next = hc

	for {
		if next = next.nextContext(); nil == next {
			break
		}

		if handler := next.cast2Exception; nil != handler {
			handler.HandleException(next, ex)
			break
		}
	}
}
