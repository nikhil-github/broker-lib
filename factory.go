package brokerlib

import (
	"fmt"
	"net/url"
)

// brokerFactory is a function type that creates a Broker instance from a URL
type brokerFactory func(brokerURL string) (Broker, error)

var factories = make(map[string]brokerFactory)

// RegisterBroker registers a broker factory for a specific URL scheme
// This is called automatically by broker implementations via init()
func RegisterBroker(scheme string, factory brokerFactory) {
	factories[scheme] = factory
}

// NewBroker creates a new Broker instance based on the URL scheme
// Currently supports: "nats://" for NATS JetStream
// Example: "nats://localhost:4222" will create a NATS broker
func NewBroker(brokerURL string) (Broker, error) {
	parsedURL, err := url.Parse(brokerURL)
	if err != nil {
		return nil, fmt.Errorf("invalid broker URL: %w", err)
	}

	scheme := parsedURL.Scheme
	if scheme == "" {
		return nil, fmt.Errorf("missing scheme in broker URL: %s", brokerURL)
	}

	factory, exists := factories[scheme]
	if !exists {
		return nil, fmt.Errorf("unsupported broker scheme: %s (supported: nats). Import the broker package to register it", scheme)
	}

	return factory(brokerURL)
}
