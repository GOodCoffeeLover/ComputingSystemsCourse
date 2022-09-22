package core

import (
	"fmt"
	"math/rand"
	"time"
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

func needsRequired(finished map[WorkID]struct{}, work Work) bool {
	for workName, _ := range work.WorksNeedToBeDone {
		if _, ok := finished[workName]; !ok {
			return false
		}
	}
	return true

}
func getRandomWork(unFinished map[WorkID]struct{}) WorkID {
	keys := make([]WorkID, len(unFinished), len(unFinished))
	i := 0
	for k := range unFinished {
		keys[i] = k
		i++
	}

	return keys[rand.Intn(len(keys)-1)]

}

type step struct {
	resource    uint32
	finishesNow map[WorkID]struct{}
}

func emplaceWork(steps []step, work Work) []step {
	for i := 0; work.Duration > 0; i++ {
		if i >= len(steps) {
			steps = append(steps, step{resource: 10, finishesNow: map[WorkID]struct{}{}})
		}

		if steps[i].resource-work.ResourceNeeds > 0 {
			steps[i].resource -= work.ResourceNeeds
			work.Duration -= 1
		}
		if work.Duration == 0 {
			steps[i].finishesNow[work.Name] = struct{}{}
		}
	}
	return steps
}

func makeStep(steps []step, finished map[WorkID]struct{}) []step {
	for workID := range steps[0].finishesNow {
		finished[workID] = struct{}{}
	}
	steps = steps[1:len(steps)]
	return steps

}

func caluculateMinimalTime(task Task, ch chan<- uint32) {
	rand.Seed(time.Now().UnixNano())

	unFinished := map[WorkID]struct{}{}
	for workId := range task.Works {
		unFinished[workId] = struct{}{}
	}

	finished := map[WorkID]struct{}{}
	steps := []step{}
	curDuration := uint32(0)
	for i := 0; i < 1000*1000 && len(unFinished) > 0; i++ {
		curWorkName := getRandomWork(unFinished)
		steps = emplaceWork(steps, task.Works[curWorkName])
		steps = makeStep(steps, finished)
		curDuration += 1
	}
	ch <- curDuration

}

func StartCalculationForTask(tasks map[string]Task, targetTaskName string) (ans uint32, err error) {
	task, ok := tasks[targetTaskName]
	if !ok {
		return ans, fmt.Errorf("unknown task %v", targetTaskName)
	}
	ch := make(chan uint32)
	caluculateMinimalTime(task, ch)
	ans = <-ch
	return
}
