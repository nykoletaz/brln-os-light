package server

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const depixInitRetryCooldown = 10 * time.Second

func (s *Server) initDepix() {
	s.depixMu.Lock()
	defer s.depixMu.Unlock()

	if s.depix != nil && s.depixErr == "" {
		return
	}
	if !s.depixInitAt.IsZero() && time.Since(s.depixInitAt) < depixInitRetryCooldown {
		return
	}
	s.depixInitAt = time.Now()

	dsn, err := ResolveNotificationsDSN(s.logger)
	if err != nil {
		s.depixErr = fmt.Sprintf("depix unavailable: %v", err)
		s.logger.Printf("%s", s.depixErr)
		return
	}

	pool := s.db
	if pool == nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		pool, err = pgxpool.New(ctx, dsn)
		if err != nil {
			s.depixErr = fmt.Sprintf("depix unavailable: failed to connect to postgres: %v", err)
			s.logger.Printf("%s", s.depixErr)
			return
		}
		s.db = pool
	}

	svc := NewDepixService(pool, s.logger)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := svc.EnsureSchema(ctx); err != nil {
		s.depixErr = fmt.Sprintf("depix unavailable: failed to init schema: %v", err)
		s.logger.Printf("%s", s.depixErr)
		return
	}

	s.depix = svc
	s.depixErr = ""
}

func (s *Server) depixService() (*DepixService, string) {
	s.initDepix()
	return s.depix, s.depixErr
}

