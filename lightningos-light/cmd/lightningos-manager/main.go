package main

import (
  "context"
  "flag"
  "log"
  "os"
  "strings"
  "time"

  "lightningos-light/internal/config"
  "lightningos-light/internal/lndclient"
  "lightningos-light/internal/reports"
  "lightningos-light/internal/server"

  "github.com/jackc/pgx/v5/pgxpool"
)

func main() {
  if len(os.Args) > 1 && os.Args[1] == "reports-run" {
    runReports(os.Args[2:])
    return
  }

  runServer(os.Args[1:])
}

func runServer(args []string) {
  fs := flag.NewFlagSet("lightningos-manager", flag.ExitOnError)
  configPath := fs.String("config", "/etc/lightningos/config.yaml", "Path to config.yaml")
  _ = fs.Parse(args)

  cfg, err := config.Load(*configPath)
  if err != nil {
    log.Fatalf("config load failed: %v", err)
  }

  logger := log.New(os.Stdout, "", log.LstdFlags)
  srv := server.New(cfg, logger)

  if err := srv.Run(); err != nil {
    logger.Fatalf("server exited: %v", err)
  }
}

func runReports(args []string) {
  fs := flag.NewFlagSet("reports-run", flag.ExitOnError)
  configPath := fs.String("config", "/etc/lightningos/config.yaml", "Path to config.yaml")
  dateStr := fs.String("date", "", "Report date (YYYY-MM-DD), defaults to yesterday")
  _ = fs.Parse(args)

  cfg, err := config.Load(*configPath)
  if err != nil {
    log.Fatalf("config load failed: %v", err)
  }

  logger := log.New(os.Stdout, "", log.LstdFlags)
  dsn, err := server.ResolveNotificationsDSN(logger)
  if err != nil {
    logger.Fatalf("reports-run failed: %v", err)
  }

  ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
  defer cancel()

  pool, err := pgxpool.New(ctx, dsn)
  if err != nil {
    logger.Fatalf("reports-run failed: %v", err)
  }
  defer pool.Close()

  svc := reports.NewService(pool, lndclient.New(cfg, logger), logger)
  if err := svc.EnsureSchema(ctx); err != nil {
    logger.Fatalf("reports-run failed: %v", err)
  }

  loc := time.Local
  reportDate := time.Now().In(loc).AddDate(0, 0, -1)
  if strings.TrimSpace(*dateStr) != "" {
    parsed, err := reports.ParseDate(*dateStr, loc)
    if err != nil {
      logger.Fatalf("reports-run failed: invalid date")
    }
    reportDate = parsed
  }

  row, err := svc.RunDaily(ctx, reportDate, loc)
  if err != nil {
    logger.Fatalf("reports-run failed: %v", err)
  }

  logger.Printf(
    "reports: stored %s (revenue %d sats, cost %d sats, net %d sats)",
    row.ReportDate.Format("2006-01-02"),
    row.Metrics.ForwardFeeRevenueSat,
    row.Metrics.RebalanceFeeCostSat,
    row.Metrics.NetRoutingProfitSat,
  )
}
