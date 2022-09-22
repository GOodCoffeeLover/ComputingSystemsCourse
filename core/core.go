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
	Duration          uint   `json:"duration"`
	ResourceNeeds     uint   `json:"resources"`
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

func isAvailable(done map[WorkID]struct{}, work Work) bool {
	for workID := range work.WorksNeedToBeDone {
		if _, ok := done[workID]; !ok {
			return false
		}
	}
	return true
}

func createSequence(works map[WorkID]Work) []WorkID {
	sequence := make([]WorkID, len(works), len(works))
	notAvailable := make(map[WorkID]struct{})
	available := make(map[int]WorkID)
	done := make(map[WorkID]struct{})

	for id, work := range works {
		if len(work.WorksNeedToBeDone) == 0 {
			available[len(available)] = id
		} else {
			notAvailable[id] = struct{}{}
		}
	}
	fmt.Println(available)
	i := 0
	for len(available) > 0 {

		randI := rand.Intn(len(available))
		available[randI], available[len(available)-1] = available[len(available)-1], available[randI]
		workID := available[len(available)-1]

		delete(available, len(available)-1)

		done[workID] = struct{}{}

		for id, _ := range notAvailable {
			if isAvailable(done, works[id]) {
				available[len(available)] = id
				delete(notAvailable, id)
			}
		}

		fmt.Println(available)
		sequence[i] = workID
		i += 1
		fmt.Println(sequence)
	}

	return sequence
}

func canEmplaceWork(resources []uint, work Work, index int) bool {
	for i := index; i < len(resources) && i < index+int(work.Duration); i++ {
		if resources[i]+work.ResourceNeeds > 10 {
			return false
		}
	}
	return true

}
func emplaceWork(resources []uint, work Work, start int) []uint {
	if len(resources) < start+int(work.Duration) {
		need := start + int(work.Duration) - len(resources)
		buff := make([]uint, need, need)
		resources = append(resources, buff...)
	}
	for i := start; i < start+int(work.Duration); i++ {
		resources[i] += work.ResourceNeeds
	}
	return resources
}

func caluculateMinimalTime(task Task) uint {

	sequence := createSequence(task.Works)
	fmt.Println(sequence)
	resources := make([]uint, 0, 0)
	i := 0
	for _, workID := range sequence {
		//fmt.Println(workID)
		//fmt.Println(resources)
		for i = 0; i <= len(resources); i++ {
			if canEmplaceWork(resources, task.Works[workID], i) {
				break
			}
		}
		//fmt.Println(i)
		resources = emplaceWork(resources, task.Works[workID], i)
	}
	//fmt.Println(resources)

	return uint(len(resources))

}

func StartCalculationForTask(tasks map[string]Task, targetTaskName string) (ans uint, err error) {
	task, ok := tasks[targetTaskName]
	if !ok {
		err = fmt.Errorf("unknown task %v", targetTaskName)
	}
	//ch := make(chan uint, 10)
	//for i := 0; i< 1000*1000
	ans = caluculateMinimalTime(task)
	return
}
