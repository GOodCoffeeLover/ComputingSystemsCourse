package storage

import (
	"ComputingSystemsCourse/internal/core"
	"fmt"
)

type TasksMapStorage struct {
	tasks map[string]core.Task
}

func NewTaskMapStorage() *TasksMapStorage {
	tms := TasksMapStorage{tasks: make(map[string]core.Task)}
	return &tms
}

func (tms *TasksMapStorage) Get(taskName string) (core.Task, error) {
	task, ok := tms.tasks[taskName]
	fmt.Println(">>>", task, ok)
	if ok {
		return task, nil
	} else {
		return task, fmt.Errorf("unknown task name %v", taskName)
	}
}

func (tms *TasksMapStorage) Set(taskName string, task core.Task) error {
	tms.tasks[taskName] = task
	return nil
}

func (tms *TasksMapStorage) Delete(taskName string) error {
	delete(tms.tasks, taskName)
	return nil
}
