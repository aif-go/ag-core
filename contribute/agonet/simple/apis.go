package simple

import "ag-core/contribute/agonet"

type (
	Message   interface{}
	Exception error
)

type Channel interface {
	ID() int64
	Write(Message) error
	Pipeline() Pipeline
	EventLoop() agonet.EventLoop

	Close(err error)
	IsActive() bool

	Write1([]byte) (n int, err error)

	// LocalAddr local address
	LocalAddr() string

	// RemoteAddr remote address
	RemoteAddr() string
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

		FireChannelRead(message Message)
		FireChannelWrite(message Message)
		FireChannelException(ex Exception)
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

	// InboundHandler defines an Inbound handler
	InboundHandler interface {
		// HandleRead(ctx InboundContext, message Message) error
		HandleRead(ctx InboundContext, message Message)
	}
	// OutboundHandler defines an Outbound handler
	OutboundHandler interface {
		// HandleWrite(ctx OutboundContext, message Message) error
		HandleWrite(ctx OutboundContext, message Message)
	}
	// ExceptionHandler defines an exception handler
	ExceptionHandler interface {
		HandleException(ctx ExceptionContext, ex Exception)
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
		Write(message Message)
	}

	// InboundContext defines an inbound handler
	InboundContext interface {
		HandlerContext
		// HandlerRead(message Message)
		FireRead(message Message)
	}

	// OutboundContext defines an outbound handler
	OutboundContext interface {
		HandlerContext
		// HandleWrite(message Message)
		FireWrite(message Message)
	}
	// ExceptionContext defines an exception handler
	ExceptionContext interface {
		HandlerContext
		// HandleException(ex Exception)
		FireExceptionCaught(ex Exception)
	}
)
