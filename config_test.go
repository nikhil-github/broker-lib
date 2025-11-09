package brokerlib

import (
	"testing"
	"time"
)

func TestDefaultConsumerConfig(t *testing.T) {
	consumerName := "test-consumer"
	config := DefaultConsumerConfig(consumerName)

	if config.ConsumerName != consumerName {
		t.Errorf("DefaultConsumerConfig() ConsumerName = %v, want %v", config.ConsumerName, consumerName)
	}

	if config.DeliverPolicy != DeliverAll {
		t.Errorf("DefaultConsumerConfig() DeliverPolicy = %v, want %v", config.DeliverPolicy, DeliverAll)
	}

	if config.AckPolicy != AckExplicit {
		t.Errorf("DefaultConsumerConfig() AckPolicy = %v, want %v", config.AckPolicy, AckExplicit)
	}

	if config.MaxDeliveries != 0 {
		t.Errorf("DefaultConsumerConfig() MaxDeliveries = %v, want %v", config.MaxDeliveries, 0)
	}

	expectedAckWait := 30 * time.Second
	if config.AckWait != expectedAckWait {
		t.Errorf("DefaultConsumerConfig() AckWait = %v, want %v", config.AckWait, expectedAckWait)
	}
}

func TestDefaultConsumerConfig_DifferentNames(t *testing.T) {
	names := []string{"consumer-1", "consumer-2", "test-consumer"}

	for _, name := range names {
		config := DefaultConsumerConfig(name)
		if config.ConsumerName != name {
			t.Errorf("DefaultConsumerConfig() ConsumerName = %v, want %v", config.ConsumerName, name)
		}
	}
}

