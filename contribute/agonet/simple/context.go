package simple

var (
	_ HandlerContext   = (*handlerContext)(nil)
	_ ActiveContext    = (*handlerContext)(nil)
	_ InactiveContext  = (*handlerContext)(nil)
	_ InboundContext   = (*handlerContext)(nil)
	_ OutboundContext  = (*handlerContext)(nil)
	_ ExceptionContext = (*handlerContext)(nil)
	_ EventContext     = (*handlerContext)(nil)
)

// handlerContext impl HandlerContext
type handlerContext struct {
	pipeline Pipeline
	prev     *handlerContext
	next     *handlerContext

	handler Handler

	cast2Inbound   InboundHandler
	cast2Outbound  OutboundHandler
	cast2Exception ExceptionHandler
	cast2Active    ActiveHandler
	cast2Inactive  InactiveHandler
	cast2Event     EventHandler
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
	hc.cast2Active, _ = handler.(ActiveHandler)
	hc.cast2Inactive, _ = handler.(InactiveHandler)
	hc.cast2Event, _ = handler.(EventHandler)

	return hc
}

func (hc *handlerContext) prevContext() *handlerContext {
	return hc.prev
}

func (hc *handlerContext) nextContext() *handlerContext {
	return hc.next
}

// Channel impl HandlerContext
func (hc *handlerContext) Channel() Channel {
	return hc.pipeline.Channel()
}

// Handler impl HandlerContext
func (hc *handlerContext) Handler() Handler {
	return hc.handler
}

// Write impl HandlerContext
func (hc *handlerContext) Write(message any) {
	defer func() {
		if err := recover(); nil != err {
			hc.pipeline.FireChannelException(err.(error))
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

// FireActive impl ActiveContext
func (hc *handlerContext) FireActive() {
	var next = hc

	for {
		if next = next.nextContext(); nil == next {
			break
		}

		if handler := next.cast2Active; nil != handler {
			handler.HandleActive(next)
			break
		}
	}
}

// FireInactive impl InactiveContext
func (hc *handlerContext) FireInactive(err error) {
	var next = hc

	for {
		if next = next.nextContext(); nil == next {
			break
		}

		if handler := next.cast2Inactive; nil != handler {
			handler.HandleInactive(next, err)
			break
		}
	}
}

// FireRead impl InboundContext
func (hc *handlerContext) FireRead(message any) {
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

// FireWrite impl OutboundContext
func (hc *handlerContext) FireWrite(message any) {
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

// FireExceptionCaught impl ExceptionContext
func (hc *handlerContext) FireExceptionCaught(ex error) {
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

// FireEvent impl EventContext
func (hc *handlerContext) FireEvent(event any) {
	// TODO
}
