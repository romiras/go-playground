package main

import (
	"fmt"
	"log"

	"net/http"
	_ "net/http/pprof"

	"github.com/gin-gonic/gin"
	"github.com/romiras/go-playground/gin-dwarf-api/helpers"
)

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	reg := helpers.NewRegistry()
	defer reg.Close()

	r := gin.Default()
	setupRoutes(r, reg)

	log.Fatal(r.Run()) // listen and serve on 0.0.0.0:8080
	// https://github.com/gin-gonic/examples/blob/master/graceful-shutdown/graceful-shutdown/server.go
}

func setupRoutes(r *gin.Engine, reg *helpers.Registry) {
	r.GET("/ping", helpers.WithRegistry(ping, reg))
	r.GET("/sleep", helpers.WithRegistry(sleep, reg))
}

func ping(c *gin.Context, reg *helpers.Registry) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}

func sleep(c *gin.Context, reg *helpers.Registry) {
	query := fmt.Sprintf("SELECT SLEEP(%0.2f)", reg.W.GetDelay())
	// fmt.Println(query)
	_, err := reg.DB.Exec(query)
	if err != nil {
		log.Print(err)
		c.JSON(500, gin.H{
			"message": "Error",
			"error":   err.Error(),
		})
		return
	}
	c.JSON(200, gin.H{
		"message": "OK",
	})
}
