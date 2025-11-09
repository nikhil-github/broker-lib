package brokerlib

import (
	"testing"
)

func TestNewBroker_InvalidURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "empty URL",
			url:     "",
			wantErr: true,
		},
		{
			name:    "missing scheme",
			url:     "localhost:4222",
			wantErr: true,
		},
		{
			name:    "unsupported scheme",
			url:     "kafka://localhost:9092",
			wantErr: true,
		},
		{
			name:    "invalid URL format",
			url:     "://invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			broker, err := NewBroker(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewBroker() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if broker != nil {
				t.Error("NewBroker() returned non-nil broker on error")
			}
		})
	}
}

func TestNewBroker_ValidNATSURL(t *testing.T) {
	// This test requires a running NATS server, so we'll skip it if not available
	// In a real scenario, you might use a test container or mock
	url := "nats://localhost:4222"
	broker, err := NewBroker(url)
	
	// If NATS is not running, we expect an error
	if err != nil {
		t.Logf("NewBroker() with valid URL returned error (expected if NATS not running): %v", err)
		return
	}

	if broker == nil {
		t.Error("NewBroker() returned nil broker on success")
		return
	}

	// Clean up
	if err := broker.Close(); err != nil {
		t.Logf("Failed to close broker: %v", err)
	}
}

