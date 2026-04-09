package simple

import (
	"ag-core/contribute/agonet"
	"context"
	"fmt"
	"log/slog"
	"time"
)

type IdleStateEvent struct {
	State IdleState
	First bool
}

type IdleState string

const (
	READER_IDLE IdleState = "READER_IDLE"
	WRITER_IDLE IdleState = "WRITER_IDLE"
	ALL_IDLE    IdleState = "ALL_IDLE"
)

var (
	FIRST_READ_IDLE_STATE_EVENT   = IdleStateEvent{READER_IDLE, true}
	READER_IDLE_STATE_EVENT       = IdleStateEvent{READER_IDLE, false}
	FIRST_WRITER_IDLE_STATE_EVENT = IdleStateEvent{WRITER_IDLE, true}
	WRITER_IDLE_STATE_EVENT       = IdleStateEvent{WRITER_IDLE, false}
	FIRST_ALL_IDLE_STATE_EVENT    = IdleStateEvent{ALL_IDLE, true}
	ALL_IDLE_STATE_EVENT          = IdleStateEvent{ALL_IDLE, false}
)

// IdleStateHandler fire IdleStateEvent after waiting for a idle time
func IdleStateHandler(readerIdleTime int64, writerIdleTime int64, allIdleTime int64, unit time.Duration) Handler {
	h := &idleHandler{
		firstReadIdleEvent:   true,
		firstWriterIdleEvent: true,
		firstAllIdleEvent:    true,
	}

	if readerIdleTime <= 0 {
		h.readerIdleTime = 0
	} else {
		h.readerIdleTime = max(time.Duration(readerIdleTime)*unit, time.Second)
	}
	if writerIdleTime <= 0 {
		h.writerIdleTime = 0
	} else {
		h.writerIdleTime = max(time.Duration(writerIdleTime)*unit, time.Second)
	}
	if allIdleTime <= 0 {
		h.allIdleTime = 0
	} else {
		h.allIdleTime = max(time.Duration(allIdleTime)*unit, time.Second)
	}

	return h
}

// idleHandler
type idleHandler struct {
	// mutex sync.RWMutex
	// state int

	// idleTime time.Duration

	readerIdleTime time.Duration
	writerIdleTime time.Duration
	allIdleTime    time.Duration

	firstReadIdleEvent   bool
	firstWriterIdleEvent bool
	firstAllIdleEvent    bool

	lastReadTime time.Time
	lastWritTime time.Time

	readTimer *time.Timer
	writTimer *time.Timer
	allTimer  *time.Timer

	handlerCtx HandlerContext
}

func (r *idleHandler) HandleActive(ctx ActiveContext) {
	r.initialize(ctx)

	ctx.FireActive()
}

func (r *idleHandler) HandleInactive(ctx InactiveContext, ex error) {
	r.handlerCtx = nil
	if r.readTimer != nil {
		r.readTimer.Stop()
		r.readTimer = nil
	}
	if r.writTimer != nil {
		r.writTimer.Stop()
		r.writTimer = nil
	}
	if r.allTimer != nil {
		r.allTimer.Stop()
		r.allTimer = nil
	}

	// post the inactive event.
	ctx.FireInactive(ex)
}

func (r *idleHandler) HandleRead(ctx InboundContext, message any) {
	ctx.FireRead(message)

	// update last read time.
	r.lastReadTime = time.Now()
	r.firstReadIdleEvent = true
	r.firstAllIdleEvent = true
	// reset timer.
	if r.readTimer != nil {
		r.readTimer.Reset(r.readerIdleTime)
	}
}

func (r *idleHandler) HandleWrite(ctx OutboundContext, message any) {
	ctx.FireWrite(message)

	// update last writ time.
	r.lastWritTime = time.Now()
	r.firstWriterIdleEvent = true
	r.firstAllIdleEvent = true
	// reset timer
	if r.writTimer != nil {
		r.writTimer.Reset(r.writerIdleTime)
	}
}

func (r *idleHandler) onTimeoutInEL(state IdleState) {
	if r.handlerCtx == nil {
		return
	}

	var idlefunc func()

	switch state {
	case READER_IDLE:
		if r.readTimer == nil {
			return
		}
		idlefunc = r.onReadTimeout
	case WRITER_IDLE:
		if r.writTimer == nil {
			return
		}
		idlefunc = r.onWriteTimeout
	case ALL_IDLE:
		if r.allTimer == nil {
			return
		}
		idlefunc = r.onAllTimeout
	default:
		slog.Error("invalid idle state", "state", state)
		return
	}

	runnable := agonet.RunnableFunc(func(_ context.Context) error {
		idlefunc()
		return nil
	})

	// run onRead in eventloop
	// 该方法执行不在eventloop中, 需要判断状态
	if r.handlerCtx == nil {
		return
	}

	err := r.handlerCtx.Channel().EventLoop().Execute(context.Background(), runnable)
	if nil != err {
		r.handlerCtx.Channel().Pipeline().FireChannelException(AsException(err))
	}
}

