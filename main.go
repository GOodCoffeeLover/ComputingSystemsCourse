package main

import (
	"ComputingSystemsCourse/core"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strings"
	"sync"
)

// get /task/:task_name
func HandleTaskAccess(tasks map[string]core.Task, mutex sync.Mutex) func(c *gin.Context) {
	return func(context *gin.Context) {
		mutex.Lock()
		defer mutex.Unlock()
		taskName := context.Param("task_name")
		task, ok := tasks[taskName]
		if !ok {
			context.JSON(404, fmt.Sprintf("unknown task %v", taskName))
			return
		}
		context.JSON(http.StatusOK, fmt.Sprintf("task: %v", task))
	}
}

// post /task  json:{"order_name":"", "start_date":""}
func HandleTaskCreation(tasks map[string]core.Task, mutex sync.Mutex) func(c *gin.Context) {

	return func(c *gin.Context) {
		task := core.Task{Works: make(map[core.WorkID]core.Work)}
		c.BindJSON(&task)
		task.Name = strings.ReplaceAll(task.Name, " ", "")

		mutex.Lock()
		err := core.CreateTask(tasks, task)
		mutex.Unlock()
		if err != nil {
			c.Error(err)
			c.Data(http.StatusConflict, "application/json", []byte(fmt.Sprintf("error: %v\n", err)))
			return
		}
		c.JSON(http.StatusOK, fmt.Sprintf("task '%v' created", task.Name))
	}
}

// post /task/:task_name
func HandleTaskUpdate(tasks map[string]core.Task, mutex sync.Mutex) func(c *gin.Context) {

	return func(c *gin.Context) {
		targetTaskName := c.Param("task_name")

		task := core.Task{}
		err := c.ShouldBindJSON(&task)
		task.Name = strings.ReplaceAll(task.Name, " ", "")

		if task.StartDate != "" {
			mutex.Lock()
			err = core.ChangeStartDate(tasks, targetTaskName, task.StartDate)
			mutex.Unlock()
			if err != nil {
				c.Error(err)
				c.Data(http.StatusConflict, "application/json", []byte(fmt.Sprintf("error: %v\n", err)))
			}
		}

		if task.Name != "" {
			mutex.Lock()
			err = core.RenameTask(tasks, targetTaskName, task.Name)
			mutex.Unlock()
			if err != nil {
				c.Error(err)
				c.Data(http.StatusConflict, "application/json", []byte(fmt.Sprintf("error: %v\n", err)))
			}
		}
		if len(c.Errors) != 0 {
			return
		}
		c.JSON(http.StatusOK, fmt.Sprintf("task '%v' updated", targetTaskName))
	}
}

// delete /task/:task_name
func HandleTaskDelete(tasks map[string]core.Task, mutex sync.Mutex) func(c *gin.Context) {

	return func(c *gin.Context) {
		targetTaskName := c.Param("task_name")

		mutex.Lock()
		err := core.DeleteTask(tasks, targetTaskName)
		mutex.Unlock()
		if err != nil {
			c.Error(err)
			c.Data(http.StatusConflict, "application/json", []byte(fmt.Sprintf("error: %v\n", err)))
			return
		}

		c.JSON(http.StatusOK, fmt.Sprintf("task '%v' deleted", targetTaskName))
	}
}

// get /work/:task_name/:work_name
func HandleWorkAccess(tasks map[string]core.Task, mutex sync.Mutex) func(c *gin.Context) {
	return func(context *gin.Context) {
		mutex.Lock()
		defer mutex.Unlock()
		taskName := context.Param("task_name")
		task, ok := tasks[taskName]
		if !ok {
			context.JSON(404, fmt.Sprintf("unknown task %v", taskName))
			return
		}
		workName := core.WorkID(context.Param("work_name"))
		work, ok := task.Works[workName]
		if !ok {
			context.JSON(404, fmt.Sprintf("unknown work %v", taskName))
			return
		}
		context.JSON(http.StatusOK, fmt.Sprintf("work: %v", work))
	}
}

