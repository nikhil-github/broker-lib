# Broker Library

A Go library that abstracts message broker functionality for pub/sub patterns using the adapter pattern. This library decouples application logic from specific broker implementations, allowing you to swap brokers seamlessly.

## Features

- **Interface-based Design**: Clean abstraction over broker implementations
- **Durable Consumers**: Support for named consumers that resume from last position
- **Channel-based Delivery**: Receive messages asynchronously via Go channels
- **Batch Pulling**: Pull multiple messages at once for efficient processing
- **URL-based Selection**: Automatically select broker implementation from URL scheme
- **Easy Extension**: Add new broker implementations by implementing the `Broker` interface

## Supported Brokers

- **NATS JetStream** (`nats://`) - Fully implemented

## Installation

```bash
go get github.com/nikhil-github/broker-lib
```

## Quick Start

### Publishing Messages

```go
package main

import (
	"log"
	"github.com/nikhil-github/broker-lib"
	_ "github.com/nikhil-github/broker-lib/nats" // Register NATS broker
)

func main() {
	// Create broker - URL scheme determines implementation
	broker, err := brokerlib.NewBroker("nats://localhost:4222")
	if err != nil {
		log.Fatal(err)
	}
	defer broker.Close()

	// Create a publisher for the stream
	publisher, err := broker.NewPublisher("orders")
	if err != nil {
		log.Fatal(err)
	}
	defer publisher.Close()

	// Publish a message
	err = publisher.Publish([]byte("Order #123: Process payment"))
	if err != nil {
		log.Fatal(err)
	}
}
```

### Subscribing with Channel (Push-based)

```go
package main

import (
	"log"
	"github.com/nikhil-github/broker-lib"
	_ "github.com/nikhil-github/broker-lib/nats"
)

func main() {
	broker, err := brokerlib.NewBroker("nats://localhost:4222")
	if err != nil {
		log.Fatal(err)
	}
	defer broker.Close()

	// Create consumer with configuration (creates consumer if it doesn't exist)
	config := brokerlib.DefaultConsumerConfig("order-processor")
	consumer, err := broker.NewConsumer("orders", "order-processor", config)
	if err != nil {
		log.Fatal(err) // Fail fast if stream doesn't exist or consumer creation fails
	}
	defer consumer.Close()

	// Subscribe - returns a channel for receiving messages
	msgChan, err := consumer.Subscribe()
	if err != nil {
		log.Fatal(err)
	}

	// Process messages from channel
	for msg := range msgChan {
		log.Printf("Received: %s", string(msg.Data))
		
		// Process message...
		
		// Acknowledge successful processing
		if err := msg.Ack(); err != nil {
			log.Printf("Failed to ack: %v", err)
		}
	}
}
```

### Pulling Batches (Pull-based)

```go
package main

import (
	"log"
	"time"
	"github.com/nikhil-github/broker-lib"
	_ "github.com/nikhil-github/broker-lib/nats"
)

func main() {
	broker, err := brokerlib.NewBroker("nats://localhost:4222")
	if err != nil {
		log.Fatal(err)
	}
	defer broker.Close()

	// Create consumer with configuration first (fail fast)
	config := brokerlib.DefaultConsumerConfig("order-processor")
	// Customize if needed:
	// config.DeliverPolicy = brokerlib.DeliverLast  // Start from last message
	// config.MaxDeliveries = 5       // Max 5 redeliveries
	
	if err := broker.CreateConsumer("orders", config); err != nil {
		log.Fatal(err) // Fail fast if stream doesn't exist
	}

	// Create a consumer for the stream
	consumer, err := broker.NewConsumer("orders", "order-processor")
	if err != nil {
		log.Fatal(err)
	}
	defer consumer.Close()

	// Pull batch of messages
	ctx := context.Background()
	for {
		messages, err := consumer.PullBatch(ctx, 10, 5*time.Second)
		if err != nil {
			log.Printf("Error: %v", err)
			continue
		}

		// Process batch
		for _, msg := range messages {
			log.Printf("Received: %s", string(msg.Data))
			msg.Ack()
		}
	}
}
```

## API Reference

### Broker Interface

