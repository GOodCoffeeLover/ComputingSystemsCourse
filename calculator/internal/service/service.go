package service

import (
	"calculator/internal/core"
	pb "calculator/pkg/calculator_pb"
	"context"
	"fmt"
	"github.com/go-redis/redis"
	"log"
	"strconv"
	"time"
)

type taskGetter interface {
	Get(taskName string) (task core.Task, err error)
}

type Service struct {
	pb.UnimplementedCalculatorServer
	tasks       taskGetter
	clientRedis *redis.Client
}

func NewService(tGetter taskGetter, redis *redis.Client) (*Service, error) {

	return &Service{
		tasks:       tGetter,
		clientRedis: redis,
	}, nil
}

func (s *Service) Calculate(ctx context.Context, in *pb.CalculateRequest) (*pb.CalculateResponse, error) {
	targetTaskName := in.GetTask()
	ansStr, err := s.clientRedis.Get(targetTaskName).Result()
	if err == nil {
		res, _ := strconv.ParseUint(ansStr, 10, 32)
		return &pb.CalculateResponse{Time: res}, nil

	}

	task, err := s.tasks.Get(targetTaskName)

	if err != nil {
		return nil, fmt.Errorf("can't get task due to %v", err)
	}
	log.Println(task)
	res, err := task.StartCalculation()
	if err != nil {
		return nil, fmt.Errorf("can't complite calculations due to %v", err)
	}
	err = s.clientRedis.Set(targetTaskName, res, 5*time.Second).Err()
	if err != nil {
		log.Printf("can't cash calculation due to %v", err)
	}
	log.Printf("Calculation result: %v", res)
	return &pb.CalculateResponse{Time: res}, nil
}
