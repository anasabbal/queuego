package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"queuego/internal/protocol"
	"queuego/pkg/client"
	"syscall"
	"time"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:9092", "broker address")
	topic := flag.String("topic", "test", "topic name")
	flag.Parse()

	cfg := client.ClientConfig{
		RetryMax:      5,
		RetryInterval: time.Second,
		ConnTimeout:   5 * time.Second,
	}

	consumer := client.NewConsumer(cfg)

	if err := consumer.Connect(*addr); err != nil {
		log.Fatal(err)
	}
	defer consumer.Disconnect()

	err := consumer.Subscribe(*topic, func(cmd *protocol.Command) {
		log.Printf("received message on topic %s: %s", cmd.Topic, string(cmd.Payload))
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("subscribed to topic %s", *topic)

	// wait for SIGINT / SIGTERM
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("shutting down consumer...")
	_ = consumer.Unsubscribe(*topic)
}
