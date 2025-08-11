package test

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"testing"
	"time"

	slogmulti "github.com/samber/slog-multi"
	slogzap "github.com/samber/slog-zap/v2"
	"go.uber.org/zap"
)

func TestSlogMultiExample1(t *testing.T) {
	logstash1, _ := net.Dial("tcp", "localhost:1000")

	logger := slog.New(
		slogmulti.Failover()(
			slog.NewJSONHandler(logstash1, &slog.HandlerOptions{}), // Failover 此失败
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}), // Failover 此成功
		),
	)

	logger.
		With(
			slog.Group("user",
				slog.String("id", "user-123"),
				slog.Time("created_at", time.Now().AddDate(0, 0, -1)),
			),
		).
		With("environment", "dev").
		With("error", fmt.Errorf("an error")).
		Error("A message")

}

func TestSlogMultiSlogZap1(t *testing.T) {
	zapLogger, _ := zap.NewProduction()

	// 从context中获取tradeid
	afc := func(ctx context.Context) []slog.Attr {
		tradeid := ctx.Value("tradeid")
		if tradeid != nil {
			return []slog.Attr{
				slog.String("tradeid", tradeid.(string)),
			}
		}
		return []slog.Attr{}
	}

	opt := slogzap.Option{
		Level:           slog.LevelDebug,
		Logger:          zapLogger,
		AddSource:       true,
		AttrFromContext: []func(ctx context.Context) []slog.Attr{afc},
	}
	fmt.Sprint(opt)

	// logger := slog.New(opt.NewZapHandler())
	logger := slog.New(
		// slogmulti.Failover()( // Failover 会依次尝试每个handler，直到成功为止. 不会执行所有handler
		slogmulti.Fanout( // Fanout 会同时将记录发送给所有handler
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}),
			opt.NewZapHandler(),
		),
	)

	// context中设置tradeid
	ctx := context.WithValue(context.Background(), "tradeid", "trade-123")
	_loginfo(logger)
	_loginfoCtx(ctx, logger)
}

func TestSlogZap(t *testing.T) {
	zaplog, _ := zap.NewProduction()
	// 默认
	lv := zaplog.Level()
	fmt.Println(lv)

	logger := slog.New(slogzap.Option{Logger: zaplog}.NewZapHandler())

	_loginfo(logger)
}

func TestSlogZap2(t *testing.T) {
	logger := slog.New(slogzap.Option{}.NewZapHandler())

	_loginfo(logger)
}

func _loginfoCtx(ctx context.Context, logger *slog.Logger) {
	logger.InfoContext(ctx, "loginfo")
	logger.DebugContext(ctx, "logdebug")
	logger.WarnContext(ctx, "logwarn")
	logger.ErrorContext(ctx, "logerror")
}

func _loginfo(logger *slog.Logger) {
	logger.Info("loginfo")
	logger.Debug("logdebug")
	logger.Warn("logwarn")
	logger.Error("logerror")
}

func mockErr() error {
	return fmt.Errorf("test error")
}
