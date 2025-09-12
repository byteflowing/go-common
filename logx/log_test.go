package logx

import (
	"context"
	"testing"

	"github.com/byteflowing/go-common/idx"
	configv1 "github.com/byteflowing/proto/gen/go/config/v1"
	enumv1 "github.com/byteflowing/proto/gen/go/enums/v1"
	"go.uber.org/zap"
)

func TestLogger_MultiOutputLevels(t *testing.T) {
	c := &configv1.ZapLogConfig{
		Mode:               enumv1.LogMode_LOG_MODE_DEV,
		Format:             enumv1.LogFormat_LOG_FORMAT_CONSOLE,
		ReportCaller:       true,
		ShortCaller:        true,
		CallerSkip:         1,
		ServiceName:        "user",
		AddStackTraceLevel: enumv1.LogLevel_LOG_LEVEL_ERROR,
		Level:              enumv1.LogLevel_LOG_LEVEL_DEBUG,
		CtxLogIdKey:        "log_id",
		LogIdKey:           "log_id",
		//Outputs: []*logv1.LogOutput{
		//	{
		//		Output: enumv1.LogOut_LOG_OUT_STDOUT,
		//		Levels: []enumv1.LogLevel{
		//			enumv1.LogLevel_LOG_LEVEL_WARN,
		//			enumv1.LogLevel_LOG_LEVEL_ERROR,
		//		},
		//		LogFile: nil,
		//	},
		//},
	}
	Init(c)
	//fmt.Println(std.Level())
	//Debug("std debug")
	//Info("std info")
	//Warn("std warn")
	//Error("std error")
	//
	logid := idx.UUIDv4()
	ctx := CtxWithLogID(context.Background(), logid)
	CtxDebug(ctx, "debug1", zap.String("key1", "value1"))
	CtxDebug(ctx, "debug2", zap.String("key2", "value2"))
	CtxDebug(ctx, "debug3", zap.String("key2", "value2"))
	CtxDebug(ctx, "debug4", zap.String("key2", "value2"))
	//CtxInfo(ctx, "info")
	//CtxWarn(ctx, "warn")
	//CtxError(ctx, "error")
	//CtxFatal(ctx, "fatal")

	//logger := newZap(c)
	//logger.Debug("debug message")
	//logger.Info("info message")
	//logger.Warn("warn message")
	//logger.Error("error message")

}
