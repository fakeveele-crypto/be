package config

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"gopkg.in/natefinch/lumberjack.v2"
)

// NewLogger membuat satu *slog.Logger untuk file log yang spesifik.
// API dan worker tidak boleh berbagi berkas log yang sama, karena mereka
// berjalan sekaligus di proses yang berbeda dan masing-masing punya siklus hidup sendiri.
func NewLogger(filename string) *slog.Logger {
	if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
		panic(err)
	}

	fileLogger := &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    10,
		MaxBackups: 3,
		MaxAge:     28,
		Compress:   true,
	}

	multiWriter := io.MultiWriter(os.Stdout, fileLogger)
	handler := slog.NewJSONHandler(multiWriter, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	return slog.New(handler)
}
