package core

import (
	"math"
	"math/rand"
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
	counter := len(works)
	for len(available) > 0 && counter > 0 {

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

func (task *Task) calculateMinimalTime() uint {

	sequence := createSequence(task.Works)
	resources := make([]uint, 0, 0)
	finishedData := make(map[WorkID]int)
	for _, workID := range sequence {
		i := 0
		work := task.Works[workID]
		for workId, _ := range work.WorksNeedToBeDone {
			if i < finishedData[workId] {
				i = finishedData[workId]
			}
		}
		for ; i <= len(resources); i++ {
			if canEmplaceWork(resources, work, i) {
				break
			}
		}
		resources = emplaceWork(resources, work, i)
		finishedData[workID] = i + int(work.Duration)
	}
	//fmt.Println(resources)
	return uint(len(resources))

}

func (task *Task) StartCalculation() (ans uint, err error) {
	maxGorutines := 10

	gather := make(chan uint, maxGorutines)
	stopper := make(chan struct{}, maxGorutines)
	result := make(chan float64)
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
		result <- min
	}()

	for i := 0; i < numOfIterations; i++ {
		stopper <- struct{}{}
		go func() {
			gather <- task.calculateMinimalTime()
			<-stopper
		}()
	}

	return uint(<-result), nil
}
