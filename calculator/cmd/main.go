package main

import (
	"calculator/internal/handlers"
	"calculator/internal/storage"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

func main() {
	router := gin.Default()
	tasksDB, err := storage.NewTasksMongoStorage()
	if err != nil {
		log.Fatalln(err)
	}

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "it's okay"})
	})

	router.GET("/calculate/:task_name", handlers.HandleTaskCalculations(tasksDB))
	err = router.Run("0.0.0.0:8090")
	if err != nil {
		log.Fatalln(err)
	}

}
