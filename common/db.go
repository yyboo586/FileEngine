package common

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

func NewDB(cfg *Config) (dbPool *sql.DB, err error) {
	dsn := fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?parseTime=true", cfg.DB.User, cfg.DB.Password, cfg.DB.Host, cfg.DB.Port, cfg.DB.DBName)

	if dbPool, err = sql.Open("mysql", dsn); err != nil {
		return nil, fmt.Errorf("NewDB(): failed to open database, error: %w", err)
	}

	if err = dbPool.Ping(); err != nil {
		return nil, fmt.Errorf("NewDB(): failed to ping database, error: %w", err)
	}

	return dbPool, nil
}
