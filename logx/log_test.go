package logx

import (
	"testing"

	enumv1 "github.com/byteflowing/proto/gen/go/enums/v1"
	logv1 "github.com/byteflowing/proto/gen/go/log/v1"
)

func TestLogger_MultiOutputLevels(t *testing.T) {
	c := &logv1.LogConfig{
		Mode:               enumv1.LogMode_LOG_MODE_DEV,
		Format:             enumv1.LogFormat_LOG_FORMAT_CONSOLE,
		ReportCaller:       false,
		ShortCaller:        true,
		CallerSkip:         1,
		ServiceName:        "",
		AddStackTraceLevel: enumv1.LogLevel_LOG_LEVEL_ERROR,
		Level:              enumv1.LogLevel_LOG_LEVEL_DEBUG,
		Outputs: []*logv1.LogOutput{
			{
				Output: enumv1.LogOut_LOG_OUT_STDOUT,
				Levels: []enumv1.LogLevel{
					enumv1.LogLevel_LOG_LEVEL_WARN,
					enumv1.LogLevel_LOG_LEVEL_ERROR,
				},
				LogFile: nil,
			},
		},
	}

	//std = newZap(c)
	//fmt.Println(std.Level())
	//Debug("std debug")
	//Info("std info")
	//Warn("std warn")
	//Error("std error")

	// 打一些不同级别日志
	c.CallerSkip = 0
	logger := newZap(c)
	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")

}
