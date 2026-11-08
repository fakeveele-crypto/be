package mailer

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFileMailerSendsAndLogs(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mail.log")
	m := NewFileMailer(path, 0, 0)

	err := m.Send(context.Background(), "student@example.com", "Selamat", "Halo dunia")
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected 1 mail log line, got %d", len(lines))
	}
	if !strings.Contains(lines[0], "student@example.com") {
		t.Fatalf("log line missing recipient: %q", lines[0])
	}
}

func TestFileMailerPermanentFailureDoesNotWriteLog(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mail.log")
	m := NewFileMailer(path, 0, 0)

	err := m.Send(context.Background(), "user@invalid.test", "Judul", "Isi")
	if !errors.Is(err, ErrPermanent) {
		t.Fatalf("expected permanent error, got %v", err)
	}

	if _, statErr := os.Stat(path); statErr == nil {
		t.Fatalf("mail log should not be created for permanent failures")
	}
}

func TestFileMailerTransientFailureReturnsNonPermanent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mail.log")
	m := NewFileMailer(path, 0, 1)

	err := m.Send(context.Background(), "student@example.com", "Judul", "Isi")
	if err == nil {
		t.Fatal("expected transient failure")
	}
	if errors.Is(err, ErrPermanent) {
		t.Fatal("transient failure must not be marked permanent")
	}
}
