package agslog

import (
	"io"
	"log/slog"
	"testing"
)

func TestAgSlogSimpleDemo1(t *testing.T) {
	builder := NewBuilder()
	// builder.RegTopHandler(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	builder.RegTopHandler(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelInfo}))
	builder.RegTopHandler(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug}))
	builder.RegTopHandler(NewNamedHandler("test1", slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug})))

	logger, _ := builder.Build()

	logger.Info("test info")
	logger.Info("test info")
	logger.Info("test info")
	logger.Info("test info")
	logger.Debug("test debug")
}

// func TestAgSlogSimpleDemo2(t *testing.T) {
// 	// Named Handler
// 	RegHandler(NewNamedHandler("test1", slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug})))
// 	RegHandler(NewNamedHandler("test2", slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug})))
// 	// Named 套娃
// 	RegHandler(NewNamedHandler("test3", NewNamedHandler("test3.1", NewNamedHandler("test3.2", slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug})))))
// 	// {"time":"2025-08-07T11:16:30.23080427+08:00","level":"DEBUG","msg":"test debug","handler_name":"test3","handler_name":"test3.1","handler_name":"test3.2"}
// 	// 套娃时打印的handler_name是test3,test3.1,test3.2 且 key是一样的，是否有问题 TODO

// 	// Top Handler
// 	log1 := slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug})
// 	RegTopHandler(log1)

// 	prop := &AgSlogProperties{
// 		TopHandler: []string{"test1", "test2", "test3"}, // 设置Named Top Handler
// 	}

// 	opt := BuildAgSlogOption(prop)

// 	logger := BuildAgSlog(opt)

// 	logger.Debug("test debug")
// }

// BenchmarkSingleHandlerUnmatchLevel-8           4593769	       259.6 ns/op	     264 B/op	       7 allocs/op
func BenchmarkSingleHandlerUnmatchLevel(b *testing.B) {
	builder := NewBuilder()
	builder.RegTopHandler(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelError}))
	logger, _ := builder.Build()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark message")
	}
}

// BenchmarkSingleHandlerUnmatchLevel_Slog-8   	 6064204	       243.0 ns/op	     264 B/op	       7 allocs/op
func BenchmarkSingleHandlerUnmatchLevel_Slog(b *testing.B) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelError}))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark message")
	}
}

// BenchmarkSingleHandlerMatchLevel-8   	     1621087	       730.8 ns/op	     264 B/op	       7 allocs/op
func BenchmarkSingleHandlerMatchLevel(b *testing.B) {
	builder := NewBuilder()
	builder.RegTopHandler(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelInfo}))
	logger, _ := builder.Build()

	// opt := &AgSlogOption{}
	// // 不输出到stdout
	// opt.TopHandlers = append(opt.TopHandlers, slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelInfo}))
	// logger := BuildAgSlog(opt)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark message")
	}
}

// BenchmarkSingleHandlerMatchLevel_Slog-8   	 1694587	       710.9 ns/op	     264 B/op	       7 allocs/op
func BenchmarkSingleHandlerMatchLevel_Slog(b *testing.B) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelInfo}))
	b.Run("slogjson", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			logger.Info("info message")
		}
	})
}

// // BenchmarkMultipleHandlers-8   	1000000	      1049 ns/op	     264 B/op	       7 allocs/op
// func BenchmarkMultipleHandlers(b *testing.B) {
// 	opt := &AgSlogOption{}
// 	// match level
// 	opt.TopHandlers = append(opt.TopHandlers, slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelInfo}))
// 	opt.TopHandlers = append(opt.TopHandlers, slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelInfo}))
// 	logger := BuildAgSlog(opt)

// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		logger.Info("benchmark message")
// 	}
// }

// // BenchmarkMultipleHandlers2-8 1621003	       882.2 ns/op	     264 B/op	       7 allocs/op
// func BenchmarkMultipleHandlers_1match_1unmatch(b *testing.B) {
// 	opt := &AgSlogOption{}
// 	// match level
// 	opt.TopHandlers = append(opt.TopHandlers, slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelInfo}))
// 	// unmatch level
// 	opt.TopHandlers = append(opt.TopHandlers, slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelError}))
// 	logger := BuildAgSlog(opt)

// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		logger.Info("benchmark message")
// 	}
// }

