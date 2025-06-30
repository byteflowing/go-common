package orm

import (
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func initMySQL(c *Config) *gorm.DB {
	db, err := gorm.Open(mysql.Open(c.MySQL.GetDSN()), getGormConfig(c))
	if err != nil {
		panic(err)
	}
	return db
}

func initPostgres(c *Config) *gorm.DB {
	db, err := gorm.Open(postgres.Open(c.Postgres.GetDSN()), getGormConfig(c))
	if err != nil {
		panic(err)
	}
	return db
}

func initSQLServer(c *Config) *gorm.DB {
	db, err := gorm.Open(sqlserver.Open(c.SQLServer.GetDSN()), getGormConfig(c))
	if err != nil {
		panic(err)
	}
	return db
}

func initSQLite(c *Config) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(c.SQLite.GetDSN()))
	if err != nil {
		panic(err)
	}
	return db
}

func getGormConfig(c *Config) *gorm.Config {
	config := &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
		Logger:                 logger.Default,
	}
	if c.Log == nil {
		config.Logger = logger.Default.LogMode(logger.Silent)
	} else {
		config.Logger = logger.New(log.New(c.Log.getLogWriter(), "\r\n", log.LstdFlags), c.Log.getLogConfig())
	}
	return config
}
