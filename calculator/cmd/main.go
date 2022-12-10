package main

import (
	"calculator/internal/handlers"
	"calculator/internal/storage"
	"github.com/gin-gonic/gin"
	pb "github.com/student31415/ComputingSystemsCourse/calculator"
	"log"
	"net/http"
)

type server struct {
	pb.UnimplementedCalculatorServer
}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("Received: %v", in.GetName())
	return &pb.HelloReply{Message: "Hello " + in.GetName()}, nil
}

func main() {
	router := gin.Default()
	tasksDB, err := storage.NewTasksMongoStorage()
	if err != nil {
		log.Fatalln(err)
	}

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "it's okay"})
	})

	router.GET("/calculate/:task_name", handlers.HandleTaskCalculations(tasksDB))
	err = router.Run(":8090")
	if err != nil {
		log.Fatalln(err)
	}

}
