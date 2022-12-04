package main

import (
	"context"
	kafka "github.com/segmentio/kafka-go"
	"log"
	"os"
	"strings"
	"sync"
)

func main() {
	brokers := os.Getenv("KAFKA_BROKERS")
	topic := os.Getenv("KAFKA_TOPIC")
	if brokers == "" && topic == "" {
		log.Fatalf("can't start consumer due to empty envs KAFKA_BROKERS(%v) and KAFKA_TOPIC(%v)", brokers, topic)
	}
	numPartitions := 2
	wg := sync.WaitGroup{}
	wg.Add(numPartitions)
	for i := 0; i < numPartitions; i++ {
		go func() {
			readLogs(strings.Split(brokers, ","), topic, i)
			wg.Done()
		}()

		log.Printf("start consumer for %v partition", i)
	}
	wg.Wait()

}

func readLogs(brokers []string, topic string, partition int) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Partition: partition,
		Topic:     topic,
		Brokers:   brokers,
	})
	for {
		msg, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Fatalf("can't read message from kafka due to %v", err)
		}
		log.Printf("Log: %v", string(msg.Value))
	}
}
