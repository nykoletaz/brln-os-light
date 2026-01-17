package server

import (
  "errors"
  "log"
  "os"
  "strings"
)

func ResolveNotificationsDSN(logger *log.Logger) (string, error) {
  dsn := os.Getenv("NOTIFICATIONS_PG_DSN")
  if isPlaceholderDSN(dsn) {
    dsn = ""
  }
  if strings.TrimSpace(dsn) == "" {
    bootstrapped, err := bootstrapNotificationsDSN(logger)
    if err != nil {
      return "", err
    }
    dsn = bootstrapped
  }
  if strings.TrimSpace(dsn) == "" {
    return "", errors.New("NOTIFICATIONS_PG_DSN not set")
  }
  return dsn, nil
}
