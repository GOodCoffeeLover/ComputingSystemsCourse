package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"main/internal/core"
	"net/http"
	"os"
	"strings"
)

type tasksStorage interface {
	Get(string) (core.Task, error)
	Set(string, core.Task) error
	Delete(string) error
}

// get /task/:task_name
func HandleTaskAccess(tasks tasksStorage) func(c *gin.Context) {
	return func(context *gin.Context) {
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
func HandleTaskCreation(tasks tasksStorage) func(c *gin.Context) {

	return func(c *gin.Context) {
		task := core.Task{Works: make(map[core.WorkID]core.Work)}
		c.BindJSON(&task)
		task.Name = strings.ReplaceAll(task.Name, " ", "")

		err := core.CreateTask(tasks, task)
		if err != nil {
			c.Error(err)
			c.Data(http.StatusConflict, "application/json", []byte(fmt.Sprintf("error: %v\n", err)))
			return
		}
		c.JSON(http.StatusOK, fmt.Sprintf("task '%v' created", task.Name))
	}
}

// post /task/:task_name
func HandleTaskUpdate(tasks tasksStorage) func(c *gin.Context) {

	return func(c *gin.Context) {
		targetTaskName := c.Param("task_name")

		task := core.Task{}
		err := c.ShouldBindJSON(&task)
		task.Name = strings.ReplaceAll(task.Name, " ", "")

		if task.StartDate != "" {
			err = core.ChangeStartDate(tasks, targetTaskName, task.StartDate)
			if err != nil {
				c.Error(err)
				c.Data(http.StatusConflict, "application/json", []byte(fmt.Sprintf("error: %v\n", err)))
			}
		}

		if task.Name != "" {
			err = core.RenameTask(tasks, targetTaskName, task.Name)
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
func HandleTaskDelete(tasks tasksStorage) func(c *gin.Context) {

	return func(c *gin.Context) {
		targetTaskName := c.Param("task_name")

		err := core.DeleteTask(tasks, targetTaskName)
		if err != nil {
			c.Error(err)
			c.Data(http.StatusConflict, "application/json", []byte(fmt.Sprintf("error: %v\n", err)))
			return
		}

		c.JSON(http.StatusOK, fmt.Sprintf("task '%v' deleted", targetTaskName))
	}
}

// get /work/:task_name/:work_name
func HandleWorkAccess(tasks tasksStorage) func(c *gin.Context) {
	return func(context *gin.Context) {
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
func HandleWorkCreation(tasks tasksStorage) func(c *gin.Context) {
	return func(context *gin.Context) {
		taskName := context.Param("task_name")
		work := core.Work{}
		context.BindJSON(&work)

		err := core.AddWorkToTask(tasks, taskName, work)

		if err != nil {
			context.JSON(http.StatusConflict, fmt.Sprintf("error: %v", err))
			return
		}
		context.JSON(http.StatusOK, fmt.Sprintf("work: %v", work))

	}
}

// post /work/:task_name/:work_name

func HandleWorkNeedsSetup(tasks tasksStorage) func(c *gin.Context) {
	type Needs struct {
		Work []string `json:"pred"`
	}
	return func(context *gin.Context) {
		taskName := context.Param("task_name")
		workName := core.WorkID(context.Param("work_name"))

		needs := Needs{}
		context.Bind(&needs)

		for _, work := range needs.Work {
			err := core.AddNeedsForWork(tasks, taskName, workName, core.WorkID(work))

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
func HandleWorkDelete(tasks tasksStorage) func(c *gin.Context) {

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
func HandleCalculation(tasks tasksStorage) func(c *gin.Context) {
	calcServiceAddr := os.Getenv("CALCULATION_SERVICE_ADDRESS")
	if calcServiceAddr == "" {
		log.Fatal("don't have calculation service address")
	} else {
		log.Printf("Calculation Address: %v", calcServiceAddr)
	}
	return func(context *gin.Context) {
		targetTaskName := context.Param("task_name")

		task, err := tasks.Get(targetTaskName)
		fmt.Println(task)
		if err != nil {
			err = fmt.Errorf("unknown task %v", targetTaskName)
			context.Error(err)
			context.Data(http.StatusConflict, "application/json", []byte(fmt.Sprintf("error: %v", err)))
			return
		}

		ans, err := getCalculation(calcServiceAddr, task)
		if err != nil {
			context.Error(err)
			context.Data(http.StatusConflict, "application/json", []byte(fmt.Sprintf("error: %v", err)))
			return
		}

		if len(context.Errors) == 0 {
			context.JSON(http.StatusOK, gin.H{"Task": targetTaskName, "MinimalTime": ans})
		}

	}
}

func getCalculation(calculationServerAddress string, task core.Task) (uint32, error) {
	calculationServerAddress = "http://" + calculationServerAddress + "/calculate"
	jsonDoc, err := json.Marshal(task)
	if err != nil {
		err = fmt.Errorf("cant marshal task due to %v", err)
		return 0, err
	}

	req, err := http.NewRequest(http.MethodGet, calculationServerAddress, bytes.NewReader(jsonDoc))
	if err != nil {
		return 0, err
	}
	rep, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, nil
	}
	repBody, err := ioutil.ReadAll(rep.Body)
	if err != nil {
		return 0, err
	}
	var ans struct {
		Answer uint32
	}
	if err = json.Unmarshal(repBody, &ans); err != nil {
		return 0, err
	}
	return ans.Answer, nil
}
