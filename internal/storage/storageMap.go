package storage

import (
	"ComputingSystemsCourse/internal/core"
	"fmt"
	"sync"
)

type TasksMapStorage struct {
	mutex sync.RWMutex
	tasks map[string]core.Task
}

func NewTaskMapStorage() *TasksMapStorage {
	return &TasksMapStorage{
		tasks: make(map[string]core.Task),
		mutex: sync.RWMutex{},
	}
}

func (tms *TasksMapStorage) Get(taskName string) (core.Task, error) {
	tms.mutex.RLock()
	defer tms.mutex.RUnlock()

	task, ok := tms.tasks[taskName]
	if ok {
		return task, nil
	} else {
		return task, fmt.Errorf("unknown task name %v", taskName)
	}
}

func (tms *TasksMapStorage) Set(taskName string, task core.Task) error {
	tms.mutex.Lock()
	defer tms.mutex.Unlock()

	tms.tasks[taskName] = task
	return nil
}

func (tms *TasksMapStorage) Delete(taskName string) error {
	tms.mutex.Lock()
	defer tms.mutex.Unlock()

	delete(tms.tasks, taskName)
	return nil
}
