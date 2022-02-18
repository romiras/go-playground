package helpers

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

type Registry struct {
	DB *sql.DB
	W  *Waiter
}

func NewRegistry() *Registry {
	return &Registry{
		W:  NewWaiter(),
		DB: newDB(),
	}
}

func (reg *Registry) Close() {
	fmt.Println("Close")
	reg.DB.Close()
}

func newDB() *sql.DB {
	db, err := sql.Open("mysql", os.Getenv("DSN"))
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func WithRegistry(handler func(*gin.Context, *Registry), reg *Registry) func(*gin.Context) {
	return func(ctx *gin.Context) {
		handler(ctx, reg)
	}
}
