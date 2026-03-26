package simple

import (
	"ag-core/contribute/agonet/simple/utils"
	"fmt"
)

var _ Pipeline = (*pipeline)(nil)

// NewPipeline create a pipeline.
func NewPipeline() Pipeline {

	p := &pipeline{}
	p.head = newHandlerContext(p, headHandler{}, nil, nil)
	p.tail = newHandlerContext(p, tailHandler{}, nil, nil)

	p.head.next = p.tail
	p.tail.prev = p.head

	// head + tail
	p.size = 2
	return p
}

// pipeline to implement Pipeline
type pipeline struct {
	head    *handlerContext
	tail    *handlerContext
	channel Channel
	size    int
}

func (p *pipeline) AddFirst(handlers ...Handler) Pipeline {
	// checking handler.
	checkHandler(handlers...)

	for _, h := range handlers {
		p.addFirst(h)
	}
	return p
}

func (p *pipeline) AddLast(handlers ...Handler) Pipeline {
	// checking handler.
	checkHandler(handlers...)

	for _, h := range handlers {
		p.addLast(h)
	}
	return p
}

func (p *pipeline) Channel() Channel {
	return p.channel
}

func (p *pipeline) ServeChannel(channel Channel) {
	p.channel = channel
}

func (p *pipeline) FireChannelRead(message Message) {
	p.head.FireRead(message)
}

func (p *pipeline) FireChannelWrite(message Message) {
	p.tail.FireWrite(message)
}

func (p *pipeline) FireChannelException(ex Exception) {
	p.head.FireExceptionCaught(ex)
}

// addFirst to add handlers head
func (p *pipeline) addFirst(handler Handler) {

	oldNext := p.head.next
	p.head.next = newHandlerContext(p, handler, p.head, oldNext)
	oldNext.prev = p.head.next
	p.size++
}

// addLast to add handlers tail
func (p *pipeline) addLast(handler Handler) {

	oldPrev := p.tail.prev
	p.tail.prev = newHandlerContext(p, handler, oldPrev, p.tail)
	oldPrev.next = p.tail.prev
	p.size++
}

// checkHandler to checking handlers
func checkHandler(handlers ...Handler) {

	for index, h := range handlers {
		switch h.(type) {
		case InboundHandler:
		case OutboundHandler:
		case ExceptionHandler:
		default:
			utils.Assert(fmt.Errorf("unrecognized Handler: %d:%T", index, h))
		}
	}
}
