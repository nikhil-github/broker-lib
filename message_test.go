package brokerlib

import (
	"errors"
	"testing"
	"time"
)

func TestMessage_Ack(t *testing.T) {
	tests := []struct {
		name      string
		ackFunc   func() error
		wantErr   bool
		wantErrMsg string
	}{
		{
			name: "successful ack",
			ackFunc: func() error {
				return nil
			},
			wantErr: false,
		},
		{
			name: "ack with error",
			ackFunc: func() error {
				return errors.New("ack failed")
			},
			wantErr:   true,
			wantErrMsg: "ack failed",
		},
		{
			name:    "nil ackFunc",
			ackFunc: nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := &Message{
				Data:    []byte("test"),
				ackFunc: tt.ackFunc,
			}

			err := msg.Ack()
			if (err != nil) != tt.wantErr {
				t.Errorf("Ack() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.wantErrMsg {
				t.Errorf("Ack() error = %v, wantErrMsg %v", err, tt.wantErrMsg)
			}
		})
	}
}

func TestMessage_Nack(t *testing.T) {
	tests := []struct {
		name      string
		nackFunc  func() error
		wantErr   bool
		wantErrMsg string
	}{
		{
			name: "successful nack",
			nackFunc: func() error {
				return nil
			},
			wantErr: false,
		},
		{
			name: "nack with error",
			nackFunc: func() error {
				return errors.New("nack failed")
			},
			wantErr:   true,
			wantErrMsg: "nack failed",
		},
		{
			name:    "nil nackFunc",
			nackFunc: nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := &Message{
				Data:     []byte("test"),
				nackFunc: tt.nackFunc,
			}

			err := msg.Nack()
			if (err != nil) != tt.wantErr {
				t.Errorf("Nack() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.wantErrMsg {
				t.Errorf("Nack() error = %v, wantErrMsg %v", err, tt.wantErrMsg)
			}
		})
	}
}

func TestMessage_Term(t *testing.T) {
	tests := []struct {
		name      string
		termFunc  func() error
		wantErr   bool
		wantErrMsg string
	}{
		{
			name: "successful term",
			termFunc: func() error {
				return nil
			},
			wantErr: false,
		},
		{
			name: "term with error",
			termFunc: func() error {
				return errors.New("term failed")
			},
			wantErr:   true,
			wantErrMsg: "term failed",
		},
		{
			name:    "nil termFunc",
			termFunc: nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := &Message{
				Data:     []byte("test"),
				termFunc: tt.termFunc,
			}

			err := msg.Term()
			if (err != nil) != tt.wantErr {
				t.Errorf("Term() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.wantErrMsg {
				t.Errorf("Term() error = %v, wantErrMsg %v", err, tt.wantErrMsg)
			}
		})
	}
}

func TestNewMessage(t *testing.T) {
	ackCalled := false
	nackCalled := false
	termCalled := false

	ackFunc := func() error {
		ackCalled = true
		return nil
	}
	nackFunc := func() error {
		nackCalled = true
		return nil
	}
	termFunc := func() error {
		termCalled = true
		return nil
	}

	data := []byte("test data")
	stream := "test-stream"
	timestamp := time.Now()
	headers := map[string]string{"key": "value"}

	msg := NewMessage(data, stream, timestamp, headers, ackFunc, nackFunc, termFunc)

	if msg == nil {
		t.Fatal("NewMessage() returned nil")
	}

	if string(msg.Data) != string(data) {
		t.Errorf("NewMessage() Data = %v, want %v", msg.Data, data)
	}

	if msg.Stream != stream {
		t.Errorf("NewMessage() Stream = %v, want %v", msg.Stream, stream)
	}

	if !msg.Timestamp.Equal(timestamp) {
		t.Errorf("NewMessage() Timestamp = %v, want %v", msg.Timestamp, timestamp)
	}

	if msg.Headers["key"] != "value" {
		t.Errorf("NewMessage() Headers = %v, want %v", msg.Headers, headers)
	}

	// Test that functions are properly set
	msg.Ack()
	if !ackCalled {
		t.Error("Ack() function was not called")
	}

	msg.Nack()
	if !nackCalled {
		t.Error("Nack() function was not called")
	}

	msg.Term()
	if !termCalled {
		t.Error("Term() function was not called")
	}
}

