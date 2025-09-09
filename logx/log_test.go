package logx

import (
	"context"
	"testing"

	"github.com/byteflowing/go-common/idx"
	enumv1 "github.com/byteflowing/proto/gen/go/enums/v1"
	logv1 "github.com/byteflowing/proto/gen/go/log/v1"
)

func TestLogger_MultiOutputLevels(t *testing.T) {
	c := &logv1.LogConfig{
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
	ctx := WithLogID(context.Background(), logid)
	CtxDebug(ctx, "debug")
	CtxInfo(ctx, "info")
	CtxWarn(ctx, "warn")
	CtxError(ctx, "error")
	CtxFatal(ctx, "fatal")

	//logger := newZap(c)
	//logger.Debug("debug message")
	//logger.Info("info message")
	//logger.Warn("warn message")
	//logger.Error("error message")

}
