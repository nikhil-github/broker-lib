package nats

import (
	"testing"

	"github.com/nats-io/nats.go"
)

func TestConvertHeaders(t *testing.T) {
	tests := []struct {
		name   string
		header nats.Header
		want   map[string]string
	}{
		{
			name:   "nil header",
			header: nil,
			want:   make(map[string]string),
		},
		{
			name:   "empty header",
			header: nats.Header{},
			want:   make(map[string]string),
		},
		{
			name: "single value header",
			header: nats.Header{
				"key1": []string{"value1"},
			},
			want: map[string]string{
				"key1": "value1",
			},
		},
		{
			name: "multiple value header - takes first",
			header: nats.Header{
				"key1": []string{"value1", "value2", "value3"},
			},
			want: map[string]string{
				"key1": "value1",
			},
		},
		{
			name: "multiple keys",
			header: nats.Header{
				"key1": []string{"value1"},
				"key2": []string{"value2"},
				"key3": []string{"value3"},
			},
			want: map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
		},
		{
			name: "empty value array - skipped",
			header: nats.Header{
				"key1": []string{"value1"},
				"key2": []string{},
			},
			want: map[string]string{
				"key1": "value1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertHeaders(tt.header)
			
			if len(got) != len(tt.want) {
				t.Errorf("convertHeaders() length = %v, want %v", len(got), len(tt.want))
				return
			}

			for key, wantValue := range tt.want {
				if gotValue, ok := got[key]; !ok {
					t.Errorf("convertHeaders() missing key %v", key)
				} else if gotValue != wantValue {
					t.Errorf("convertHeaders() [%v] = %v, want %v", key, gotValue, wantValue)
				}
			}
		})
	}
}

