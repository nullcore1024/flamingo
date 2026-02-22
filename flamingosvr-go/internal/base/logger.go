package base

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"sync"
)

type Logger struct {
	log *zap.Logger
}

var (
	loggerInstance *Logger
	loggerOnce     sync.Once
)

func GetLogger() *Logger {
	loggerOnce.Do(func() {
		// 配置编码器
		encoderConfig := zap.NewProductionEncoderConfig()
		encoderConfig.TimeKey = "timestamp"
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

		// 创建Core
		core := zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			zapcore.AddSync(os.Stdout),
			zapcore.InfoLevel,
		)

		// 创建Logger
		loggerInstance = &Logger{
			log: zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1)),
		}
	})
	return loggerInstance
}

func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.log.Info(msg, fields...)
}

func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.log.Error(msg, fields...)
}

func (l *Logger) Debug(msg string, fields ...zap.Field) {
	l.log.Debug(msg, fields...)
}

func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.log.Warn(msg, fields...)
}

func (l *Logger) Fatal(msg string, fields ...zap.Field) {
	l.log.Fatal(msg, fields...)
}

// AsyncLog 异步日志接口
func AsyncLog() *Logger {
	return GetLogger()
}
