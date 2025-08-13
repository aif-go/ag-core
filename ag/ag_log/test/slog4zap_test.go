package test

//func TestLogFileSpit(t *testing.T) {
//zapLogger := newZap3()
//defer zapLogger.Sync()
//handler := ag_log.NewZapHandler(zapLogger)
//logger := ag_log.NewZapSlog(handler)
//logger.Info("***", slog.Group("group", "hello", "word"))
//}

//func TestSlog4zap(t *testing.T) {
////zapLogger := newZap1()
////zapLogger := newZap2()
//zapLogger := newZap3()
//defer zapLogger.Sync()

//// 创建一个基于 zap 的 slog 处理器
//handler := ag_log.NewZapHandler(zapLogger)
//logger := ag_log.NewZapSlog(handler)
//logger.Info("This is an info log", "key", "value")

//// 使用 slog 进行日志记录
//slog.Info("This is an info log", "key", "value")
//slog.Warn("This is a warning log", "count", 10)
////	slog.Error("This is an error log", "err", "something went wrong")

//// 记录带有时间的日志
//slog.Info("Log with time", "time", time.Now())
//}

//func newZap() *zap.Logger {
//return newZap1()
//}

//func newZap1() *zap.Logger {
//// 创建一个启用调用者信息的 zap 日志实例
//zapConfig := zap.NewProductionConfig()
//zapConfig.Encoding = "console"
//zapConfig.DisableCaller = false
//zapConfig.EncoderConfig.CallerKey = "caller" // 设置 caller 字段名
//zapLogger, _ := zapConfig.Build()
//return zapLogger
//}

//func newZap2() *zap.Logger {
//encoderConfig := zapcore.EncoderConfig{
//TimeKey:        "ts",
//LevelKey:       "level",
//NameKey:        "logger",
//CallerKey:      "caller",
//MessageKey:     "msg",
//StacktraceKey:  "stacktrace",
//LineEnding:     zapcore.DefaultLineEnding,
//EncodeDuration: zapcore.SecondsDurationEncoder,
//EncodeTime:     zapcore.ISO8601TimeEncoder,
//EncodeLevel:    zapcore.LowercaseLevelEncoder,
//EncodeCaller:   zapcore.ShortCallerEncoder,
//}
//zapConfig := &zap.Config{
////Level:             zap.NewAtomicLevelAt(zapcore.Level(level)),
//Level:             zap.NewAtomicLevelAt(zapcore.DebugLevel),
//Development:       true,
//DisableCaller:     false,
//DisableStacktrace: true,
//Sampling:          &zap.SamplingConfig{Initial: 100, Thereafter: 100},
//Encoding:          "json",
//EncoderConfig:     encoderConfig,
//OutputPaths:       []string{"stderr"},
//ErrorOutputPaths:  []string{"stderr"},
//}
//l, err := zapConfig.Build(zap.AddCallerSkip(3))
//if err != nil {
//fmt.Printf("zap build logger fail err=%v", err)
//return nil
//}
//return l
//}

//func newZap3() *zap.Logger {
//// conf := viper.New()
//// conf.SetDefault("hello", "hello")
//// //	conf.SetDefault("log.log_file_name", "hzwzaptest.log")
//// conf.SetDefault("log.log_level", "debug")
//// conf.SetDefault("log.encoding", "console")
////conf.SetDefault("env", "prod")
//env, _ := ag_conf.NewStandardEnvironment()
//mpp := &ag_conf.MapPropertySource{
//NamedPropertySource: ag_conf.NamedPropertySource{
//Name: "logtest",
//},
//Source: map[string]any{
//"hello":         "hello",
//"log.log_level": "debug",
//},
//}
//env.GetPropertySources().AddLast(mpp)

//zaplog := ag_log.NewZapLog(env)
//return zaplog
//}

//func testEnv() map[string]any {

//// res := make(map[string]any, 10)
//bytearr, err := os.ReadFile("test.yaml")
//if err != nil {
//panic(err)
//}
//mapcontext := make(map[string]any)
//err = yaml.Unmarshal(bytearr, mapcontext)
//if err != nil {
//panic(err)
//}

//mapcontext, err = ag_ext.GetFlattenedMap(mapcontext)
//if err != nil {
//panic(err)
//}
//return mapcontext

//}
