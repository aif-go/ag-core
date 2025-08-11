package logzap

import (
	"ag-core/ag/ag_conf"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

const ctxLoggerKey = "zapLogger"

type ZlogProperties struct {
	LogFileName string `value:"${log_file_name:}"`
	LogLevel    string `value:"${log_level:}"`
	MaxSize     int    `value:"${max_size:100}"`
	MaxBackUps  int    `value:"${max_backups:0}"`
	Compress    bool   `value:"${conpress:false}"`
	MaxAge      int    `value:"${max_age:0}"`
	Console     bool   `value:"${console:false}"`
	Prod        bool   `value:"${prod:false}"`
}

func BindZlogProperties(binder ag_conf.IBinder, lced ag_conf.LocalConfigLoded) (*ZlogProperties, error) {
	zlogp := ZlogProperties{}
	// err := ag_conf.Binder.Bind(&zlogp, "log")
	err := binder.Bind(&zlogp, "log")
	if err != nil {
		return nil, err
	}
	return &zlogp, nil
}

func NewZapLogP(p *ZlogProperties) *zap.Logger {
	lp := p.LogFileName
	lv := p.LogLevel
	var level zapcore.Level
	//debug<info<warn<error<fatal<panic
	switch lv {
	case "debug":
		level = zap.DebugLevel
	case "info":
		level = zap.InfoLevel
	case "warn":
		level = zap.WarnLevel
	case "error":
		level = zap.ErrorLevel
	default:
		level = zap.InfoLevel
	}
	ws := []zapcore.WriteSyncer{}
	ws = append(ws, zapcore.AddSync(os.Stdout))
	// 多套hook 对应多个zaplogger -->sloger [rpclogger,tradeinfologger,heartbealogger]
	if lp != "" {
		hook := lumberjack.Logger{
			Filename:   lp,           // Log file path
			MaxSize:    p.MaxSize,    // Maximum size unit for each log file: M
			MaxBackups: p.MaxBackUps, // The maximum number of backups that can be saved for log files
			MaxAge:     p.MaxAge,     // Maximum number of days the file can be saved
			Compress:   p.Compress,   // Compression or not
		}

		ws = append(ws, zapcore.AddSync(&hook))
	}

	var encoder zapcore.Encoder
	if p.Console {
		encoder = zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
			TimeKey:        "ts",
			LevelKey:       "level",
			NameKey:        "Logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseColorLevelEncoder,
			EncodeTime:     timeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			// EncodeCaller:   zapcore.FullCallerEncoder,
			EncodeCaller: zapcore.ShortCallerEncoder,
		})
	} else {
		encoder = zapcore.NewJSONEncoder(zapcore.EncoderConfig{
			TimeKey:        "ts",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			FunctionKey:    zapcore.OmitKey,
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.EpochTimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		})
	}
	core := zapcore.NewCore(
		encoder,
		//zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(&hook)), // Print to console and file
		zapcore.NewMultiWriteSyncer(ws...), // Print to console and file
		level,
	)
	if p.Prod {
		// 放入多个
		return zap.New(zapcore.NewTee(core), zap.Development(), zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))
	}
	return zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))

}

func timeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	//enc.AppendString(t.Format("2006-01-02 15:04:05"))
	enc.AppendString(t.Format("2006-01-02 15:04:05.000000000"))
}

func NewZapLog() (*zap.Logger, error) {
	// TODO 构建zaplog
	log, err := zap.NewProduction()

	return log, err
	// return nil, nil
}
