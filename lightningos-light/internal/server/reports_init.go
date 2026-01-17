package server

import (
  "context"
  "fmt"
  "time"

  "lightningos-light/internal/reports"

  "github.com/jackc/pgx/v5/pgxpool"
)

func (s *Server) initReports() {
  s.reportsOnce.Do(func() {
    dsn, err := ResolveNotificationsDSN(s.logger)
    if err != nil {
      s.reportsErr = fmt.Sprintf("reports unavailable: %v", err)
      s.logger.Printf("%s", s.reportsErr)
      return
    }

    pool := s.db
    if pool == nil {
      ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
      defer cancel()
      pool, err = pgxpool.New(ctx, dsn)
      if err != nil {
        s.reportsErr = fmt.Sprintf("reports unavailable: failed to connect to postgres: %v", err)
        s.logger.Printf("%s", s.reportsErr)
        return
      }
      s.db = pool
    }

    svc := reports.NewService(pool, s.lnd, s.logger)
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    if err := svc.EnsureSchema(ctx); err != nil {
      s.reportsErr = fmt.Sprintf("reports unavailable: failed to init schema: %v", err)
      s.logger.Printf("%s", s.reportsErr)
      return
    }

    s.reports = svc
    s.reportsErr = ""
  })
}

func (s *Server) reportsService() (*reports.Service, string) {
  s.initReports()
  return s.reports, s.reportsErr
}
