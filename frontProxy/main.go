package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/segmentio/kafka-go"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"strconv"
	"strings"
)

func main() {
	router := gin.Default()
	envs, err := getEnvs([]string{
		"MAIN_SERVICE_ADDRESS",
		"CALCULATOR_SERVICE_ADDRESS",
		"KAFKA_BROKERS",
	})
	if err != nil {
		log.Fatalln(err)
	}
	kafkaWriter, err := setupKafka("logs", strings.Split(envs["KAFKA_BROKERS"], ","))

	router.GET("/task/:task_name", Proxy(envs["MAIN_SERVICE_ADDRESS"], kafkaWriter))
	router.POST("/task", Proxy(envs["MAIN_SERVICE_ADDRESS"], kafkaWriter))
	router.POST("/task/:task_name", Proxy(envs["MAIN_SERVICE_ADDRESS"], kafkaWriter))
	router.DELETE("/task/:task_name", Proxy(envs["MAIN_SERVICE_ADDRESS"], kafkaWriter))

	router.GET("/work/:task_name/:work_name", Proxy(envs["MAIN_SERVICE_ADDRESS"], kafkaWriter))
	router.POST("/work/:task_name/:work_name", Proxy(envs["MAIN_SERVICE_ADDRESS"], kafkaWriter))
	router.POST("/work/:task_name", Proxy(envs["MAIN_SERVICE_ADDRESS"], kafkaWriter))
	router.DELETE("/work/:task_name/:work_name", Proxy(envs["MAIN_SERVICE_ADDRESS"], kafkaWriter))

	router.GET("/calculate/:task_name", Proxy(envs["CALCULATOR_SERVICE_ADDRESS"], kafkaWriter))

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
			req.URL.Scheme = "http"
			req.URL.Host = target
			req.Host = target
			err := w.WriteMessages(context.Background(), kafka.Message{
				Value:     []byte(req.URL.Path),
				Partition: i % 2,
			})
			i += 1
			if err != nil {
				log.Printf("can't send message(%v) to kafka due to:%v", req.URL.Path, err)
			}
		}

		proxy := &httputil.ReverseProxy{Director: director}
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}
