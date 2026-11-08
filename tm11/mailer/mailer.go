package mailer

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var ErrPermanent = errors.New("kegagalan permanen")

type Mailer interface {
	Send(ctx context.Context, to, subject, body string) error
}

type FileMailer struct {
	path     string
	delay    time.Duration
	failRate float64
	mu       sync.Mutex
}

func NewFileMailer(path string, delay time.Duration, failRate float64) Mailer {
	return &FileMailer{path: path, delay: delay, failRate: failRate}
}

func (m *FileMailer) Send(ctx context.Context, to, subject, body string) error {
	if m.delay > 0 {
		timer := time.NewTimer(m.delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}

	domain := strings.ToLower(strings.TrimSpace(to))
	if idx := strings.LastIndex(domain, "@"); idx >= 0 && domain[idx+1:] == "invalid.test" {
		return fmt.Errorf("%w: alamat %s tidak terdaftar", ErrPermanent, to)
	}

	if m.failRate > 0 && rand.Float64() < m.failRate {
		return errors.New("server email sibuk")
	}

	if err := m.appendLog(to, subject); err != nil {
		return err
	}
	return nil
}

func (m *FileMailer) appendLog(to, subject string) error {
	dir := filepath.Dir(m.path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("membuat folder mail log: %w", err)
		}
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	file, err := os.OpenFile(m.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("membuka berkas log email: %w", err)
	}
	defer file.Close()

	line := fmt.Sprintf("%s %s %s\n", time.Now().Format(time.RFC3339), to, subject)
	if _, err := file.WriteString(line); err != nil {
		return fmt.Errorf("menulis log email: %w", err)
	}
	return nil
}
