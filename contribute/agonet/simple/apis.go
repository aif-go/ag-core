package simple

import "ag-core/contribute/agonet"

type Channel interface {
	// ID get channel id
	ID() int64
	// Pipeline get pipeline
	Pipeline() Pipeline
	// EventLoop get event loop
	EventLoop() agonet.EventLoop
	// IsActive check channel is active
	IsActive() bool
	// LocalAddr local address
	LocalAddr() string
	// RemoteAddr remote address
	RemoteAddr() string

	// Trigger user event
	Trigger(event any)
	// Write write message to pipeline
	Write(any) error
	// Write1 write message to conn
	Write1([]byte) (n int, err error)
	// Close close channel
	Close(err error)
}

type (
	// Pipeline defines a message processing pipeline.
	Pipeline interface {
		// AddFirst add a handler to the first.
		AddFirst(handlers ...Handler) Pipeline

		// AddLast add a handler to the last.
		AddLast(handlers ...Handler) Pipeline

		// Channel get channel.
		Channel() Channel

		// ServeChannel serve the channel.
		ServeChannel(channel Channel)

		FireChannelActive()
		FireChannelInactive(ex error)
		FireChannelRead(message any)
		FireChannelWrite(message any)
		FireChannelException(ex error)
		FireChannelEvent(event any)
	}
)

type (
	// Handler defines an any handler
	Handler interface {
	}

	DuplexHandler interface {
		InboundHandler
		OutboundHandler
	}

	// ActiveHandler defines an active handler
	ActiveHandler interface {
		HandleActive(ctx ActiveContext)
	}

	// InactiveHandler defines an inactive handler
	InactiveHandler interface {
		HandleInactive(ctx InactiveContext, ex error)
	}

	// InboundHandler defines an Inbound handler
	InboundHandler interface {
		// HandleRead(ctx InboundContext, message Message) error
		HandleRead(ctx InboundContext, message any)
	}

	// OutboundHandler defines an Outbound handler
	OutboundHandler interface {
		// HandleWrite(ctx OutboundContext, message Message) error
		HandleWrite(ctx OutboundContext, message any)
	}

	// ExceptionHandler defines an exception handler
	ExceptionHandler interface {
		HandleException(ctx ExceptionContext, ex error)
	}

	// EventHandler defines an event handler
	EventHandler interface {
		HandleEvent(ctx EventContext, event any)
	}

	// CodecHandler defines an codec handler
	CodecHandler interface {
		Name() string
		DuplexHandler
	}

	DecoderHandler interface {
		Name() string
		InboundHandler
	}

	EncoderHandler interface {
		Name() string
		OutboundHandler
	}
)

type (
	// HandlerContext defines a base handler context
	HandlerContext interface {
		Channel() Channel
		Handler() Handler
		Write(message any)
		Trigger(event any)
	}

	// ActiveContext defines an active handler
	ActiveContext interface {
		HandlerContext
		FireActive()
	}

	// InactiveContext defines an inactive handler
	InactiveContext interface {
		HandlerContext
		FireInactive(ex error)
	}

	// InboundContext defines an inbound handler
	InboundContext interface {
		HandlerContext
		// HandlerRead(message Message)
		FireRead(message any)
	}

	// OutboundContext defines an outbound handler
	OutboundContext interface {
		HandlerContext
		// HandleWrite(message Message)
		FireWrite(message any)
	}
	// ExceptionContext defines an exception handler
	ExceptionContext interface {
		HandlerContext
		// HandleException(ex Exception)
		FireExceptionCaught(ex error)
	}

	// EventContext defines an event handler
	EventContext interface {
		HandlerContext
		FireEvent(event any)
	}
)
