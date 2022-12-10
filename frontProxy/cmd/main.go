package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/segmentio/kafka-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	pb "main/pkg/calculator_pb"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	router := gin.Default()
	envs, err := getEnvs([]string{
		"MAIN_SERVICE_ADDRESS",
		"CALCULATOR_SERVICE_ADDRESS",
		"KAFKA_BROKERS",
		"KAFKA_TOPIC",
	})
	if err != nil {
		log.Fatalln(err)
	}
	kafkaWriter, err := setupKafka(envs["KAFKA_TOPIC"], strings.Split(envs["KAFKA_BROKERS"], ","))

	router.GET("/task/:task_name", Proxy(envs["MAIN_SERVICE_ADDRESS"], kafkaWriter))
	router.POST("/task", Proxy(envs["MAIN_SERVICE_ADDRESS"], kafkaWriter))
	router.POST("/task/:task_name", Proxy(envs["MAIN_SERVICE_ADDRESS"], kafkaWriter))
	router.DELETE("/task/:task_name", Proxy(envs["MAIN_SERVICE_ADDRESS"], kafkaWriter))

	router.GET("/work/:task_name/:work_name", Proxy(envs["MAIN_SERVICE_ADDRESS"], kafkaWriter))
	router.POST("/work/:task_name/:work_name", Proxy(envs["MAIN_SERVICE_ADDRESS"], kafkaWriter))
	router.POST("/work/:task_name", Proxy(envs["MAIN_SERVICE_ADDRESS"], kafkaWriter))
	router.DELETE("/work/:task_name/:work_name", Proxy(envs["MAIN_SERVICE_ADDRESS"], kafkaWriter))
	conn, err := grpc.Dial(envs["CALCULATOR_SERVICE_ADDRESS"]+"8090", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	calc := pb.NewCalculatorClient(conn)

	// Contact the server and print out its response.

	router.GET("/calculate/:task_name", func(ctx *gin.Context) {
		task := ctx.Param("task_name")
		ctx_rpc, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		r, err := calc.Calculate(ctx_rpc, &pb.CalculateRequest{Task: task})
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("could not greet: %v", err))
		}
		ctx.JSON(http.StatusOK, gin.H{"task": task, "MinimalTime": r.GetTime()})

	})

	err = router.Run(":8080")
	if err != nil {
		log.Fatalln(err)
	}
}
func getEnvs(names []string) (map[string]string, error) {
	m := map[string]string{}
	for _, name := range names {
		v := os.Getenv(name)
		if v == "" {
			return nil, fmt.Errorf("can't get env: %v", name)
		}
		m[name] = v
	}
	return m, nil
}
func setupKafka(topic string, brokers []string) (*kafka.Writer, error) {
	for _, broker := range brokers {
		if err := createTopic(topic, broker); err != nil {
			return nil, err
		}
	}

	writer := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}
	return writer, nil
}

func createTopic(topic, broker string) error {
	conn, err := kafka.Dial("tcp", broker)
	if err != nil {
		return fmt.Errorf("can't connect to kafka due to: %v", err)
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return fmt.Errorf("can't connect to kafka controller due to: %v", err)
	}
	var controllerConn *kafka.Conn
	controllerConn, err = kafka.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		return fmt.Errorf("can't connect to kafka controller due to: %v", err)
	}
	defer controllerConn.Close()

	topicConfigs := kafka.TopicConfig{
		Topic:             topic,
		NumPartitions:     2,
		ReplicationFactor: 1,
	}

	err = controllerConn.CreateTopics(topicConfigs)
	if err != nil {
		return fmt.Errorf("can't create topic %v in kafka(%v) due to: %v", topic, broker, err)
	}
	return nil
}

func Proxy(target string, w *kafka.Writer) gin.HandlerFunc {
	i := 0
	return func(c *gin.Context) {
		director := func(req *http.Request) {
			go func() {
				err := w.WriteMessages(context.Background(), kafka.Message{
					Value:     []byte(req.URL.Path),
					Partition: i % 2,
				})
				i += 1
				if err != nil {
					log.Printf("can't send message(%v) to kafka due to:%v", req.URL.Path, err)
				}
				log.Println("Send message to kafka broker")
			}()
			req.URL.Scheme = "http"
			req.URL.Host = target
			req.Host = target
		}

		proxy := &httputil.ReverseProxy{Director: director}
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}
