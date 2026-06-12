package test

import (
	"github.com/aif-go/ag-core/ag/ag_conf"
	"github.com/aif-go/ag-core/ag/ag_log"
	"github.com/aif-go/ag-core/ag/ag_log/agslog"
	"github.com/aif-go/ag-core/fxs"
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	_ "net/http/pprof"

	"go.uber.org/fx"
)

func TestAgAsyncLog(t *testing.T) {

	defer func() {
		time.Sleep(time.Second * 10)
	}()

	os.Setenv(ag_conf.AppConfKey, "agslog_async.yaml")

	fxapp := fx.New(
		// 加载配置
		fxs.FxAgConfModule,

		/* aglog */
		// // agslog
		// agslog.FxAgSlogProvide,
		// // fanout
		// fanout.FxAgSlogFanoutProvide,
		// // async
		// async.FxAglogAsyncProvide,
		// slogzap.FxAgSlogZapProvide,

		ag_log.FxAglogMode,
	)

	fxapp.Start(context.Background())

	logger := agslog.GetSlogByName("asynclog1")
	_test_log(logger, "asynclog_0")

	// 测试并发获取logger
	startchan := make(chan struct{}, 0)
	for i := 0; i < 3; i++ {
		go func(i int) {
			<-startchan
			logger := agslog.GetSlogByName("asynclog1")
			// 打印logger的地址
			_test_log(logger, fmt.Sprintf("asynclog_%d", i))
		}(i)
	}
	close(startchan)

}
