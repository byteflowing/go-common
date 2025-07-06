package orm

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"time"

	"gorm.io/gorm/logger"

	"github.com/byteflowing/go-common/rotation"
)

type Config struct {
	DbType    string `default:"mysql"`
	Log       *LogConfig
	Conn      *ConnConfig
	MySQL     *MySQLConfig
	Postgres  *PostgresConfig
	SQLServer *SQLServerConfig
	SQLite    *SQLiteConfig
}

type SQLConfig struct {
	DBType   string   `default:"mysql"` // 数据库类型 mysql postgres sqlserver sqlite
	SQL      []string // sql语句
	FilePath []string // sql文件或者目录
}

type LogConfig struct {
	// 慢日志阈值，单位ms
	SlowThreshold uint `default:"200""`
	// 输出
	// stdout, file
	Out string `default:"stdout"`
	// 是否彩色打印日志
	Colorful bool `default:"true"`
	// 忽略RecordNotFoundError
	IgnoreRecordNotFoundError bool `default:"false"`
	// 不在日志中打印参数
	ParameterizedQueries bool `default:"false"`
	// 日志级别
	// silent
	// error
	// warn
	// info
	Level string `default:"silent"`
	// 日志轮转
	LogRotation *LogRotationConfig
}

func (l *LogConfig) getLogWriter() io.Writer {
	switch l.Out {
	case "stdout":
		return os.Stdout
	case "file":
		return rotation.NewRotation(&rotation.Config{
			LogFile:    l.LogRotation.LogFile,
			MaxSize:    l.LogRotation.MaxSize,
			MaxAge:     l.LogRotation.MaxAge,
			MaxBackups: l.LogRotation.MaxBackups,
			Compress:   l.LogRotation.Compress,
			LocalTime:  l.LogRotation.LocalTime,
		})
	default:
		return os.Stdout
	}
}

func (l *LogConfig) getLogLevel() logger.LogLevel {
	switch l.Level {
	case "silent":
		return logger.Silent
	case "error":
		return logger.Error
	case "warn":
		return logger.Warn
	case "info":
		return logger.Info
	}
	return logger.Silent
}

func (l *LogConfig) getLogConfig() logger.Config {
	return logger.Config{
		SlowThreshold:             time.Duration(l.SlowThreshold) * time.Millisecond,
		Colorful:                  l.Colorful,
		IgnoreRecordNotFoundError: l.IgnoreRecordNotFoundError,
		ParameterizedQueries:      l.ParameterizedQueries,
		LogLevel:                  l.getLogLevel(),
	}
}

type LogRotationConfig struct {
	LogFile    string `json:"logFile"`    // 文件名
	MaxSize    int    `json:"maxSize"`    // 文件大小
	MaxAge     int    `json:"maxAge"`     // 一个文件记录日志时长
	MaxBackups int    `json:"maxBackups"` // 保留几份日志
	Compress   bool   `json:"compress"`   // 是否启用压缩
	LocalTime  bool   `json:"localTime"`  // 是否使用本地时间
}

func (c *Config) GetDatabaseType() DatabaseType {
	return getDBType(c.DbType)
}

type DatabaseType string

const (
	MySQL     DatabaseType = "mysql"
	Postgres  DatabaseType = "postgres"
	SQLServer DatabaseType = "sqlserver"
	SQLite    DatabaseType = "sqlite"
)

func (db DatabaseType) String() string {
	return string(db)
}

type ConnConfig struct {
	ConnMaxLifetime int `default:"1800"` // 单位：秒
	MaxIdleTime     int `default:"600"`  // 单位：秒
	MaxIdleConnes   int `default:"20"`   // 最大空闲连接
	MaxOpenConnes   int `default:"100"`  // 最大打开连接
}

func (c *ConnConfig) GetMaxIdleTime() time.Duration {
	return time.Duration(c.MaxIdleTime) * time.Second
}

func (c *ConnConfig) GetMaxIdleConnes() int {
	return c.MaxIdleConnes
}

func (c *ConnConfig) GetMaxOpenConnes() int {
	return c.MaxOpenConnes
}

func (c *ConnConfig) GetConnMaxLifetime() time.Duration {
	return time.Duration(c.ConnMaxLifetime) * time.Second
}

type MySQLConfig struct {
	Host         string // 数据库地址
	User         string // 数据库用户名
	Password     string // 数据库密码
	DBName       string // 数据库名
	Port         int    `default:"3306"`    // 端口号
	Charset      string `default:"utf8mb4"` // 字符集
	Location     string `default:"Local"`   // 时区
	ConnTimeout  int    `default:"30"`      // 单位：秒
	ReadTimeout  int    `default:"30"`      // 单位：秒
	WriteTimeout int    `default:"30"`      // 单位：秒
}

func (m *MySQLConfig) GetDSN() string {
	escapedLoc := url.QueryEscape(m.Location)
	const format = "%s:%s@tcp(%s:%d)/%s?parseTime=True&charset=%s&loc=%s&timeout=%v&readTimeout=%v&writeTimeout=%v"
	return fmt.Sprintf(
		format,
		m.User,
		m.Password,
		m.Host,
		m.Port,
		m.DBName,
		m.Charset,
		escapedLoc,
		time.Duration(m.ConnTimeout)*time.Second,
		time.Duration(m.ReadTimeout)*time.Second,
		time.Duration(m.WriteTimeout)*time.Second,
	)
}

type PostgresConfig struct {
	Host     string // 数据库地址
	User     string // 数据库用户名
	Password string // 数据库密码
	DBName   string // 数据库名
	SSLMode  bool   // 是否使用ssl
	Port     int    // 端口号
	TimeZone string // 时区
	Schema   string // schema
}

func (p *PostgresConfig) GetDSN() string {
	const format = "host=%s, user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s search_path=%s"
	return fmt.Sprintf(
		format,
		p.Host,
		p.User,
		p.Password,
		p.DBName,
		p.Port,
		p.getSSLMode(),
		p.TimeZone,
		p.Schema,
	)
}

func (p *PostgresConfig) getSSLMode() string {
	switch p.SSLMode {
	case true:
		return "enable"
	case false:
		return "disable"
	default:
		return "disable"
	}
}

type SQLServerConfig struct {
	Host     string // 数据库地址
	User     string // 数据库用户名
	Password string // 数据库密码
	DBName   string // 数据库名
	Port     int    `default:"1433"` // 端口号
}

func (s *SQLServerConfig) GetDSN() string {
	const format = "sqlserver://%s:%s@%s:%d?database=%s"
	return fmt.Sprintf(
		format,
		s.User,
		s.Password,
		s.Host,
		s.Port,
		s.DBName,
	)
}

type SQLiteConfig struct {
	DBPath string
}

func (s *SQLiteConfig) GetDSN() string {
	return s.DBPath
}
