package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"os"
)

func main() {
	fmt.Println("Print from the Go program")
	fmt.Println(os.Getenv("PORT_ENV"))

	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	_ = r.Run(":8084")

}
