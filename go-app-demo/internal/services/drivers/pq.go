package drivers

import (
	"database/sql"

	_ "github.com/lib/pq"
	"github.com/romiras/go-app-demo/internal/logger"
	"github.com/spf13/viper"
)

// https://pkg.go.dev/github.com/lib/pq

type DB struct {
	conn   *sql.DB
	logger logger.Logger
}

func NewDB(v *viper.Viper, logger logger.Logger) *DB {
	connStr := "" // "user=pqgotest dbname=pqgotest sslmode=verify-full"
	conn := createConnection(connStr, logger)
	return &DB{
		conn:   conn,
		logger: logger,
	}
}

func createConnection(dsn string, logger logger.Logger) *sql.DB {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		logger.Panic(err)
	}

	db.SetMaxIdleConns(100)
	db.SetMaxOpenConns(100)

	return db
}
