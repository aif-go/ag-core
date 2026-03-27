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
	// fmt.Fprintln(os.Stderr,
	// 	"An HandleException() event was fired, and it reached at the tail of the pipeline.",
	// 	"It usually means the last handler in the pipeline did not handle the exception.",
	// 	"We will close the channel, If you don't want to close the channel please add HandleException() to the pipeline.\n",
	// 	"Exception throw on ", ctx.Channel().RemoteAddr(), "\n",
	// 	ex,
	// )
	slog.Error(fmt.Sprintf("Exception throw on %s, err: %v", ctx.Channel().RemoteAddr(), ex))

	// TODO 需反馈到EventLoop，进行链接关闭
	// ctx.Channel().EventLoop()
	// err := aerrors.ErrConnectionShutdown
	// 将err传递给EventLoop 进行连接关闭
}

func (fn ActiveHandlerFunc) HandleActive(ctx ActiveContext) {
	fn(ctx)
}

func (fn InactiveHandlerFunc) HandleInactive(ctx InactiveContext, ex error) {
	fn(ctx, ex)
}
