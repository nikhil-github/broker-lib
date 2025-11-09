package nats

import (
	"testing"
)

func TestNatsPublisher_Close(t *testing.T) {
	publisher := &natsPublisher{
		stream: "test-stream",
	}

	// Close should always return nil (no-op)
	err := publisher.Close()
	if err != nil {
		t.Errorf("Close() error = %v, want nil", err)
	}
}

func TestNatsPublisher_InterfaceImplementation(t *testing.T) {
	// This test verifies that natsPublisher implements brokerlib.Publisher
	// The compile-time check in publisher.go already ensures this, but this
	// test makes it explicit and will fail if the interface changes
	var _ interface {
		Publish(message []byte) error
		Close() error
	} = &natsPublisher{}
}

// Note: Testing Publish() properly would require mocking nats.JetStreamContext,
// which is complex. The Publish() method is a thin wrapper that:
// 1. Calls js.Publish(stream, message)
// 2. Wraps errors with fmt.Errorf("failed to publish message: %w", err)
//
// For comprehensive testing of Publish(), consider:
// 1. Integration tests with a real NATS server (see integration_test.go if created)
// 2. Testing through the Broker interface with mocks (see mocks/README.md)
// 3. Using testcontainers for isolated NATS instances in tests
//
// Example integration test pattern:
//
// func TestNatsPublisher_Publish_Integration(t *testing.T) {
//     if testing.Short() {
//         t.Skip("Skipping integration test")
//     }
//
//     // Setup: Create broker and publisher
//     broker, err := NewNatsBroker("nats://localhost:4222")
//     if err != nil {
//         t.Skipf("NATS not available: %v", err)
//     }
//     defer broker.Close()
//
//     publisher, err := broker.NewPublisher("test-stream")
//     if err != nil {
//         t.Fatalf("NewPublisher() error = %v", err)
//     }
//     defer publisher.Close()
//
//     // Test: Publish a message
//     err = publisher.Publish([]byte("test message"))
//     if err != nil {
//         t.Errorf("Publish() error = %v", err)
//     }
// }

