package main

import (
	"calculator/internal/handlers"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

func main() {
	router := gin.Default()
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "it's okay"})
	})
	router.GET("/calculate", handlers.HandleTaskCalculations())
	err := router.Run("0.0.0.0:8090")
	if err != nil {
		log.Fatalln(err)
	}

}
