package test

import (
	"github.com/aif-go/ag-core/ag/ag_conf"
	"github.com/aif-go/ag-core/ag/ag_log"
	"github.com/aif-go/ag-core/ag/ag_log/agslog"
	"github.com/aif-go/ag-core/fxs"
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	_ "net/http/pprof"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

func TestAgLevelMw(t *testing.T) {

	defer func() {
		time.Sleep(time.Second)
	}()

	os.Setenv(ag_conf.AppConfKey, "agslog_levelmw.yaml")
	slog.Error("======================1")

	logcase := map[string]*slog.Logger{}

	addlogcase := func(name string) {
		logcase[name] = agslog.GetSlogByName(name)
	}
	testlogcase := func(i int) {
		for k, v := range logcase {
			_test_log(v, fmt.Sprintf("%s_%d", k, i))
		}
	}

	toplog := agslog.GetSlog()
	logcase["toplog"] = toplog
	addlogcase("log1")
	addlogcase("log2")
	addlogcase("log3")

	testlogcase(0)

	fxapp := fx.New(

		fx.WithLogger(func() fxevent.Logger {
			return &fxevent.SlogLogger{
				Logger: toplog,
			}
		}),

		// 加载配置
		fxs.FxAgConfModule,

		ag_log.FxAglogMode,
	)

	fxapp.Start(context.Background())
	slog.Error("======================2")

	_test_log(slog.Default(), "slog default")

	testlogcase(2)

	ctx := context.Background()
	fmt.Printf("log level enable:\n")
	fmt.Printf("  debug: %v\n", toplog.Enabled(ctx, slog.LevelDebug))
	fmt.Printf("  info: %v\n", toplog.Enabled(ctx, slog.LevelInfo))
	fmt.Printf("  warn: %v\n", toplog.Enabled(ctx, slog.LevelWarn))
	fmt.Printf("  error: %v\n", toplog.Enabled(ctx, slog.LevelError))

}

// TestAgLevelMwCheck 测试log level是否生效
func TestAgLevelMwCheck(t *testing.T) {

	os.Setenv(ag_conf.AppConfKey, "agslog_levelmw.yaml")

	level1 := agslog.GetSlogByName("level1")
	level2 := agslog.GetSlogByName("level2")
	level3 := agslog.GetSlogByName("level3")
	level4 := agslog.GetSlogByName("level4")

	fxapp := fx.New(
		fxs.FxAgConfModule,
		ag_log.FxAglogMode,
	)
	ctx := context.Background()
	fxapp.Start(ctx)

	if !level1.Enabled(ctx, slog.LevelDebug) {
		t.Fail()
	}
	if !level2.Enabled(ctx, slog.LevelInfo) {
		t.Fail()
	}
	if !level3.Enabled(ctx, slog.LevelWarn) {
		t.Fail()
	}
	if !level4.Enabled(ctx, slog.LevelError) {
		t.Fail()
	}

}

func TestAgLevelMw3(t *testing.T) {
	defer func() {
		time.Sleep(time.Microsecond)
	}()

	slog.Info("==============1")

	var handler slog.Handler
	// handler = slog.Default().Handler()
	handler = slog.NewTextHandler(os.Stdout, nil)

	handler = agslog.NewReplaceableHandler("hzw", handler)

	log := slog.New(handler)

	slog.SetDefault(log)
	log.Info("==============2")

	slog.Info("==============3")

}