func (r *idleHandler) onReadTimeout() {
	if r.handlerCtx == nil || r.readTimer == nil {
		return
	}

	// check if the idle time expires.
	expired := time.Since(r.lastReadTime) >= r.readerIdleTime
	ctx := r.handlerCtx

	if expired && ctx != nil {
		firstReadIdleEvent := r.firstReadIdleEvent
		r.firstReadIdleEvent = false
		// trigger event.
		func() {
			// capture exception.
			defer func() {
				if err := recover(); nil != err {
					ctx.Channel().Pipeline().FireChannelException(AsException(err))
				}
			}()

			event, _ := newIdleStateEvent(READER_IDLE, firstReadIdleEvent)
			ctx.Trigger(event)
		}()
	}

	// reset timer
	if r.readTimer != nil {
		r.readTimer.Reset(r.readerIdleTime)
	}
}

func (r *idleHandler) onWriteTimeout() {
	if r.handlerCtx == nil || r.writTimer == nil {
		return
	}

	// check if the idle time expires.
	expired := time.Since(r.lastWritTime) >= r.writerIdleTime
	ctx := r.handlerCtx

	if expired && ctx != nil {
		firstWriterIdleEvent := r.firstWriterIdleEvent
		r.firstWriterIdleEvent = false
		// trigger event.
		func() {
			// capture exception.
			defer func() {
				if err := recover(); nil != err {
					ctx.Channel().Pipeline().FireChannelException(AsException(err))
				}
			}()

			event, _ := newIdleStateEvent(WRITER_IDLE, firstWriterIdleEvent)
			ctx.Trigger(event)
		}()
	}

	// reset timer
	if r.writTimer != nil {
		r.writTimer.Reset(r.writerIdleTime)
	}
}

func (r *idleHandler) onAllTimeout() {
	if r.handlerCtx == nil || r.allTimer == nil {
		return
	}

	// check if the idle time expires.
	expired := time.Since(r.lastReadTime) >= r.allIdleTime && time.Since(r.lastWritTime) >= r.allIdleTime
	ctx := r.handlerCtx

	if expired && ctx != nil {
		firstAllIdleEvent := r.firstAllIdleEvent
		r.firstAllIdleEvent = false

		// trigger event.
		func() {
			// capture exception.
			defer func() {
				if err := recover(); nil != err {
					ctx.Channel().Pipeline().FireChannelException(AsException(err))
				}
			}()

			event, _ := newIdleStateEvent(ALL_IDLE, firstAllIdleEvent)
			ctx.Trigger(event) // FIXME 从当前IdleStateHandler所在的ctx开始传播事件
		}()
	}

	// reset timer
	if r.allTimer != nil {
		r.allTimer.Reset(r.allIdleTime)
	}
}

func (r *idleHandler) initialize(ctx HandlerContext) {

	// cache context.
	r.handlerCtx = ctx

	now := time.Now()

	r.lastReadTime = now
	r.lastWritTime = now

	r.firstReadIdleEvent = true
	r.firstWriterIdleEvent = true
	r.firstAllIdleEvent = true

	if r.readerIdleTime > 0 {
		r.readTimer = time.AfterFunc(r.readerIdleTime, func() { r.onTimeoutInEL(READER_IDLE) })
	}

	if r.writerIdleTime > 0 {
		r.writTimer = time.AfterFunc(r.writerIdleTime, func() { r.onTimeoutInEL(WRITER_IDLE) })
	}

	if r.allIdleTime > 0 {
		r.allTimer = time.AfterFunc(r.allIdleTime, func() { r.onTimeoutInEL(ALL_IDLE) })
	}

}

// func (r *readIdleHandler) onReadTimeout() {
// 	var expired bool
// 	var ctx HandlerContext
// 	var firstReadIdleEvent bool

// 	r.withReadLock(func() {
// 		// check if the idle time expires.
// 		expired = time.Since(r.lastReadTime) >= r.idleTime
// 		ctx = r.handlerCtx
// 		firstReadIdleEvent = r.firstReadIdleEvent
// 		r.firstReadIdleEvent = false
// 	})

// 	if expired && ctx != nil {
// 		// trigger event.
// 		func() {
// 			// capture exception.
// 			defer func() {
// 				if err := recover(); nil != err {
// 					ctx.Channel().Pipeline().FireChannelException(AsException(err))
// 				}
// 			}()

// 			event, err := newIdleStateEvent(READER_IDLE, firstReadIdleEvent)
// 			if nil != err {
// 				ctx.Channel().Pipeline().FireChannelException(err)
// 			}
// 			ctx.Trigger(event)
// 		}()
// 	}

// 	// reset timer
// 	r.withReadLock(func() {
// 		if r.readTimer != nil {
// 			r.readTimer.Reset(r.idleTime)
// 		}
// 	})
// }

func newIdleStateEvent(state IdleState, first bool) (IdleStateEvent, error) {
	switch state {
	case READER_IDLE:
		if first {
			return FIRST_READ_IDLE_STATE_EVENT, nil
		}
		return READER_IDLE_STATE_EVENT, nil
	case WRITER_IDLE:
		if first {
			return FIRST_WRITER_IDLE_STATE_EVENT, nil
		}
		return WRITER_IDLE_STATE_EVENT, nil
	case ALL_IDLE:
		if first {
			return FIRST_ALL_IDLE_STATE_EVENT, nil
		}
		return ALL_IDLE_STATE_EVENT, nil
	default:
		return IdleStateEvent{}, fmt.Errorf("unsupported IdleState: %s", state)
	}

}
