package simple

import (
	"ag-core/contribute/agonet/simple/utils"
	"fmt"
	"log/slog"
)

type (
	headHandler struct{}

	tailHandler struct{}

	ActiveHandlerFunc func(ctx ActiveContext)

	InactiveHandlerFunc func(ctx InactiveContext, ex error)

	EventHandlerFunc func(ctx EventContext, event any)

	ExceptionHandlerFunc func(ctx ExceptionContext, ex error)
)

func (headHandler) HandleWrite(ctx OutboundContext, message any) {
	// head 处理器负责将消息写入channel
	var ch = ctx.Channel()
	switch m := message.(type) {
	case []byte:
		utils.AssertLength(ch.Write1(m))
	default:
		panic(fmt.Errorf("unsupported type: %T", m))
	}
}

func (tailHandler) HandleException(ctx ExceptionContext, ex error) {
	// The final closing operation will be provided when the user registered handler is not processing.
	slog.Error(fmt.Sprintf("An HandleException() event was fired, channel will be close. Exception throw on %s, err: %v", ctx.Channel().RemoteAddr(), ex))

	// FIXME 内部实现从eventloop中执行关闭操作
	ctx.Channel().Close(ex)
}

func (fn ActiveHandlerFunc) HandleActive(ctx ActiveContext) {
	fn(ctx)
}

func (fn InactiveHandlerFunc) HandleInactive(ctx InactiveContext, ex error) {
	fn(ctx, ex)
}

func (fn EventHandlerFunc) HandleEvent(ctx EventContext, event any) {
	fn(ctx, event)
}

func (fn ExceptionHandlerFunc) HandleException(ctx ExceptionContext, ex error) {
	fn(ctx, ex)
}
