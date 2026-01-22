package main

import (
	"flag"
	"log"
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
		// Publish waits for ACK internally
		if err := producer.Publish(*topic, []byte(*message)); err != nil {
			log.Fatal(err)
		}

		// No second ReadResponse here — Publish already consumed the ACK
		log.Printf("message published and ACK received on topic %s", *topic)
	}
}
