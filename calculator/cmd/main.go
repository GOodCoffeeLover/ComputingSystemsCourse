package main

import (
	"calculator/internal/service"
	"calculator/internal/storage"
	pb "calculator/pkg/calculator_pb"
	"github.com/go-redis/redis"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
)

func main() {
	lis, err := net.Listen("tcp", ":8090")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	redisAddres := os.Getenv("REDIS_ADDRESS")
	if redisAddres == "" {
		log.Fatalln("don't have redis address")
	}
	clientRedis := redis.NewClient(&redis.Options{
		Addr: redisAddres,
		DB:   0,
	})

	tasks, err := storage.NewTasksMongoStorage()
	if err != nil {
		log.Fatalln(err)
	}
	defer tasks.Disconnect()
	s := grpc.NewServer()
	service, err := service.NewService(tasks, clientRedis)

	pb.RegisterCalculatorServer(s, service)
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
