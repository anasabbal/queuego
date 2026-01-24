# Go Message Broker

A lightweight **message broker** written in Go that supports basic publish/subscribe messaging.

## Features

- Topic-based publish/subscribe
- Message acknowledgement (ACK)
- Simple command protocol (CONNECT, PUBLISH, SUBSCRIBE, etc.)
- In-memory storage for fast prototyping
- TCP-based communication

## Installation

Make sure you have Go installed (version ≥ 1.18). Then clone and build the project:
```bash
git clone https://github.com/anasabbal/queuego.git
cd queuego
go build ./cmd/producer
go build ./cmd/consumer
```

Or run directly using Go:
```bash
go run ./cmd/producer/main.go
go run ./cmd/consumer/main.go
```

## Usage

### Producer

Send messages to a topic:
```bash
go run ./cmd/producer/main.go --topic=test --message="Hello queuego" --count=1
```

**Flags:**
- `--topic`: Topic name to publish the message
- `--message`: Message content
- `--count`: Number of times to send the message

### Consumer

Subscribe to a topic and receive messages:
```bash
go run ./cmd/consumer/main.go --topic=test
```

**Flags:**
- `--topic`: Topic name to subscribe to

Messages published to the topic will appear in the consumer terminal.

## Example

Open two terminals:

**Terminal 1 – Consumer:**
```bash
go run ./cmd/consumer/main.go --topic=test
```

**Terminal 2 – Producer:**
```bash
go run ./cmd/producer/main.go --topic=test --message="Hello queuego" --count=5
```

The consumer will receive and display the messages published by the producer.