//post /work/:task_name json:{"task":"", "duration":0, "resources":0}
func HandleWorkCreation(tasks map[string]core.Task, mutex sync.Mutex) func(c *gin.Context) {
	return func(context *gin.Context) {
		taskName := context.Param("task_name")
		work := core.Work{}
		context.BindJSON(&work)

		mutex.Lock()
		defer mutex.Unlock()
		err := core.AddWorkToTask(tasks, taskName, work)

		if err != nil {
			context.JSON(http.StatusConflict, fmt.Sprintf("error: %v", err))
			return
		}
		context.JSON(http.StatusOK, fmt.Sprintf("work: %v", work))

	}
}

// post /work/:task_name/:work_name

func HandleWorkNeedsSetup(tasks map[string]core.Task, mutex sync.Mutex) func(c *gin.Context) {
	type Needs struct {
		Work []string `json:"pred"`
	}
	return func(context *gin.Context) {
		taskName := context.Param("task_name")
		workName := core.WorkID(context.Param("work_name"))

		needs := Needs{}
		context.Bind(&needs)

		for _, work := range needs.Work {
			mutex.Lock()
			err := core.AddNeedsForWork(tasks, taskName, workName, core.WorkID(work))
			mutex.Unlock()

			if err != nil {
				context.Error(err)
				context.Data(http.StatusConflict, "application/json", []byte(fmt.Sprintf("error: %v", err)))
			}

		}
		if len(context.Errors) == 0 {
			context.JSON(http.StatusOK, fmt.Sprintf("needs %v for work succesfuly added", needs.Work))
		}

	}
}

// delte /work/:task_name/:work_name
func HandleWorkDelete(tasks map[string]core.Task, mutex sync.Mutex) func(c *gin.Context) {

	return func(context *gin.Context) {
		taskName := context.Param("task_name")
		workName := core.WorkID(context.Param("work_name"))

		err := core.DeleteWork(tasks, taskName, workName)
		if err != nil {
			context.Error(err)
			context.Data(http.StatusConflict, "application/json", []byte(fmt.Sprintf("error: %v", err)))
		}

		if len(context.Errors) == 0 {
			context.JSON(http.StatusOK, fmt.Sprintf("work %v from task %v successfuly deleted", taskName, workName))
		}

	}
}

//get  /task/calculate/:task_name
func HandleCalculation(tasks map[string]core.Task, mutex sync.Mutex) func(c *gin.Context) {

	return func(context *gin.Context) {
		taskName := context.Param("task_name")
		mutex.Lock()
		ans, err := core.StartCalculationForTask(tasks, taskName)
		mutex.Unlock()
		if err != nil {
			context.Error(err)
			context.Data(http.StatusConflict, "application/json", []byte(fmt.Sprintf("error: %v", err)))
		}

		if len(context.Errors) == 0 {
			context.JSON(http.StatusOK, fmt.Sprintf("minimal time for execution task %v  is %v", taskName, ans))
		}

	}
}

func main() {
	tasks := make(map[string]core.Task)
	mutex := sync.Mutex{}
	router := gin.Default()

	router.GET("/check", func(context *gin.Context) {
		context.JSON(http.StatusOK, gin.H{"message": "It's okay"})
	})

	router.GET("/task/:task_name", HandleTaskAccess(tasks, mutex))
	router.GET("/task/calculate/:task_name", HandleCalculation(tasks, mutex))
	router.POST("/task", HandleTaskCreation(tasks, mutex))
	router.POST("/task/:task_name", HandleTaskUpdate(tasks, mutex))
	router.DELETE("/task/:task_name", HandleTaskDelete(tasks, mutex))

	router.GET("/work/:task_name/:work_name", HandleWorkAccess(tasks, mutex))
	router.POST("/work/:task_name/:work_name", HandleWorkNeedsSetup(tasks, mutex))
	router.POST("/work/:task_name", HandleWorkCreation(tasks, mutex))
	router.DELETE("/work/:task_name/:work_name", HandleWorkDelete(tasks, mutex))

	err := router.Run(":8080")
	if err != nil {
		log.Fatalln(err)
	}

}
