package main

import (
	"ComputingSystemsCourse/internal/handlers"
	"ComputingSystemsCourse/internal/storage"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"sync"
)

func main() {
	tasks := storage.NewTaskMapStorage()
	mutex := sync.Mutex{}
	router := gin.Default()

	router.GET("/check", func(context *gin.Context) {
		context.JSON(http.StatusOK, gin.H{"message": "It's okay"})
	})

	router.GET("/task/:task_name", handlers.HandleTaskAccess(tasks, mutex))
	router.GET("/task/calculate/:task_name", handlers.HandleCalculation(tasks, mutex))
	router.POST("/task", handlers.HandleTaskCreation(tasks, mutex))
	router.POST("/task/:task_name", handlers.HandleTaskUpdate(tasks, mutex))
	router.DELETE("/task/:task_name", handlers.HandleTaskDelete(tasks, mutex))

	router.GET("/work/:task_name/:work_name", handlers.HandleWorkAccess(tasks, mutex))
	router.POST("/work/:task_name/:work_name", handlers.HandleWorkNeedsSetup(tasks, mutex))
	router.POST("/work/:task_name", handlers.HandleWorkCreation(tasks, mutex))
	router.DELETE("/work/:task_name/:work_name", handlers.HandleWorkDelete(tasks, mutex))

	err := router.Run(":8080")
	if err != nil {
		log.Fatalln(err)
	}

}
