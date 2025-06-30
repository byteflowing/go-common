package rotation

import "gopkg.in/natefinch/lumberjack.v2"

type Config struct {
	LogFile    string `json:"logFile"`    // 文件名
	MaxSize    int    `json:"maxSize"`    // 文件大小
	MaxAge     int    `json:"maxAge"`     // 一个文件记录日志时长
	MaxBackups int    `json:"maxBackups"` // 保留几份日志
	Compress   bool   `json:"compress"`   // 是否启用压缩
	LocalTime  bool   `json:"localTime"`  // 是否使用本地时间
}

func NewRotation(opts *Config) *lumberjack.Logger {
	if opts == nil {
		panic("Config must not be nil")
	}
	return &lumberjack.Logger{
		Filename:   opts.LogFile,
		MaxSize:    opts.MaxSize,
		MaxAge:     opts.MaxAge,
		MaxBackups: opts.MaxBackups,
		LocalTime:  opts.LocalTime,
		Compress:   opts.Compress,
	}
}
