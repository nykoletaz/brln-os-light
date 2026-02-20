package server

import (
  "errors"
  "testing"
)

func TestNormalizeShortcutURL(t *testing.T) {
  t.Run("adds https when missing", func(t *testing.T) {
    got, err := normalizeShortcutURL("wallet.br-ln.com")
    if err != nil {
      t.Fatalf("unexpected error: %v", err)
    }
    if got != "https://wallet.br-ln.com" {
      t.Fatalf("unexpected url: %s", got)
    }
  })

  t.Run("rejects non http schemes", func(t *testing.T) {
    _, err := normalizeShortcutURL("javascript:alert(1)")
    if !errors.Is(err, ErrShortcutInvalidURL) {
      t.Fatalf("expected ErrShortcutInvalidURL, got %v", err)
    }
  })
}

func TestNormalizeShortcutEmoji(t *testing.T) {
	t.Run("accepts emoji", func(t *testing.T) {
		got, err := normalizeShortcutEmoji("\U0001F680")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "\U0001F680" {
			t.Fatalf("unexpected emoji: %s", got)
		}
	})

  t.Run("rejects empty", func(t *testing.T) {
    _, err := normalizeShortcutEmoji("   ")
    if !errors.Is(err, ErrShortcutInvalidEmoji) {
      t.Fatalf("expected ErrShortcutInvalidEmoji, got %v", err)
    }
  })
}

func TestShortcutNameFromURL(t *testing.T) {
  got := shortcutNameFromURL("https://www.services.br-ln.com/path")
  if got != "services.br-ln.com" {
    t.Fatalf("unexpected name: %s", got)
  }
}
