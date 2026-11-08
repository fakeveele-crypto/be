package worker

import (
	"testing"
	"time"

	"github.com/hibiken/asynq"
)

func TestRetryDelay(t *testing.T) {
	tests := []struct {
		name string
		n    int
		want time.Duration
	}{
		{name: "pertama", n: 0, want: 2 * time.Second},
		{name: "kedua", n: 1, want: 4 * time.Second},
		{name: "ketiga", n: 2, want: 6 * time.Second},
		{name: "keempat", n: 3, want: 8 * time.Second},
		{name: "kelima", n: 4, want: 10 * time.Second},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := RetryDelay(tc.n, nil, &asynq.Task{})
			if got != tc.want {
				t.Fatalf("RetryDelay(%d) = %v, want %v", tc.n, got, tc.want)
			}
		})
	}
}
