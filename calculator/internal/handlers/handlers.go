package handlers

import (
	"calculator/internal/core"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

func HandleTaskCalculations() func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		data, err := ctx.GetRawData()
		if err != nil {
			ctx.AbortWithError(http.StatusConflict, fmt.Errorf("can't get raw data due to %v", err))
			return
		}
		task := core.Task{}

		if err = json.Unmarshal(data, &task); err != nil {
			ctx.AbortWithError(http.StatusConflict, fmt.Errorf("can't unmarshal raw data due to %v", err))
			return
		}
		fmt.Println(task)
		res, err := task.StartCalculation()
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("can't complite calculations due to %v", err))
			return
		}
		fmt.Printf("Calculation result: %v", res)
		ctx.JSON(http.StatusOK, gin.H{"Answer": res})

	}
}
