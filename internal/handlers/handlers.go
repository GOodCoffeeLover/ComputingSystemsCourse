package handlers

import (
	"ComputingSystemsCourse/internal/core"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"sync"
)

type tasksStorage interface {
	Get(string) (core.Task, error)
	Set(string, core.Task) error
	Delete(string) error
}

// get /task/:task_name
func HandleTaskAccess(tasks tasksStorage, mutex sync.Mutex) func(c *gin.Context) {
	return func(context *gin.Context) {
		mutex.Lock()
		defer mutex.Unlock()
		taskName := context.Param("task_name")
		task, err := tasks.Get(taskName)
		if err != nil {
			context.JSON(404, fmt.Sprintf("unknown task %v", taskName))
			return
		}
		context.JSON(http.StatusOK, fmt.Sprintf("task: %v", task))
	}
}

// post /task  json:{"order_name":"", "start_date":""}
func HandleTaskCreation(tasks tasksStorage, mutex sync.Mutex) func(c *gin.Context) {

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
func HandleTaskUpdate(tasks tasksStorage, mutex sync.Mutex) func(c *gin.Context) {

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
func HandleTaskDelete(tasks tasksStorage, mutex sync.Mutex) func(c *gin.Context) {

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
func HandleWorkAccess(tasks tasksStorage, mutex sync.Mutex) func(c *gin.Context) {
	return func(context *gin.Context) {
		mutex.Lock()
		defer mutex.Unlock()
		taskName := context.Param("task_name")
		task, err := tasks.Get(taskName)
		if err != nil {
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
func HandleWorkCreation(tasks tasksStorage, mutex sync.Mutex) func(c *gin.Context) {
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

func HandleWorkNeedsSetup(tasks tasksStorage, mutex sync.Mutex) func(c *gin.Context) {
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
			context.JSON(http.StatusOK, fmt.Sprintf("needs %v for work %v  succesfuly added", needs.Work, workName))
		}

	}
}

// delte /work/:task_name/:work_name
func HandleWorkDelete(tasks tasksStorage, mutex sync.Mutex) func(c *gin.Context) {

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
func HandleCalculation(tasks tasksStorage, mutex sync.Mutex) func(c *gin.Context) {

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
			context.JSON(http.StatusOK, fmt.Sprintf("minimal duration for execution task %v  is %v", taskName, ans))
		}

	}
}
