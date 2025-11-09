package nats

import (
	"context"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	brokerlib "github.com/nikhil-github/broker-lib"
)

// natsBroker implements the brokerlib.Broker interface for NATS JetStream
type natsBroker struct {
	conn *nats.Conn
	js   nats.JetStreamContext
}

// NewNatsBroker creates a new NATS JetStream broker instance
func NewNatsBroker(brokerURL string) (brokerlib.Broker, error) {
	conn, err := nats.Connect(brokerURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	js, err := conn.JetStream()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to get JetStream context: %w", err)
	}

	return &natsBroker{
		conn: conn,
		js:   js,
	}, nil
}

// Subscribe returns a channel for receiving messages from an existing consumer
// The consumer must have been created via CreateConsumer first
// Note: 'stream' is used as subject for PullSubscribe
func (n *natsBroker) Subscribe(stream string, consumerName string) (<-chan brokerlib.Message, error) {
	// Verify consumer exists (fail fast if not created)
	streamName := stream
	_, err := n.js.ConsumerInfo(streamName, consumerName)
	if err != nil {
		return nil, fmt.Errorf("consumer does not exist: %s (must call CreateConsumer first): %w", consumerName, err)
	}

	// Create pull subscription using stream as subject
	subject := stream
	sub, err := n.js.PullSubscribe(subject, consumerName)
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
				if n.conn.IsClosed() {
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

// PullBatch pulls a batch of messages from an existing durable consumer
// The consumer must have been created via CreateConsumer first
// Returns up to batchSize messages, waiting up to timeout duration
// If ctx is cancelled, the function will return immediately with ctx.Err()
// Note: 'stream' is used as subject for PullSubscribe
func (n *natsBroker) PullBatch(ctx context.Context, stream string, consumerName string, batchSize int, timeout time.Duration) ([]brokerlib.Message, error) {
	// Check context cancellation before starting
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Verify consumer exists (fail fast if not created)
	_, err := n.js.ConsumerInfo(stream, consumerName)
	if err != nil {
		return nil, fmt.Errorf("consumer does not exist: %s (must call CreateConsumer first): %w", consumerName, err)
	}

	// Create pull subscription using stream as subject
	subject := stream
	sub, err := n.js.PullSubscribe(subject, consumerName)
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

// Close closes the NATS connection and releases resources
func (n *natsBroker) Close() error {
	if n.conn != nil && !n.conn.IsClosed() {
		n.conn.Close()
	}
	return nil
}

// convertDeliverPolicy converts brokerlib.DeliverPolicy to NATS DeliverPolicy
func convertDeliverPolicy(policy brokerlib.DeliverPolicy) nats.DeliverPolicy {
	switch policy {
	case brokerlib.DeliverAll:
		return nats.DeliverAllPolicy
	case brokerlib.DeliverLast:
		return nats.DeliverLastPolicy
	case brokerlib.DeliverNew:
		return nats.DeliverNewPolicy
	case brokerlib.DeliverLastPerSubject:
		return nats.DeliverLastPerSubjectPolicy
	default:
		return nats.DeliverAllPolicy
	}
}

// convertAckPolicy converts brokerlib.AckPolicy to NATS AckPolicy
func convertAckPolicy(policy brokerlib.AckPolicy) nats.AckPolicy {
	switch policy {
	case brokerlib.AckExplicit:
		return nats.AckExplicitPolicy
	case brokerlib.AckNone:
		return nats.AckNonePolicy
	case brokerlib.AckAll:
		return nats.AckAllPolicy
	default:
		return nats.AckExplicitPolicy
	}
}

// NewPublisher creates a publisher bound to a specific stream
func (n *natsBroker) NewPublisher(stream string) (brokerlib.Publisher, error) {
	return &natsPublisher{
		js:     n.js,
		stream: stream,
	}, nil
}

// NewConsumer creates a consumer bound to a specific stream and consumer name
// Creates the durable consumer with the given configuration if it doesn't exist
// Assumes the stream already exists - returns error if stream doesn't exist
// Note: In NATS JetStream, 'stream' is used as the stream name (for StreamInfo, AddConsumer)
// and also as the subject (for PullSubscribe). This works when stream name equals subject.
func (n *natsBroker) NewConsumer(stream string, consumerName string, config brokerlib.ConsumerConfig) (brokerlib.Consumer, error) {
	// Ensure consumerName in config matches the parameter
	if config.ConsumerName != consumerName {
		return nil, fmt.Errorf("consumer name mismatch: config.ConsumerName (%s) must match consumerName parameter (%s)", config.ConsumerName, consumerName)
	}

	// Check if stream exists
	streamName := stream
	_, err := n.js.StreamInfo(streamName)
	if err != nil {
		return nil, fmt.Errorf("stream does not exist: %s: %w", streamName, err)
	}

	// Check if consumer already exists
	existingConsumer, err := n.js.ConsumerInfo(streamName, consumerName)
	if err == nil && existingConsumer != nil {
		// Consumer already exists - check if config matches
		// If config matches, return success (idempotent)
		// If config differs, return error to enforce explicit recreation
		existingConfig := existingConsumer.Config

		// Convert our config to NATS config for comparison
		deliverPolicy := convertDeliverPolicy(config.DeliverPolicy)
		ackPolicy := convertAckPolicy(config.AckPolicy)

		// Check if key config matches
		configMatches := existingConfig.Durable == consumerName &&
			existingConfig.DeliverPolicy == deliverPolicy &&
			existingConfig.AckPolicy == ackPolicy &&
			(config.MaxDeliveries == 0 || existingConfig.MaxDeliver == config.MaxDeliveries) &&
			(config.AckWait == 0 || existingConfig.AckWait == config.AckWait)

		if !configMatches {
			// Consumer exists but config differs - must be recreated
			return nil, fmt.Errorf("consumer already exists with different config: %s (delete and recreate with desired config)", consumerName)
		}
		// Consumer exists with matching config - proceed to create consumer object
	} else {
		// Consumer doesn't exist - create it
		// Convert deliver policy and ack policy
		deliverPolicy := convertDeliverPolicy(config.DeliverPolicy)
		ackPolicy := convertAckPolicy(config.AckPolicy)

		// Build NATS consumer config
		natsConfig := &nats.ConsumerConfig{
			Durable:       consumerName,
			DeliverPolicy: deliverPolicy,
			AckPolicy:     ackPolicy,
		}

		// Set optional fields
		if config.MaxDeliveries > 0 {
			natsConfig.MaxDeliver = config.MaxDeliveries
		}
		if config.AckWait > 0 {
			natsConfig.AckWait = config.AckWait
		}

		// Create durable consumer on the stream
		_, err = n.js.AddConsumer(streamName, natsConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create consumer: %w", err)
		}
	}

	// Create and return the consumer object
	return &natsConsumer{
		js:           n.js,
		conn:         n.conn,
		stream:       stream,
		consumerName: consumerName,
	}, nil
}

// init registers the NATS broker factory
func init() {
	brokerlib.RegisterBroker("nats", NewNatsBroker)
}
