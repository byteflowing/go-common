package logx

import (
	"context"
	"sync"

	enumv1 "github.com/byteflowing/proto/gen/go/enums/v1"
	logv1 "github.com/byteflowing/proto/gen/go/log/v1"
	"go.uber.org/zap"
)

var (
	std       *zap.Logger
	once      sync.Once
	stdConfig *StdConfig
)

type StdConfig struct {
	CtxLogIdKey string
	LogIdKey    string
}

func WithLogID(ctx context.Context, logID string) context.Context {
	return context.WithValue(ctx, stdConfig.CtxLogIdKey, logID)
}

func GetLogID(ctx context.Context) string {
	if stdConfig != nil && stdConfig.CtxLogIdKey != "" {
		if v, ok := ctx.Value(stdConfig.CtxLogIdKey).(string); ok {
			return v
		}
	}
	return ""
}

// Init CallerSkip设置为1刚好可以显示记录日志的那行代码
func Init(config *logv1.LogConfig) {
	once.Do(func() {
		std = newZap(config)
		stdConfig = &StdConfig{
			CtxLogIdKey: config.CtxLogIdKey,
			LogIdKey:    config.LogIdKey,
		}
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
	stdConfig = &StdConfig{}
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

func CtxDebug(ctx context.Context, msg string, fields ...zap.Field) {
	fs := addLogIdToFields(ctx, fields...)
	std.Debug(msg, fs...)
}

func CtxInfo(ctx context.Context, msg string, fields ...zap.Field) {
	fs := addLogIdToFields(ctx, fields...)
	std.Info(msg, fs...)
}

func CtxWarn(ctx context.Context, msg string, fields ...zap.Field) {
	fs := addLogIdToFields(ctx, fields...)
	std.Warn(msg, fs...)
}

func CtxError(ctx context.Context, msg string, fields ...zap.Field) {
	fs := addLogIdToFields(ctx, fields...)
	std.Error(msg, fs...)
}

func CtxPanic(ctx context.Context, msg string, fields ...zap.Field) {
	fs := addLogIdToFields(ctx, fields...)
	std.Panic(msg, fs...)
}

func CtxFatal(ctx context.Context, msg string, fields ...zap.Field) {
	fs := addLogIdToFields(ctx, fields...)
	std.Fatal(msg, fs...)
}

func addLogIdToFields(ctx context.Context, fields ...zap.Field) []zap.Field {
	logId := GetLogID(ctx)
	logFieldName := stdConfig.LogIdKey
	if logId == "" || logFieldName == "" {
		return fields
	}
	fields = append(fields, zap.String(logFieldName, logId))
	return fields
}
