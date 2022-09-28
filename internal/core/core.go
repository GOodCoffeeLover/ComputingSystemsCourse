package core

import (
	"fmt"
	"math"
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
		sequence[i] = workID
		i += 1
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

func (task *Task) caluculateMinimalTime() uint {

	sequence := createSequence(task.Works)
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

func StartCalculationForTask(tasks tasksStorage, targetTaskName string) (ans uint, err error) {
	task, err := tasks.Get(targetTaskName)
	if err != nil {
		err = fmt.Errorf("unknown task %v", targetTaskName)
	}
	maxGorutines := 10

	gather := make(chan uint, maxGorutines)
	//stopper := make(chan struct{}, maxGorutines)
	res := make(chan float64)
	numOfIterations := 1000 * 1000

	go func() {
		var min float64
		for i := 0; i < numOfIterations; i++ {
			res := float64(<-gather)
			if i != 0 {
				min = math.Min(min, res)
			} else {
				min = res
			}
		}
		res <- min
	}()

	for i := 0; i < numOfIterations; i++ {
		//stopper <- struct{}{}
		go func() {
			gather <- task.caluculateMinimalTime()
			//<-stopper
		}()
	}

	return uint(<-res), nil
}
