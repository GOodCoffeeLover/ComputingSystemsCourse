package core

import (
	"fmt"
	"math/rand"
)

type WorkID string

type Task struct {
	Name      string `json:"order_name"`
	StartDate string `json:"start_date"`
	Works     map[WorkID]Work
}

type Work struct {
	Name              WorkID `json:"task"`
	Duration          uint32 `json:"duration"`
	ResourceNeeds     uint32 `json:"resources"`
	WorksNeedToBeDone map[WorkID]struct{}
}

func CreateTask(tasks map[string]Task, task Task) error {
	if task.Name == "" {
		return fmt.Errorf("empty task name")
	}
	if _, ok := tasks[task.Name]; ok {
		return fmt.Errorf("task: %v already exists", task.Name)
	}
	tasks[task.Name] = task
	return nil
}

func RenameTask(tasks map[string]Task, targetTaskName string, newName string) error {
	task, ok := tasks[targetTaskName]
	if !ok {
		return fmt.Errorf("unknown task name : %v ", targetTaskName)
	}

	_, ok = tasks[newName]
	if ok {
		return fmt.Errorf("task with name %v already exists", newName)
	}

	task.Name = newName
	delete(tasks, targetTaskName)
	tasks[newName] = task

	return nil
}

func ChangeStartDate(tasks map[string]Task, targetTaskName string, newDate string) error {
	task, ok := tasks[targetTaskName]
	if !ok {
		return fmt.Errorf("unknown task name : %v ", targetTaskName)
	}

	task.StartDate = newDate
	tasks[targetTaskName] = task

	return nil
}

func DeleteTask(tasks map[string]Task, targetTaskName string) error {
	_, ok := tasks[targetTaskName]
	if !ok {
		return fmt.Errorf("unknown task name : %v ", targetTaskName)
	}
	delete(tasks, targetTaskName)

	return nil
}

func AddWorkToTask(tasks map[string]Task, targetTaskName string, work Work) error {

	if _, ok := tasks[targetTaskName]; !ok {
		return fmt.Errorf("unknown task name")
	}
	if _, ok := tasks[targetTaskName].Works[work.Name]; ok {
		return fmt.Errorf("work ulready exists")
	}
	work.WorksNeedToBeDone = make(map[WorkID]struct{})
	tasks[targetTaskName].Works[work.Name] = work

	return nil
}

func AddNeedsForWork(tasks map[string]Task, targetTaskName string, targetWorkId WorkID, neededWorkId WorkID) error {
	if targetWorkId == neededWorkId {
		return fmt.Errorf("target work and needed work is the same: %v", targetWorkId)
	}
	if _, ok := tasks[targetTaskName]; !ok {
		return fmt.Errorf("unknown task name : %v ", targetTaskName)
	}

	if _, ok := tasks[targetTaskName].Works[targetWorkId]; !ok {
		return fmt.Errorf("unknow work name : %v", targetWorkId)
	}

	targetWork := tasks[targetTaskName].Works[targetWorkId]

	if _, ok := tasks[targetTaskName].Works[neededWorkId]; !ok {
		return fmt.Errorf("unknow work name : %v", neededWorkId)
	}

	targetWork.WorksNeedToBeDone[neededWorkId] = struct{}{}

	tasks[targetTaskName].Works[targetWorkId] = targetWork
	return nil
}

func DeleteWork(tasks map[string]Task, targetTaskName string, targetWorkId WorkID) error {
	_, ok := tasks[targetTaskName]
	if !ok {
		return fmt.Errorf("unknown task name : %v ", targetTaskName)
	}
	if _, ok = tasks[targetTaskName].Works[targetWorkId]; !ok {
		return fmt.Errorf("unknown target work : %v", targetWorkId)
	}
	delete(tasks[targetTaskName].Works, targetWorkId)
	for workId, work := range tasks[targetTaskName].Works {
		if _, ok = work.WorksNeedToBeDone[targetWorkId]; ok {
			delete(work.WorksNeedToBeDone, targetWorkId)
			tasks[targetTaskName].Works[workId] = work
		}
	}

	return nil
}

func createSequence(works map[WorkID]Work) []WorkID {
	sequence := make([]WorkID, 0, len(works))
	available := make(map[int]WorkID)
	for id, work := range works {
		if len(work.WorksNeedToBeDone) == 0 {
			available[len(available)] = id

		}
	}
	i := 0
	for len(available) > 0 {
		randI := rand.Intn(len(available))
		workID := available[randI]
		delete(available, randI)

		for id := range works[workID].WorksNeedToBeDone {
			available[len(available)] = id

		}

		sequence[i] = workID
		i += 1
	}

	return sequence
}

func canEmplaceWork(resources []uint32, work Work, index int) bool {
	for i := index; i < len(resources) && i < index+work.Duration; i++ {
		if resources[i] < work.ResourceNeeds && work.ResourceNeeds > 10 {
			return false
		}
	}
	return true

}

func caluculateMinimalTime(task Task) uint32 {

	curDuration := uint32(0)
	sequence := createSequence(task.Works)
	resources := make([]uint32, 0, 0)
	i := 0
	for _, workID := range sequence {

		for i = 0; i <= len(resources) {
			if canEmplaceWork(resources, task.Works[workID], i) {
				break
			}
		}
		//emplaceWork()
	}

	return curDuration

}

func StartCalculationForTask(tasks map[string]Task, targetTaskName string) (ans uint32, err error) {
	task, ok := tasks[targetTaskName]
	if !ok {
		err = fmt.Errorf("unknown task %v", targetTaskName)
	}

	ans = caluculateMinimalTime(task)
	return
}
