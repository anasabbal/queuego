#!/bin/bash

# Set broker address
BROKER_ADDR="127.0.0.1:9092"
TOPIC="test"

# -------------------------
# Step 1: Start broker
# -------------------------
echo "Starting QueueGo broker..."
# Run broker in background
go run ./cmd/broker/main.go &
BROKER_PID=$!
echo "Broker PID: $BROKER_PID"

# Wait a few seconds for broker to start
sleep 2

# -------------------------
# Step 2: Start consumer
# -------------------------
echo "Starting consumer for topic '$TOPIC'..."
go run ./cmd/consumer/main.go --topic="$TOPIC" &
CONSUMER_PID=$!
echo "Consumer PID: $CONSUMER_PID"

# Wait a second
sleep 1

# -------------------------
# Step 3: Start producer
# -------------------------
echo "Publishing a test message to topic '$TOPIC'..."
go run ./cmd/producer/main.go --topic="$TOPIC" --message="Hello World" --count=1

# -------------------------
# Optional: Keep the broker and consumer running
# -------------------------
echo "Press Ctrl+C to stop broker and consumer..."
wait $BROKER_PID
wait $CONSUMER_PID
