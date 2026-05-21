package test

import (
	"ag-core/ag/ag_conf"
	"ag-core/ag/ag_log/agslog"
	"ag-core/ag/ag_log/async"
	"ag-core/ag/ag_log/fanout"
	"ag-core/ag/ag_log/slogzap"
	"ag-core/fxs"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"testing"
	"time"

	_ "net/http/pprof"

	"go.uber.org/fx"
)

var (
	rlog *slog.Logger
)

// func init() {
// 	rlog = agslog.GetSlogByName("zap1")
// 	_test_log(rlog, "rlog000")
// }

func TestAgSlog(t *testing.T) {

	defer func() {
		time.Sleep(time.Millisecond * 500)
	}()
	// 启动 http pprof
	go func() {
		http.ListenAndServe(":6060", nil)
	}()

	os.Setenv(ag_conf.AppConfKey, "agslog.yaml")

	_test_log(rlog, "rlog111")

	fxapp := fx.New(
		// 加载配置
		fxs.FxAgConfModule,

		/* aglog */
		// agslog
		agslog.FxAgSlogProvide,
		// fanout
		fanout.FxAgSlogFanoutProvide,
		// async
		async.FxAglogAsyncProvide,

		// ag_log.FxAglogMode,

		// slogzap
		slogzap.FxAgSlogZapProvide,

		fx.Invoke(func(logger *slog.Logger) {
			log := logger
			_test_log(log, "test1")
			_test_log(rlog, "rlog111_222")
		}),
	)

	fxapp.Start(context.Background())

	log := agslog.TopLogger()
	log2 := log.WithGroup("hzwg1")

	_test_log(rlog, "rlog222")

	_test_log(log2, "000")

	// logger := agslog.GetSlogByName("zap1")
	logger := agslog.GetSlogByName("f1")
	_test_log(logger, "test1")

	logger = agslog.GetSlogByName("f3")
	_test_log(logger, "test3")

	// 测试并发获取logger
	startchan := make(chan struct{}, 0)
	for i := 0; i < 3; i++ {
		go func(i int) {
			<-startchan
			logger := agslog.GetSlogByName("gftest")
			// 打印logger的地址
			_test_log(logger, fmt.Sprintf("gftest_log:%p_%d", logger, i))
		}(i)
	}
	close(startchan)

	_test_log(rlog, "rlog333")

}

// go test -bench=BenchmarkAgSlogInfoLevel
// go test -bench='^BenchmarkAgSlogInfoLevel$' -run='^$' -benchmem  -v
func BenchmarkAgSlogInfoLevel(b *testing.B) {
	defer func() {
		time.Sleep(time.Millisecond * 500)
	}()
	// 启动 http pprof
	go func() {
		http.ListenAndServe(":6060", nil)
	}()

	os.Setenv(ag_conf.AppConfKey, "agslog.yaml")

	rlog := agslog.GetSlogByName("zap1")
	_test_log(rlog, "rlog1")

	fxapp := fx.New(
		// 加载配置
		fxs.FxAgConfModule,

		// 提供slog.Logger
		agslog.FxAgSlogProvide,

		slogzap.FxAgSlogZapProvide,

		fanout.FxAgSlogFanoutProvide,

		fx.Invoke(func(logger *slog.Logger) {
			log := logger
			_test_log(log, "test1")
		}),
	)

	fxapp.Start(context.Background())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// _test_log(rlog, fmt.Sprintf("benlog_%d", i))
		rlog.Info(fmt.Sprintf("benlog_%d", i))
	}

}
