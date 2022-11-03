package handlers

import (
	"calculator/internal/core"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type taskGetter interface {
	Get(taskName string) (task core.Task, err error)
}

func HandleTaskCalculations(tasks taskGetter) func(ctx *gin.Context) {
	redisAddres := os.Getenv("REDIS_ADDRESS")
	if redisAddres == "" {
		log.Fatal("don't have redis address")
	} else {
		log.Printf("Redis Address: %v", redisAddres)
	}
	clientRedis := redis.NewClient(&redis.Options{
		Addr: redisAddres,
		DB:   0,
	})

	return func(ctx *gin.Context) {
		targetTaskName := ctx.Param("task_name")
		ansStr, err := clientRedis.Get(targetTaskName).Result()
		if err == nil {
			res, err := strconv.ParseUint(ansStr, 10, 32)
			if err == nil {
				ctx.JSON(http.StatusOK, gin.H{"MinimalTime": res, "TaskName": targetTaskName})
				return
			}
		}

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
		err = clientRedis.Set(targetTaskName, res, 5*time.Second).Err()
		if err != nil {
			log.Printf("can't cash calculation due to %v", err)
		}
		log.Printf("Calculation result: %v", res)
		ctx.JSON(http.StatusOK, gin.H{"MinimalTime": res, "TaskName": targetTaskName})

	}
}
