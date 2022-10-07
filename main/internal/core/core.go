package core

import (
	"fmt"
)

type WorkID string

type Task struct {
	Name      string          `json:"task_name" bson:"_id"`
	StartDate string          `json:"start_date" bson:"start_date"`
	Works     map[WorkID]Work `json:"works" bson:"works"`
}

type Work struct {
	Name              WorkID              `json:"work_name" bson:"work_name"`
	Duration          uint                `json:"duration" bson:"duration"`
	ResourceNeeds     uint                `json:"resources" bson:"resource_needs"`
	WorksNeedToBeDone map[WorkID]struct{} `json:"works_need_to_be_done" bson:"works_need_to_be_done"`
}

type tasksStorage interface {
	Get(string) (Task, error)
	Set(string, Task) error
	Delete(string) error
}

func CreateTask(tasks tasksStorage, task Task) error {
	if task.Name == "" {
		return fmt.Errorf("empty task name")
	}
	if _, err := tasks.Get(task.Name); err == nil {
		return fmt.Errorf("task: %v already exists", task.Name)
	}
	tasks.Set(task.Name, task)
	return nil
}

func RenameTask(tasks tasksStorage, targetTaskName string, newName string) error {
	task, err := tasks.Get(targetTaskName)
	if err != nil {
		return fmt.Errorf("unknown task name : %v ", targetTaskName)
	}

	_, err = tasks.Get(newName)
	if err == nil {
		return fmt.Errorf("task with name %v already exists", newName)
	}

	task.Name = newName
	if err = tasks.Delete(targetTaskName); err != nil {
		return err
	}
	tasks.Set(newName, task)

	return nil
}

func ChangeStartDate(tasks tasksStorage, targetTaskName string, newDate string) error {
	task, err := tasks.Get(targetTaskName)
	if err != nil {
		return fmt.Errorf("unknown task name : %v ", targetTaskName)
	}

	task.StartDate = newDate
	tasks.Set(targetTaskName, task)

	return nil
}

func DeleteTask(tasks tasksStorage, targetTaskName string) error {
	_, err := tasks.Get(targetTaskName)
	if err != nil {
		return fmt.Errorf("unknown task name : %v ", targetTaskName)
	}

	return tasks.Delete(targetTaskName)
}

func AddWorkToTask(tasks tasksStorage, targetTaskName string, work Work) error {
	task, err := tasks.Get(targetTaskName)
	if err != nil {
		return err
	}
	if _, ok := task.Works[work.Name]; ok {
		return fmt.Errorf("work ulready exists")
	}
	work.WorksNeedToBeDone = make(map[WorkID]struct{})
	task.Works[work.Name] = work

	return tasks.Set(targetTaskName, task)
}

func AddNeedsForWork(tasks tasksStorage, targetTaskName string, targetWorkId WorkID, neededWorkId WorkID) error {
	if targetWorkId == neededWorkId {
		return fmt.Errorf("target work and needed work is the same: %v", targetWorkId)
	}
	task, err := tasks.Get(targetTaskName)
	if err != nil {
		return fmt.Errorf("unknown task name : %v ", targetTaskName)
	}

	if _, ok := task.Works[targetWorkId]; !ok {
		return fmt.Errorf("unknow work name : %v", targetWorkId)
	}

	targetWork := task.Works[targetWorkId]

	if _, ok := task.Works[neededWorkId]; !ok {
		return fmt.Errorf("unknow work name : %v", neededWorkId)
	}

	targetWork.WorksNeedToBeDone[neededWorkId] = struct{}{}

	task.Works[targetWorkId] = targetWork

	return tasks.Set(targetTaskName, task)
}

func DeleteWork(tasks tasksStorage, targetTaskName string, targetWorkId WorkID) error {
	task, err := tasks.Get(targetTaskName)
	if err != nil {
		return fmt.Errorf("unknown task name : %v ", targetTaskName)
	}
	if _, ok := task.Works[targetWorkId]; !ok {
		return fmt.Errorf("unknown target work : %v", targetWorkId)
	}
	delete(task.Works, targetWorkId)
	for workId, work := range task.Works {
		if _, ok := work.WorksNeedToBeDone[targetWorkId]; ok {
			delete(work.WorksNeedToBeDone, targetWorkId)
			task.Works[workId] = work
		}
	}

	return tasks.Set(targetTaskName, task)
}
