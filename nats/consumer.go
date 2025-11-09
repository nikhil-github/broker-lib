package nats

import (
	"context"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"

	brokerlib "github.com/nikhil-github/broker-lib"
)

// convertHeaders converts NATS headers to a map
func convertHeaders(header nats.Header) map[string]string {
	if header == nil {
		return make(map[string]string)
	}

	result := make(map[string]string)
	for key, values := range header {
		if len(values) > 0 {
			result[key] = values[0] // Take first value
		}
	}
	return result
}

// natsConsumer implements brokerlib.Consumer for a specific stream and consumer name
type natsConsumer struct {
	js           nats.JetStreamContext
	conn         *nats.Conn
	stream       string
	consumerName string
}

// Subscribe returns a channel for receiving messages from the bound consumer
func (c *natsConsumer) Subscribe() (<-chan brokerlib.Message, error) {
	// Verify consumer exists (fail fast if not created)
	_, err := c.js.ConsumerInfo(c.stream, c.consumerName)
	if err != nil {
		return nil, fmt.Errorf("consumer does not exist: %s (must call CreateConsumer first): %w", c.consumerName, err)
	}

	// Create pull subscription using stream as subject
	subject := c.stream
	sub, err := c.js.PullSubscribe(subject, c.consumerName)
	if err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	msgChan := make(chan brokerlib.Message, 100) // buffered channel

	// Start goroutine to fetch and deliver messages
	go func() {
		defer close(msgChan)
		for {
			// Fetch one message at a time with timeout
			msgs, err := sub.Fetch(1, nats.MaxWait(5*time.Second))
			if err != nil {
				// Timeout or connection error - check if connection is still alive
				if c.conn.IsClosed() {
					return
				}
				// Continue on timeout, break on other errors
				if err == nats.ErrTimeout {
					continue
				}
				return
			}

			for _, msg := range msgs {
				// Convert NATS message to brokerlib.Message
				meta, _ := msg.Metadata()
				brokerMsg := *brokerlib.NewMessage(
					msg.Data,
					msg.Subject, // Subject is the stream name
					meta.Timestamp,
					convertHeaders(msg.Header),
					func() error { return msg.Ack() },
					func() error { return msg.Nak() },
					func() error { return msg.Term() },
				)

				select {
				case msgChan <- brokerMsg:
				case <-time.After(10 * time.Second):
					// Channel full or receiver gone, nack the message
					msg.Nak()
					return
				}
			}
		}
	}()

	return msgChan, nil
}

// PullBatch pulls a batch of messages from the bound consumer
func (c *natsConsumer) PullBatch(ctx context.Context, batchSize int, timeout time.Duration) ([]brokerlib.Message, error) {
	// Check context cancellation before starting
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Verify consumer exists (fail fast if not created)
	_, err := c.js.ConsumerInfo(c.stream, c.consumerName)
	if err != nil {
		return nil, fmt.Errorf("consumer does not exist: %s (must call CreateConsumer first): %w", c.consumerName, err)
	}

	// Create pull subscription using stream as subject
	subject := c.stream
	sub, err := c.js.PullSubscribe(subject, c.consumerName)
	if err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}
	// Ensure subscription is cleaned up
	defer sub.Unsubscribe()

	// Create context with timeout - if parent context is cancelled, this will be cancelled too
	fetchCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Fetch batch of messages - NATS Fetch supports context directly
	natsMsgs, err := sub.Fetch(batchSize, nats.Context(fetchCtx))

	if err != nil {
		// Check if error is due to context cancellation
		if fetchCtx.Err() != nil {
			return nil, fetchCtx.Err()
		}
		if err == nats.ErrTimeout {
			// Timeout is acceptable, return empty slice
			return []brokerlib.Message{}, nil
		}
		return nil, fmt.Errorf("failed to fetch messages: %w", err)
	}

	// Convert NATS messages to brokerlib messages
	messages := make([]brokerlib.Message, len(natsMsgs))
	for i, msg := range natsMsgs {
		meta, _ := msg.Metadata()
		messages[i] = *brokerlib.NewMessage(
			msg.Data,
			msg.Subject, // Subject is the stream name
			meta.Timestamp,
			convertHeaders(msg.Header),
			func() error { return msg.Ack() },
			func() error { return msg.Nak() },
			func() error { return msg.Term() },
		)
	}

	return messages, nil
}

// Close is a no-op for consumer (connection is managed by broker)
func (c *natsConsumer) Close() error {
	return nil
}

// Ensure natsConsumer implements brokerlib.Consumer
var _ brokerlib.Consumer = (*natsConsumer)(nil)
