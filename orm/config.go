package orm

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"time"

	"gorm.io/gorm/logger"

	"github.com/byteflowing/go-common/rotation"
	configv1 "github.com/byteflowing/proto/gen/go/config/v1"
	enumv1 "github.com/byteflowing/proto/gen/go/enums/v1"
)

func getLogWriter(config *configv1.DbLog) io.Writer {
	switch config.Out {
	case enumv1.LogOut_LOG_OUT_STDOUT:
		return os.Stdout
	case enumv1.LogOut_LOG_OUT_FILE:
		return rotation.NewRotation(config.Rotation)
	default:
		return os.Stdout
	}
}

func getLogLevel(config *configv1.DbLog) logger.LogLevel {
	switch config.Level {
	case enumv1.DbLogLevel_DB_LOG_LEVEL_SILENT:
		return logger.Silent
	case enumv1.DbLogLevel_DB_LOG_LEVEL_ERROR:
		return logger.Error
	case enumv1.DbLogLevel_DB_LOG_LEVEL_WARN:
		return logger.Warn
	case enumv1.DbLogLevel_DB_LOG_LEVEL_INFO:
		return logger.Info
	}
	return logger.Silent
}

func getLogConfig(config *configv1.DbLog) logger.Config {
	return logger.Config{
		SlowThreshold:             time.Duration(config.SlowThreshold) * time.Millisecond,
		Colorful:                  config.Colorful,
		IgnoreRecordNotFoundError: config.IgnoreRecordNotFoundErr,
		ParameterizedQueries:      config.ParameterizedQueries,
		LogLevel:                  getLogLevel(config),
	}
}

func getMaxIdleTime(config *configv1.DbConn) time.Duration {
	return time.Duration(config.MaxIdleTime) * time.Second
}

func getMaxIdleConnes(config *configv1.DbConn) int {
	return int(config.MaxIdleConnes)
}

func getMaxOpenConnes(config *configv1.DbConn) int {
	return int(config.MaxOpenConnes)
}

func getConnMaxLifetime(config *configv1.DbConn) time.Duration {
	return time.Duration(config.ConnMaxLifeTime) * time.Second
}

func getMySqlDSN(config *configv1.DbMysql) string {
	escapedLoc := url.QueryEscape(config.Location)
	const format = "%s:%s@tcp(%s:%d)/%s?parseTime=True&charset=%s&loc=%s&timeout=%v&readTimeout=%v&writeTimeout=%v"
	return fmt.Sprintf(
		format,
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.DbName,
		config.Charset,
		escapedLoc,
		time.Duration(config.ConnTimeout)*time.Second,
		time.Duration(config.ReadTimeout)*time.Second,
		time.Duration(config.WriteTimeout)*time.Second,
	)
}

func getPostgresSSLMode(config *configv1.DbPostgres) string {
	if config.SslMode {
		return "enable"
	}
	return "disable"
}

func getPostgresDSN(config *configv1.DbPostgres) string {
	const format = "host=%s, user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s search_path=%s"
	return fmt.Sprintf(
		format,
		config.Host,
		config.User,
		config.Password,
		config.DbName,
		config.Port,
		getPostgresSSLMode(config),
		config.TimeZone,
		config.Schema,
	)
}

func getSQLServerDSN(config *configv1.DbSQLServer) string {
	const format = "sqlserver://%s:%s@%s:%d?database=%s"
	return fmt.Sprintf(
		format,
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.DbName,
	)
}

func getSqliteDSN(config *configv1.DbSQLite) string {
	return config.DbPath
}
