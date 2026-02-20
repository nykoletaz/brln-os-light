package server

import (
  "context"
  "fmt"
  "time"

  "github.com/jackc/pgx/v5/pgxpool"
)

const shortcutsInitRetryCooldown = 10 * time.Second

func (s *Server) initShortcuts() {
  s.shortcutsMu.Lock()
  defer s.shortcutsMu.Unlock()

  if s.shortcuts != nil && s.shortcutsErr == "" {
    return
  }
  if !s.shortcutsInitAt.IsZero() && time.Since(s.shortcutsInitAt) < shortcutsInitRetryCooldown {
    return
  }
  s.shortcutsInitAt = time.Now()

  dsn, err := ResolveNotificationsDSN(s.logger)
  if err != nil {
    s.shortcutsErr = fmt.Sprintf("shortcuts unavailable: %v", err)
    s.logger.Printf("%s", s.shortcutsErr)
    return
  }

  pool := s.db
  if pool == nil {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    pool, err = pgxpool.New(ctx, dsn)
    if err != nil {
      s.shortcutsErr = fmt.Sprintf("shortcuts unavailable: failed to connect to postgres: %v", err)
      s.logger.Printf("%s", s.shortcutsErr)
      return
    }
    s.db = pool
  }

  svc := NewShortcutsService(pool, s.logger)
  ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
  defer cancel()
  if err := svc.EnsureSchema(ctx); err != nil {
    s.shortcutsErr = fmt.Sprintf("shortcuts unavailable: failed to init schema: %v", err)
    s.logger.Printf("%s", s.shortcutsErr)
    return
  }

  s.shortcuts = svc
  s.shortcutsErr = ""
}

func (s *Server) shortcutsService() (*ShortcutsService, string) {
  s.initShortcuts()
  return s.shortcuts, s.shortcutsErr
}
