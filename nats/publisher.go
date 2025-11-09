package nats

import (
	"fmt"

	"github.com/nats-io/nats.go"

	brokerlib "github.com/nikhil-github/broker-lib"
)

// natsPublisher implements brokerlib.Publisher for a specific stream
type natsPublisher struct {
	js     nats.JetStreamContext
	stream string
}

// Publish sends a message to the bound stream
func (p *natsPublisher) Publish(message []byte) error {
	_, err := p.js.Publish(p.stream, message)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}
	return nil
}

// Close is a no-op for publisher (connection is managed by broker)
func (p *natsPublisher) Close() error {
	return nil
}

// Ensure natsPublisher implements brokerlib.Publisher
var _ brokerlib.Publisher = (*natsPublisher)(nil)

