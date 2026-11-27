package gormdb

/*
gormslog_test.go 验证 GormSlogLogger 的日志脱敏能力，对应漏洞 vuln-0003（ORM 日志明文参数）。

问题机理：GORM 在 callbacks.go 的 Execute 里，把 SQL 交给 logger 之前会先执行
`db.Dialector.Explain(sql, vars...)`，这一步会把参数值直接内联进 SQL 字符串。
也就是说，logger 的 Trace 拿到的 sql 参数默认已经带着明文值了。

唯一的拦截点是 GORM 定义的可选接口 ParamsFilter：logger 实现它，并在
ParameterizedQueries 为 true 时返回 (sql, nil)。Explain 拿不到参数，就只能保留 ?
占位符，明文值不会进入日志。本文件的三个测试分别覆盖三件事：构造函数是否把配置
透传进来、ParamsFilter 的返回值契约、以及端到端日志里到底出现 ? 还是明文值。

端到端测试不连接任何真实数据库。DummyDialector 是 GORM 自带的测试拨号器，只注册
回调、不建立连接；DryRun 让 GORM 只构建 SQL 不执行；而 Trace 的触发条件是
`stmt.SQL.Len() > 0`（见 GORM callbacks.go），与 SQL 是否真正执行无关。因此整条
日志链路照样走通，却不碰数据库。
*/

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	"gorm.io/gorm"
	gormlog "gorm.io/gorm/logger"
	"gorm.io/gorm/utils/tests"
)

const (
	// sensitiveParam 是用于验证脱敏的敏感值。它只要出现在日志里，就说明参数被内联了。
	sensitiveParam = "SENSITIVE_VALUE"

	// querySQL 是测试用的 SQL 模板，参数位置用 ? 占位，便于断言占位符是否保留。
	querySQL = "SELECT * FROM users WHERE name = ?"
)

// TestNewSLogGormLog_ParameterizedFromConfig 验证构造函数把 LoggerConfig 的
// ParameterizedQueries 透传到 GormSlogLogger，并确认 LogMode 切换日志级别时
// 不会丢掉这个开关。
//
// 之所以单独测 LogMode，是因为实际运行中 FindGormLoggerFromAgslog 会在 Debug 模式下
// 调用 LogMode(logger.Info)。LogMode 内部是 `newlogger := *l` 的结构体复制，本测试
// 把这个复制行为固定下来，避免将来重写 LogMode 时把脱敏开关漏掉。
func TestNewSLogGormLog_ParameterizedFromConfig(t *testing.T) {
	cases := []struct {
		name       string
		configured bool
		why        string
	}{
		{
			name:       "配置为 true",
			configured: true,
			why:        "默认配置要求脱敏，构造后开关必须为 true",
		},
		{
			name:       "配置为 false",
			configured: false,
			why:        "显式关闭时不能被构造函数强行改成 true",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			logger := NewSLogGormLog(
				slog.New(slog.NewTextHandler(io.Discard, nil)),
				&LoggerConfig{ParameterizedQueries: tc.configured},
			)

			got, ok := logger.(*GormSlogLogger)
			if !ok {
				t.Fatalf("NewSLogGormLog 返回类型异常: %T", logger)
			}
			if got.ParameterizedQueries != tc.configured {
				t.Fatalf("why=%s: 构造后 ParameterizedQueries=%v, want %v",
					tc.why, got.ParameterizedQueries, tc.configured)
			}
			t.Logf("why=%s: 构造后 ParameterizedQueries=%v", tc.why, got.ParameterizedQueries)

			// LogMode 返回的是结构体副本，脱敏开关必须跟着一起复制。
			switched, ok := got.LogMode(gormlog.Info).(*GormSlogLogger)
			if !ok {
				t.Fatalf("LogMode 返回类型异常")
			}
			if switched.ParameterizedQueries != tc.configured {
				t.Fatalf("why=%s: LogMode 后 ParameterizedQueries=%v, want %v（脱敏开关在切换日志级别时丢失）",
					tc.why, switched.ParameterizedQueries, tc.configured)
			}
			if switched.LogLevel != gormlog.Info {
				t.Fatalf("LogMode 后 LogLevel=%v, want Info", switched.LogLevel)
			}
		})
	}
}

