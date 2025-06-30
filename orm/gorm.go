package orm

import (
	"gorm.io/gorm"
	"gorm.io/rawsql"
)

func New(c *Config) *gorm.DB {
	var db *gorm.DB
	switch c.GetDatabaseType() {
	case MySQL:
		db = initMySQL(c)
	case Postgres:
		db = initPostgres(c)
	case SQLServer:
		db = initSQLServer(c)
	case SQLite:
		db = initSQLite(c)
	default:
		panic("unknown database type:" + c.DbType)
	}
	if c.GetDatabaseType() != SQLite && c.Conn != nil {
		sqlDb, err := db.DB()
		if err != nil {
			panic(err)
		}
		sqlDb.SetConnMaxIdleTime(c.Conn.GetMaxIdleTime())
		sqlDb.SetConnMaxLifetime(c.Conn.GetConnMaxLifetime())
		sqlDb.SetMaxIdleConns(c.Conn.GetMaxIdleConnes())
		sqlDb.SetMaxOpenConns(c.Conn.GetMaxOpenConnes())
	}
	return db
}

// NewBySQL 通过SQL创建db,主要用于gen生成struct
func NewBySQL(c *SQLConfig) *gorm.DB {
	conf := rawsql.Config{
		DriverName: string(getDBType(c.DBType)),
		FilePath:   c.FilePath,
		SQL:        c.SQL,
	}
	gormDB, err := gorm.Open(rawsql.New(conf))
	if err != nil {
		panic(err)
	}
	return gormDB
}

func getDBType(t string) DatabaseType {
	var dbType DatabaseType
	switch t {
	case "mysql", "":
		dbType = MySQL
	case "postgres":
		dbType = Postgres
	case "sqlserver":
		dbType = SQLServer
	case "sqlite":
		dbType = SQLite
	default:
		panic("unknown database type:" + t)
	}
	return dbType
}