// func BenchmarkNestedHandlers(b *testing.B) {
// 	opt := &AgSlogOption{}
// 	// RegHandler(NewNamedHandler("bench1", slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug})))
// 	opt.TopHandlers = append(opt.TopHandlers, NewNamedHandler("bench1", slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug})))
// 	// RegHandler(NewNamedHandler("bench2", NewNamedHandler("bench2.1", slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug}))))
// 	opt.TopHandlers = append(opt.TopHandlers, NewNamedHandler("bench2", NewNamedHandler("bench2.1", slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug}))))

// 	logger := BuildAgSlog(opt)

// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		logger.Debug("benchmark message")
// 	}
// }

// // BaseLevel Info
// // BenchmarkDifferentLevels/InfoLevel-8         	 1272805	       919.2 ns/op	     264 B/op	       7 allocs/op
// // BenchmarkDifferentLevels/DebugLevel-8        	 5294932	       232.1 ns/op	     264 B/op	       7 allocs/op
// // BenchmarkDifferentLevels/ErrorLevel-8        	 1275723	       907.4 ns/op	     264 B/op	       7 allocs/op
// // BenchmarkDifferentLevels/name-InfoLevel-8    	  744681	      1419 ns/op	     264 B/op	       7 allocs/op
// // BenchmarkDifferentLevels/name-DebugLevel-8   	 5188203	       232.7 ns/op	     264 B/op	       7 allocs/op
// // BenchmarkDifferentLevels/name-ErrorLevel-8   	  762216	      1450 ns/op	     264 B/op	       7 allocs/op
// // BenchmarkDifferentLevels/name2-InfoLevel-8   	  753090	      1604 ns/op	     264 B/op	       7 allocs/op
// // BenchmarkDifferentLevels/name2-DebugLevel-8  	 4849008	       272.2 ns/op	     264 B/op	       7 allocs/op
// // BenchmarkDifferentLevels/name2-ErrorLevel-8  	  681216	      1614 ns/op	     264 B/op	       7 allocs/op
// func BenchmarkDifferentLevels(b *testing.B) {
// 	opt := &AgSlogOption{}
// 	opt.TopHandlers = append(opt.TopHandlers, slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelInfo}))
// 	logger := BuildAgSlog(opt)

// 	b.Run("InfoLevel", func(b *testing.B) {
// 		for i := 0; i < b.N; i++ {
// 			logger.Info("info message")
// 		}
// 	})

// 	b.Run("DebugLevel", func(b *testing.B) {
// 		for i := 0; i < b.N; i++ {
// 			logger.Debug("debug message")
// 		}
// 	})

// 	b.Run("ErrorLevel", func(b *testing.B) {
// 		for i := 0; i < b.N; i++ {
// 			logger.Error("error message")
// 		}
// 	})

// 	opt2 := &AgSlogOption{}
// 	opt2.TopHandlers = append(opt.TopHandlers, NewNamedHandler("benche1", slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelInfo})))
// 	logger2 := BuildAgSlog(opt2)

// 	b.Run("name-InfoLevel", func(b *testing.B) {
// 		for i := 0; i < b.N; i++ {
// 			logger2.Info("info message")
// 		}
// 	})

// 	b.Run("name-DebugLevel", func(b *testing.B) {
// 		for i := 0; i < b.N; i++ {
// 			logger2.Debug("debug message")
// 		}
// 	})

// 	b.Run("name-ErrorLevel", func(b *testing.B) {
// 		for i := 0; i < b.N; i++ {
// 			logger2.Error("error message")
// 		}
// 	})

// 	opt3 := &AgSlogOption{}
// 	opt3.TopHandlers = append(opt.TopHandlers, NewNamedHandler("benche1", NewNamedHandler("benche1.1", slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelInfo}))))
// 	logger3 := BuildAgSlog(opt3)

// 	b.Run("name2-InfoLevel", func(b *testing.B) {
// 		for i := 0; i < b.N; i++ {
// 			logger3.Info("info message")
// 		}
// 	})

// 	b.Run("name2-DebugLevel", func(b *testing.B) {
// 		for i := 0; i < b.N; i++ {
// 			logger3.Debug("debug message")
// 		}
// 	})

// 	b.Run("name2-ErrorLevel", func(b *testing.B) {
// 		for i := 0; i < b.N; i++ {
// 			logger3.Error("error message")
// 		}
// 	})
// }
