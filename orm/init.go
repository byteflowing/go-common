package orm

import (
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	dbv1 "github.com/byteflowing/go-common/gen/db/v1"
)

func initMySQL(c *dbv1.DbConfig) *gorm.DB {
	db, err := gorm.Open(mysql.Open(getMySqlDSN(c.Mysql)), getGormConfig(c))
	if err != nil {
		panic(err)
	}
	return db
}

func initPostgres(c *dbv1.DbConfig) *gorm.DB {
	db, err := gorm.Open(postgres.Open(getPostgresDSN(c.Postgres)), getGormConfig(c))
	if err != nil {
		panic(err)
	}
	return db
}

func initSQLServer(c *dbv1.DbConfig) *gorm.DB {
	db, err := gorm.Open(sqlserver.Open(getSQLServerDSN(c.Sqlserver)), getGormConfig(c))
	if err != nil {
		panic(err)
	}
	return db
}

func initSQLite(c *dbv1.DbConfig) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(getSqliteDSN(c.Sqlite)), getGormConfig(c))
	if err != nil {
		panic(err)
	}
	return db
}

func getGormConfig(c *dbv1.DbConfig) *gorm.Config {
	config := &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
		Logger:                 logger.Default,
	}
	if c.Log == nil {
		config.Logger = logger.Default.LogMode(logger.Silent)
	} else {
		config.Logger = logger.New(log.New(getLogWriter(c.Log), "\r\n", log.LstdFlags), getLogConfig(c.Log))
	}
	return config
}
