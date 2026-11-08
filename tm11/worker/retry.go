package worker

import (
	"time"

	"github.com/hibiken/asynq"
)

const MaxRetry = 5

// MaxRetry = 5 berarti satu percobaan awal ditambah lima percobaan ulang,
// total enam kali coba. Pola backoff ini mengikuti slide 2, 4, 6, 8, 10 detik.
// Asynq mengirim n = 0 untuk percobaan pertama yang gagal, jadi jeda
// untuk retry ke-1 adalah (0+1)*2 detik, lalu bertambah dua detik tiap kali.
func RetryDelay(n int, err error, t *asynq.Task) time.Duration {
	if n < 0 {
		n = 0
	}
	if n >= MaxRetry {
		n = MaxRetry - 1
	}
	return time.Duration(n+1) * 2 * time.Second
}
