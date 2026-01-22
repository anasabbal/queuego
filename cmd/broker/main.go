package main

import (
	"log"
	"os"
	"os/signal"
	"queuego/config"
	"queuego/internal/broker"
	"queuego/internal/server"
	"strconv"
	"syscall"
	"time"
)

var (
	version = "0.1.0"
)

func main() {
	// load config
	cfg := config.New()

	_ = cfg.LoadFromFile("config/config.yml")
	cfg.LoadFromEnv()

	if err := cfg.Validate(); err != nil {
		log.Fatal("config error: ", err)
	}

	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	log.Println("===================================")
	log.Println(" QueueGo Broker")
	log.Println(" Version:", version)
	log.Println(" Listening on:", cfg.Server.Host, cfg.Server.Port)
	log.Println(" Storage:", cfg.Storage.Type)
	log.Println("===================================")

	// create broker
	br := broker.NewBroker(broker.BrokerConfig{
		MaxQueueSize:    cfg.Broker.DefaultQueueSize,
		MessageTTL:      cfg.Broker.MessageTTL,
		CleanupInterval: time.Minute,
	})

	br.Start()

	// create server
	addr := cfg.Server.Host + ":" + strconv.Itoa(cfg.Server.Port)
	srv, err := server.NewServer(addr, br)
	if err != nil {
		log.Fatal("failed to start server:", err)
	}

	go srv.Start()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("shutting down...")

	srv.Stop()
	br.Stop()

	log.Println("shutdown complete")
}