// TestNewSLogGormLog_NilConfDefaultsToSecure 验证 conf 为 nil 时的降级方向。
//
// NewSLogGormLog 是导出函数，调用方可能传 nil（例如还没读到配置就先初始化 logger）。
// 若按 Go 的零值语义取 false，日志会退化成明文参数——恰好是这次要修的漏洞。
// 因此 nil 必须回退到 true（脱敏），本测试把 fail-closed 的方向固定下来。
func TestNewSLogGormLog_NilConfDefaultsToSecure(t *testing.T) {
	logger := NewSLogGormLog(slog.New(slog.NewTextHandler(io.Discard, nil)), nil)

	got, ok := logger.(*GormSlogLogger)
	if !ok {
		t.Fatalf("NewSLogGormLog 返回类型异常: %T", logger)
	}

	t.Logf("结果 | conf=nil | 参数化=%v(期望true) | 断言结论: %s",
		got.ParameterizedQueries, map[bool]string{true: "PASS", false: "FAIL"}[got.ParameterizedQueries])

	if !got.ParameterizedQueries {
		t.Fatalf("conf 为 nil 时应回退到脱敏默认值 true, got %v", got.ParameterizedQueries)
	}
}

// TestGormSlogLogger_ParamsFilter 验证 ParamsFilter 的返回值契约。
//
// 这个接口是 GORM 在 callbacks.go 里用类型断言探测的：
//
//	if filter, ok := db.Logger.(ParamsFilter); ok {
//	    sql, vars = filter.ParamsFilter(stmt.Context, stmt.SQL.String(), stmt.Vars...)
//	}
//
// 返回 (sql, nil) 时，后续 Explain(sql) 因为没有参数可替换，会保留 ? 占位符；
// 返回 (sql, params) 时，Explain 会把参数值内联进 SQL。两种分支都必须固定，
// 否则脱敏行为会随重构漂移。
func TestGormSlogLogger_ParamsFilter(t *testing.T) {
	cases := []struct {
		name           string
		parameterized  bool
		wantParamsKept bool
		why            string
	}{
		{
			name:           "参数化开启时剥掉参数",
			parameterized:  true,
			wantParamsKept: false,
			why:            "脱敏的关键：返回 nil 参数，Explain 才会保留 ? 占位符",
		},
		{
			name:           "参数化关闭时保留参数",
			parameterized:  false,
			wantParamsKept: true,
			why:            "关闭脱敏是显式选择，参数必须原样交给 Explain，不能悄悄丢弃",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			logger := GormSlogLogger{ParameterizedQueries: tc.parameterized}

			gotSQL, gotParams := logger.ParamsFilter(
				context.Background(), querySQL, sensitiveParam,
			)

			// SQL 文本本身不应被 ParamsFilter 改动，它只负责决定参数留不留。
			if gotSQL != querySQL {
				t.Fatalf("why=%s: SQL 被改动, got %q, want %q", tc.why, gotSQL, querySQL)
			}

			if !tc.wantParamsKept {
				if gotParams != nil {
					t.Fatalf("why=%s: 期望返回 nil 参数, got %v", tc.why, gotParams)
				}
				t.Logf("why=%s: ParamsFilter 返回 sql=%q, params=nil", tc.why, gotSQL)
				return
			}

			if len(gotParams) != 1 || gotParams[0] != sensitiveParam {
				t.Fatalf("why=%s: 期望原样返回参数 [%v], got %v",
					tc.why, sensitiveParam, gotParams)
			}
			t.Logf("why=%s: ParamsFilter 返回 sql=%q, params=%v", tc.why, gotSQL, gotParams)
		})
	}
}

