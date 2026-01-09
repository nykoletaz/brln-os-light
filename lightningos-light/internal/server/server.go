package server

import (
  "crypto/tls"
  "fmt"
  "log"
  "net/http"
  "sync"
  "time"

  "lightningos-light/internal/config"
  "lightningos-light/internal/lndclient"
)

type Server struct {
  cfg    *config.Config
  logger *log.Logger
  lnd    *lndclient.Client
  lndRestartMu sync.RWMutex
  lastLNDRestart time.Time
}

func New(cfg *config.Config, logger *log.Logger) *Server {
  return &Server{
    cfg:    cfg,
    logger: logger,
    lnd:    lndclient.New(cfg, logger),
  }
}

func (s *Server) Run() error {
  addr := fmt.Sprintf("%s:%d", s.cfg.Server.Host, s.cfg.Server.Port)

  tlsCfg := &tls.Config{
    MinVersion: tls.VersionTLS12,
  }

  httpServer := &http.Server{
    Addr:              addr,
    Handler:           s.routes(),
    ReadHeaderTimeout: 10 * time.Second,
    TLSConfig:         tlsCfg,
  }

  s.logger.Printf("listening on https://%s", addr)
  return httpServer.ListenAndServeTLS(s.cfg.Server.TLSCert, s.cfg.Server.TLSKey)
}
