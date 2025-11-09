package brokerlib

import (
	"context"
	"time"
)

// Consumer represents a consumer bound to a specific stream and consumer name
// This allows consuming from a stream without repeatedly passing stream/consumerName
type Consumer interface {
	// Subscribe returns a channel for receiving messages
	// Messages are delivered asynchronously via the returned channel
	Subscribe() (<-chan Message, error)

	// PullBatch pulls a batch of messages
	// Returns up to batchSize messages, waiting up to timeout duration
	// If ctx is cancelled, the function will return immediately with ctx.Err()
	PullBatch(ctx context.Context, batchSize int, timeout time.Duration) ([]Message, error)

	// Close releases any resources held by the consumer
	Close() error
}