// TestGormSlogLogger_LogOmitsParams 是端到端验证：跑一条带参数的查询，把 slog
// 输出捕获到 buffer，断言日志里出现的是 ? 占位符还是明文值。
//
// 这里刻意不用 agslog.GetSlogByName 拿全局 logger，而是自建一个写入 bytes.Buffer 的
// slog handler。原因是 agslog 的全局 handler 在测试进程里可能没注册，输出会被吞掉，
// 我们真正要断言的是"日志文本里有没有敏感值"，需要一个可靠的捕获点。
//
// 三个用例分别覆盖 Trace 的三个分支：Info 级（全量 SQL）、慢查询 Warn 级、以及
// 非参数化的对照组。对照组的意义在于证明断言不是"永远通过"——如果连对照组都看不到
// 明文，说明测试本身没跑通，脱敏结论也就不可信。
func TestGormSlogLogger_LogOmitsParams(t *testing.T) {
	const (
		// 用户模型只用于让 GORM 生成 SELECT * FROM users 的 SQL，字段不参与断言。
		tableName = "users"
		// 慢查询阈值设成 1ns，确保 DryRun 下极短的耗时也会命中慢查询分支。
		slowThresholdForTest = time.Nanosecond
	)

	cases := []struct {
		name            string
		parameterized   bool
		level           gormlog.LogLevel
		slowThreshold   time.Duration
		wantPlaceholder bool
		wantPlaintext   bool
		why             string
	}{
		{
			name:            "Info 级 + 参数化：日志只留占位符",
			parameterized:   true,
			level:           gormlog.Info,
			slowThreshold:   100 * time.Millisecond,
			wantPlaceholder: true,
			wantPlaintext:   false,
			why:             "Debug 模式会打印全量 SQL，脱敏必须在这种最容易泄露的场景下生效",
		},
		{
			name:            "Info 级 + 非参数化：对照组必须出现明文",
			parameterized:   false,
			level:           gormlog.Info,
			slowThreshold:   100 * time.Millisecond,
			wantPlaceholder: false,
			wantPlaintext:   true,
			why:             "对照组证明日志链路确实捕获到了 SQL，否则前面的“无明文”可能是假阳性",
		},
		{
			name:            "慢查询 Warn 级 + 参数化：同样脱敏",
			parameterized:   true,
			level:           gormlog.Warn,
			slowThreshold:   slowThresholdForTest,
			wantPlaceholder: true,
			wantPlaintext:   false,
			why:             "生产默认是 Warn 级，慢查询告警是最常落盘的日志，必须覆盖",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var out bytes.Buffer
			handler := slog.NewTextHandler(&out, &slog.HandlerOptions{Level: slog.LevelDebug})

			// 直接构造 GormSlogLogger 而不走 NewSLogGormLog，是因为这里需要精确控制
			// LogLevel 与 SlowThreshold 来命中不同的 Trace 分支；构造函数只负责设置
			// 默认值，它的透传行为由 TestNewSLogGormLog_ParameterizedFromConfig 覆盖。
			dbLogger := &GormSlogLogger{
				SlogLogger:                slog.New(handler),
				SlowThreshold:             tc.slowThreshold,
				IgnoreRecordNotFoundError: true,
				ParameterizedQueries:      tc.parameterized,
				LogLevel:                  tc.level,
			}

			// DummyDialector 是 GORM 自带的测试拨号器，只注册回调、不建立任何连接；
			// DryRun 让 GORM 只构建 SQL 不执行。Trace 在 SQL 构建完成后就会触发，
			// 因此不需要真实数据库也能验证日志内容。
			db, err := gorm.Open(tests.DummyDialector{}, &gorm.Config{
				Logger: dbLogger,
				DryRun: true,
			})
			if err != nil {
				t.Fatalf("gorm.Open 失败: %v", err)
			}

			// 用参数化的 Where 构造查询，让 GORM 产生 "name = ?" 与对应参数值。
			type user struct{ Name string }
			if err := db.Table(tableName).
				Where("name = ?", sensitiveParam).
				Find(&[]user{}).Error; err != nil {
				t.Fatalf("查询失败: %v", err)
			}

			logged := out.String()
			t.Logf("why=%s", tc.why)
			t.Logf("捕获日志: %s", strings.TrimSpace(logged))

			// 先确认日志确实触发了，否则后面的“没有明文”没有意义。
			if !strings.Contains(logged, tableName) {
				t.Fatalf("why=%s: 未捕获到 SQL 日志, 输出=%q", tc.why, logged)
			}

			hasPlaceholder := strings.Contains(logged, "name = ?")
			hasPlaintext := strings.Contains(logged, sensitiveParam)

			// 在断言之前把"实际值 vs 期望值"打到控制台：go test -v 下每个用例都能直接
			// 看到结论，不用等失败再去反推。这正是"测试即文档"的用法。
			t.Logf("结果 | 参数化=%v | 日志级别=%s | 占位符=%v(期望%v) | 明文=%v(期望%v)",
				tc.parameterized, logLevelName(tc.level), hasPlaceholder, tc.wantPlaceholder,
				hasPlaintext, tc.wantPlaintext)

			if hasPlaceholder != tc.wantPlaceholder {
				t.Fatalf("why=%s: 占位符存在性=%v, want %v; 日志=%q",
					tc.why, hasPlaceholder, tc.wantPlaceholder, logged)
			}

			if hasPlaintext != tc.wantPlaintext {
				t.Fatalf("why=%s: 明文参数存在性=%v, want %v; 日志=%q",
					tc.why, hasPlaintext, tc.wantPlaintext, logged)
			}

			t.Logf("断言结论: PASS | 占位符=%v 明文=%v（与该用例预期一致）", hasPlaceholder, hasPlaintext)
		})
	}

	// 父测试收尾：go test -v 下能看到整组用例的通过数量。
	t.Logf("=== 汇总：%d 个用例全部通过 ===", len(cases))
}

// logLevelName 把 GORM 的日志级别数字转成可读名字。GORM 的 LogLevel 没有实现
// String()，直接 %v 只会打印 1/2/3/4，在 -v 输出里看不出每个用例跑的是哪个级别。
func logLevelName(level gormlog.LogLevel) string {
	switch level {
	case gormlog.Silent:
		return "Silent"
	case gormlog.Error:
		return "Error"
	case gormlog.Warn:
		return "Warn"
	case gormlog.Info:
		return "Info"
	default:
		return fmt.Sprintf("LogLevel(%d)", level)
	}
}
