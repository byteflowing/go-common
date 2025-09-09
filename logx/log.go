package logx

import (
	"sync"

	enumv1 "github.com/byteflowing/proto/gen/go/enums/v1"
	logv1 "github.com/byteflowing/proto/gen/go/log/v1"
	"go.uber.org/zap"
)

var (
	std  *zap.Logger
	once sync.Once
)

func Init(config *logv1.LogConfig) {
	once.Do(func() {
		config.CallerSkip = 1
		std = newZap(config)
	})
}

func init() {
	defaultConfig := &logv1.LogConfig{
		Mode:               enumv1.LogMode_LOG_MODE_DEV,
		Format:             enumv1.LogFormat_LOG_FORMAT_CONSOLE,
		Level:              enumv1.LogLevel_LOG_LEVEL_INFO,
		ReportCaller:       true,
		ShortCaller:        true,
		CallerSkip:         1,
		AddStackTraceLevel: enumv1.LogLevel_LOG_LEVEL_ERROR,
	}
	conf := getConfig(defaultConfig)
	opts := getOptions(defaultConfig)
	logger, err := conf.Build(opts...)
	if err != nil {
		panic(err)
	}
	std = logger
}

func NewZapLogger(config *logv1.LogConfig) *zap.Logger {
	return newZap(config)
}

func GetStdLogger() *zap.Logger {
	return std
}

func Sync() error {
	return std.Sync()
}

func Debug(msg string, fields ...zap.Field) {
	std.Debug(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	std.Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	std.Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	std.Error(msg, fields...)
}

func Panic(msg string, fields ...zap.Field) {
	std.Panic(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	std.Fatal(msg, fields...)
}
