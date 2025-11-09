package brokerlib

// Publisher represents a publisher bound to a specific stream
// This allows publishing to a stream without repeatedly passing the stream name
type Publisher interface {
	// Publish sends a message to the bound stream
	Publish(message []byte) error
	// Close releases any resources held by the publisher
	Close() error
}
