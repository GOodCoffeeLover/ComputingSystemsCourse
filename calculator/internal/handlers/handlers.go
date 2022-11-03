package handlers

import (
	"calculator/internal/core"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

type taskGetter interface {
	Get(taskName string) (task core.Task, err error)
}

func HandleTaskCalculations(tasks taskGetter) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		targetTaskName := ctx.Param("task_name")

		task, err := tasks.Get(targetTaskName)

		if err != nil {
			ctx.AbortWithError(http.StatusConflict, fmt.Errorf("can't get task due to %v", err))
			return
		}
		log.Println(task)
		res, err := task.StartCalculation()
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("can't complite calculations due to %v", err))
			return
		}
		log.Printf("Calculation result: %v", res)
		ctx.JSON(http.StatusOK, gin.H{"Answer": res})

	}
}
