package brokerlib

import (
	"time"
)

// Message represents a message received from the broker
type Message struct {
	Data      []byte
	Stream    string
	Timestamp time.Time
	Headers   map[string]string
	ackFunc   func() error
	nackFunc  func() error
	termFunc  func() error
}

// NewMessage creates a new Message with acknowledgment functions
// This is used internally by broker implementations
func NewMessage(data []byte, stream string, timestamp time.Time, headers map[string]string, ackFunc, nackFunc, termFunc func() error) *Message {
	return &Message{
		Data:      data,
		Stream:    stream,
		Timestamp: timestamp,
		Headers:   headers,
		ackFunc:   ackFunc,
		nackFunc:  nackFunc,
		termFunc:  termFunc,
	}
}

// Ack acknowledges the message, indicating successful processing
func (m *Message) Ack() error {
	if m.ackFunc != nil {
		return m.ackFunc()
	}
	return nil
}

// Nack negatively acknowledges the message, requesting redelivery
func (m *Message) Nack() error {
	if m.nackFunc != nil {
		return m.nackFunc()
	}
	return nil
}

// Term terminates the message, stopping redelivery permanently
// This is useful when a message cannot be processed and should not be retried
func (m *Message) Term() error {
	if m.termFunc != nil {
		return m.termFunc()
	}
	return nil
}

// DeliverPolicy specifies where to start delivering messages from
type DeliverPolicy string

const (
	// DeliverAll starts from the beginning of the stream
	DeliverAll DeliverPolicy = "all"
	// DeliverLast starts from the last message
	DeliverLast DeliverPolicy = "last"
	// DeliverNew only delivers new messages (after subscription)
	DeliverNew DeliverPolicy = "new"
	// DeliverLastPerSubject starts from the last message per subject
	DeliverLastPerSubject DeliverPolicy = "last_per_subject"
)

// AckPolicy specifies the acknowledgment policy
type AckPolicy string

const (
	// AckExplicit requires explicit acknowledgment (default)
	AckExplicit AckPolicy = "explicit"
	// AckNone no acknowledgment needed
	AckNone AckPolicy = "none"
	// AckAll automatically acknowledges all messages
	AckAll AckPolicy = "all"
)

// ConsumerConfig holds configuration for creating a durable consumer
// Broker implementations can extend this with broker-specific options
type ConsumerConfig struct {
	// ConsumerName is the unique name for the durable consumer
	ConsumerName string

	// DeliverPolicy specifies where to start delivering messages from
	DeliverPolicy DeliverPolicy

	// AckPolicy specifies the acknowledgment policy
	AckPolicy AckPolicy

	// MaxDeliveries is the maximum number of times a message will be redelivered
	// 0 means unlimited
	MaxDeliveries int

	// AckWait is the time to wait for acknowledgment before redelivery
	AckWait time.Duration
}

// DefaultConsumerConfig returns a ConsumerConfig with sensible defaults
func DefaultConsumerConfig(consumerName string) ConsumerConfig {
	return ConsumerConfig{
		ConsumerName:  consumerName,
		DeliverPolicy: DeliverAll,  // Start from beginning if no ack pending
		AckPolicy:     AckExplicit, // Require explicit acknowledgment
		MaxDeliveries: 0,           // Unlimited
		AckWait:       30 * time.Second,
	}
}

// Broker defines the interface for message broker operations
// All broker implementations must satisfy this interface
type Broker interface {
	// Close closes the broker connection and releases resources
	Close() error

	// NewPublisher creates a publisher bound to a specific stream
	// This is convenient when publishing to the same stream repeatedly
	// Example: publisher := broker.NewPublisher("payouts"); publisher.Publish(data)
	NewPublisher(stream string) (Publisher, error)

	// NewConsumer creates a consumer bound to a specific stream and consumer name
	// Creates the durable consumer with the given configuration if it doesn't exist
	// Assumes the stream already exists - returns error if stream doesn't exist
	// This allows fail-fast behavior and is convenient when consuming from the same stream/consumer repeatedly
	// Example: consumer := broker.NewConsumer("orders", "order-processor", config); consumer.PullBatch(ctx, 10, 5*time.Second)
	NewConsumer(stream string, consumerName string, config ConsumerConfig) (Consumer, error)
}
