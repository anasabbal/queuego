package main

import (
	"flag"
	"log"
	"queuego/internal/protocol"
	"queuego/pkg/client"
	"time"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:9092", "broker address")
	topic := flag.String("topic", "test", "topic name")
	message := flag.String("message", "hello", "message payload")
	count := flag.Int("count", 1, "number of messages")
	flag.Parse()

	cfg := client.ClientConfig{
		RetryMax:      5,
		RetryInterval: time.Second,
		ConnTimeout:   5 * time.Second,
	}

	producer := client.NewProducer(cfg)

	if err := producer.Connect(*addr); err != nil {
		log.Fatal(err)
	}
	defer producer.Disconnect()

	for i := 0; i < *count; i++ {
		if err := producer.Publish(*topic, []byte(*message)); err != nil {
			log.Fatal(err)
		}

		// Wait for ACK
		resp, err := producer.ReadResponse()
		if err != nil {
			log.Fatal("failed to read ACK:", err)
		}
		if resp.Type != protocol.ACK {
			log.Fatalf("expected ACK but got %s", resp.Type)
		}

		log.Printf("message published and ACK received on topic %s", *topic)
	}
}
