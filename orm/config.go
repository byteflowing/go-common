package orm

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"time"

	"gorm.io/gorm/logger"

	dbv1 "github.com/byteflowing/go-common/gen/db/v1"
	enumv1 "github.com/byteflowing/go-common/gen/enums/v1"
	"github.com/byteflowing/go-common/rotation"
)

func getLogWriter(config *dbv1.DbLog) io.Writer {
	switch config.Out {
	case enumv1.DbLogOut_DB_LOG_OUT_STDOUT:
		return os.Stdout
	case enumv1.DbLogOut_DB_LOG_OUT_FILE:
		return rotation.NewRotation(config.Rotation)
	default:
		return os.Stdout
	}
}

func getLogLevel(config *dbv1.DbLog) logger.LogLevel {
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

func getLogConfig(config *dbv1.DbLog) logger.Config {
	return logger.Config{
		SlowThreshold:             time.Duration(config.SlowThreshold) * time.Millisecond,
		Colorful:                  config.Colorful,
		IgnoreRecordNotFoundError: config.IgnoreRecordNotFoundErr,
		ParameterizedQueries:      config.ParameterizedQueries,
		LogLevel:                  getLogLevel(config),
	}
}

func getMaxIdleTime(config *dbv1.DbConn) time.Duration {
	return time.Duration(config.MaxIdleTime) * time.Second
}

func getMaxIdleConnes(config *dbv1.DbConn) int {
	return int(config.MaxIdleConnes)
}

func getMaxOpenConnes(config *dbv1.DbConn) int {
	return int(config.MaxOpenConnes)
}

func getConnMaxLifetime(config *dbv1.DbConn) time.Duration {
	return time.Duration(config.ConnMaxLifeTime) * time.Second
}

func getMySqlDSN(config *dbv1.DbMysql) string {
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

func getPostgresSSLMode(config *dbv1.DbPostgres) string {
	if config.SslMode {
		return "enable"
	}
	return "disable"
}

func getPostgresDSN(config *dbv1.DbPostgres) string {
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

func getSQLServerDSN(config *dbv1.DbSQLServer) string {
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

func getSqliteDSN(config *dbv1.DbSQLite) string {
	return config.DbPath
}
