package config

import "github.com/hibiken/asynq"

// RedisClientOpt membuat koneksi penyimpanan task untuk relay dan worker.
// API tidak menyentuh Redis, jadi ia tidak perlu dipanggil di mode api.
func RedisClientOpt() asynq.RedisClientOpt {
	return asynq.RedisClientOpt{
		Addr:     GetEnv("REDIS_ADDR", "localhost:6379"),
		Password: GetEnv("REDIS_PASSWORD", ""),
		DB:       GetEnvInt("REDIS_DB", 0),
	}
}