```go
type Broker interface {
    // Publish sends a message to the specified stream
    Publish(stream string, message []byte) error

    // CreateConsumer creates a durable consumer with the given configuration
    // Assumes the stream already exists - returns error if stream doesn't exist
    // Must be called before Subscribe or PullBatch. This allows fail-fast behavior.
    CreateConsumer(stream string, config ConsumerConfig) error

    // Subscribe returns a channel for receiving messages from an existing consumer
    // The consumer must have been created via CreateConsumer first
    Subscribe(stream string, consumerName string) (<-chan Message, error)

    // PullBatch pulls a batch of messages from an existing durable consumer
    // The consumer must have been created via CreateConsumer first
    PullBatch(ctx context.Context, stream string, consumerName string, batchSize int, timeout time.Duration) ([]Message, error)

    // Close closes the broker connection and releases resources
    Close() error

    // NewPublisher creates a publisher bound to a specific stream
    NewPublisher(stream string) (Publisher, error)

    // NewConsumer creates a consumer bound to a specific stream and consumer name
    NewConsumer(stream string, consumerName string) (Consumer, error)
}
```

### ConsumerConfig

```go
type ConsumerConfig struct {
    ConsumerName  string        // Unique name for the durable consumer
    DeliverPolicy DeliverPolicy // DeliverAll, DeliverLast, DeliverNew, DeliverLastPerSubject
    AckPolicy     AckPolicy     // AckExplicit, AckNone, AckAll
    MaxDeliveries int           // Max redeliveries (0 = unlimited)
    AckWait       time.Duration // Time to wait for ack before redelivery
}

// DeliverPolicy constants
const (
    DeliverAll            DeliverPolicy = "all"
    DeliverLast           DeliverPolicy = "last"
    DeliverNew            DeliverPolicy = "new"
    DeliverLastPerSubject DeliverPolicy = "last_per_subject"
)

// AckPolicy constants
const (
    AckExplicit AckPolicy = "explicit"
    AckNone     AckPolicy = "none"
    AckAll      AckPolicy = "all"
)

// Helper function to create config with defaults
config := brokerlib.DefaultConsumerConfig("my-consumer")
```

### Message Type

```go
type Message struct {
    Data      []byte
    Stream    string
    Timestamp time.Time
    Headers   map[string]string
}

// Ack acknowledges the message, indicating successful processing
func (m *Message) Ack() error

// Nack negatively acknowledges the message, requesting redelivery
func (m *Message) Nack() error
```

## Important Notes

### Stream Requirements

**The library assumes streams already exist.** If a stream doesn't exist, `NewConsumer` will return an error. You must create streams using your broker's native tools or admin API before using this library.

For NATS JetStream, you can create a stream using the NATS CLI:

```bash
nats stream add ORDERS --subjects orders
```

Or programmatically:

```go
js, _ := nc.JetStream()
js.AddStream(&nats.StreamConfig{
    Name:     "ORDERS",
    Subjects: []string{"orders"},
})
```

### Durable Consumers and Fail-Fast Pattern

- **Consumer Creation**: Consumers are created automatically when you call `NewConsumer()` with a configuration
- **Fail-Fast Behavior**: If the stream doesn't exist or consumer creation fails, you'll know immediately
- **Configuration**: All consumer settings (deliver policy, ack policy, etc.) are set at creation time
- **Idempotent**: If a consumer with the same name and config already exists, `NewConsumer()` succeeds (idempotent)
- **Persistent State**: Consumer names are persistent across application restarts
- **Resume Position**: Consumers resume from the last acknowledged message position
- **Isolation**: Each consumer maintains its own position in the stream

### Message Acknowledgment

- **Always acknowledge messages** after successful processing using `msg.Ack()`
- Unacknowledged messages will be redelivered
- Use `msg.Nack()` to request redelivery if processing fails

## Architecture

The library uses the **Adapter Pattern** to decouple application code from broker implementations:

1. **Interface Layer** (`broker.go`): Defines the `Broker` interface
2. **Factory Layer** (`factory.go`): URL-based broker creation
3. **Adapter Layer** (`nats/nats.go`): NATS JetStream implementation

### Adding New Broker Implementations

To add support for a new broker (e.g., Kafka, Redis Streams):

1. Create a new package (e.g., `kafka/kafka.go`)
2. Implement the `Broker` interface
3. Register the factory in `init()`:

```go
package kafka

import "github.com/nikhil-github/broker-lib"

func init() {
    brokerlib.RegisterBroker("kafka", NewKafkaBroker)
}
```

4. Import the package in your application:

```go
import _ "github.com/nikhil-github/broker-lib/kafka"
```

## Examples

See the `examples/` directory for complete working examples:

- `publisher.go`: Publishing messages
- `subscriber_channel.go`: Channel-based subscription
- `subscriber_batch.go`: Batch pulling

## License

[Add your license here]

# broker-lib
