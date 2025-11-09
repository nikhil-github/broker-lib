# Mocks

This directory contains generated mocks for the broker-lib interfaces using [go.uber.org/mock](https://github.com/uber-go/mock).

## Generated Mocks

- `mock_broker.go` - Mock for `brokerlib.Broker` interface
- `mock_publisher.go` - Mock for `brokerlib.Publisher` interface
- `mock_consumer.go` - Mock for `brokerlib.Consumer` interface

## Regenerating Mocks

If you modify the interfaces, regenerate the mocks using:

```bash
mockgen -source=../broker.go -destination=mock_broker.go -package=mocks
mockgen -source=../publisher.go -destination=mock_publisher.go -package=mocks
mockgen -source=../consumer.go -destination=mock_consumer.go -package=mocks
```

## Usage Example

```go
package yourpackage

import (
    "testing"
    "context"
    "time"
    
    "github.com/nikhil-github/broker-lib"
    "github.com/nikhil-github/broker-lib/mocks"
    "go.uber.org/mock/gomock"
)

func TestYourFunction(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    // Create mocks
    mockBroker := mocks.NewMockBroker(ctrl)
    mockPublisher := mocks.NewMockPublisher(ctrl)
    mockConsumer := mocks.NewMockConsumer(ctrl)

    // Setup expectations
    mockBroker.EXPECT().
        NewPublisher("test-stream").
        Return(mockPublisher, nil)

    mockPublisher.EXPECT().
        Publish([]byte("test message")).
        Return(nil)

    mockPublisher.EXPECT().
        Close().
        Return(nil)

    // Use the mocks in your code
    publisher, err := mockBroker.NewPublisher("test-stream")
    if err != nil {
        t.Fatalf("NewPublisher() error = %v", err)
    }

    err = publisher.Publish([]byte("test message"))
    if err != nil {
        t.Fatalf("Publish() error = %v", err)
    }

    err = publisher.Close()
    if err != nil {
        t.Fatalf("Close() error = %v", err)
    }
}

func TestConsumerExample(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockBroker := mocks.NewMockBroker(ctrl)
    mockConsumer := mocks.NewMockConsumer(ctrl)

    config := brokerlib.DefaultConsumerConfig("test-consumer")

    // Setup expectations
    mockBroker.EXPECT().
        NewConsumer("test-stream", "test-consumer", config).
        Return(mockConsumer, nil)

    msgChan := make(chan brokerlib.Message, 1)
    mockConsumer.EXPECT().
        Subscribe().
        Return(msgChan, nil)

    mockConsumer.EXPECT().
        Close().
        Return(nil)

    // Use the mocks
    consumer, err := mockBroker.NewConsumer("test-stream", "test-consumer", config)
    if err != nil {
        t.Fatalf("NewConsumer() error = %v", err)
    }

    ch, err := consumer.Subscribe()
    if err != nil {
        t.Fatalf("Subscribe() error = %v", err)
    }

    // Test with messages
    _ = ch

    err = consumer.Close()
    if err != nil {
        t.Fatalf("Close() error = %v", err)
    }
}
```

