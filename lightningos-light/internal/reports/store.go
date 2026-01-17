package reports

import (
  "context"
  "time"

  "github.com/jackc/pgx/v5/pgtype"
  "github.com/jackc/pgx/v5/pgxpool"
)

func EnsureSchema(ctx context.Context, db *pgxpool.Pool) error {
  if db == nil {
    return nil
  }
  _, err := db.Exec(ctx, `
create table if not exists reports_daily (
  report_date date primary key,
  forward_fee_revenue_sats bigint not null default 0,
  rebalance_fee_cost_sats bigint not null default 0,
  net_routing_profit_sats bigint not null default 0,
  forward_count integer not null default 0,
  rebalance_count integer not null default 0,
  routed_volume_sats bigint not null default 0,
  onchain_balance_sats bigint null,
  lightning_balance_sats bigint null,
  total_balance_sats bigint null,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);
`)
  return err
}

func UpsertDaily(ctx context.Context, db *pgxpool.Pool, row Row) error {
  if db == nil {
    return nil
  }
  query, args := buildUpsertDaily(row)
  _, err := db.Exec(ctx, query, args...)
  return err
}

func buildUpsertDaily(row Row) (string, []any) {
  reportDate := normalizeReportDate(row.ReportDate)
  metrics := row.Metrics

  args := []any{
    reportDate,
    metrics.ForwardFeeRevenueSat,
    metrics.RebalanceFeeCostSat,
    metrics.NetRoutingProfitSat,
    metrics.ForwardCount,
    metrics.RebalanceCount,
    metrics.RoutedVolumeSat,
    nullableInt64(metrics.OnchainBalanceSat),
    nullableInt64(metrics.LightningBalanceSat),
    nullableInt64(metrics.TotalBalanceSat),
  }

  query := `
insert into reports_daily (
  report_date,
  forward_fee_revenue_sats,
  rebalance_fee_cost_sats,
  net_routing_profit_sats,
  forward_count,
  rebalance_count,
  routed_volume_sats,
  onchain_balance_sats,
  lightning_balance_sats,
  total_balance_sats
) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
on conflict (report_date) do update set
  forward_fee_revenue_sats = excluded.forward_fee_revenue_sats,
  rebalance_fee_cost_sats = excluded.rebalance_fee_cost_sats,
  net_routing_profit_sats = excluded.net_routing_profit_sats,
  forward_count = excluded.forward_count,
  rebalance_count = excluded.rebalance_count,
  routed_volume_sats = excluded.routed_volume_sats,
  onchain_balance_sats = excluded.onchain_balance_sats,
  lightning_balance_sats = excluded.lightning_balance_sats,
  total_balance_sats = excluded.total_balance_sats,
  updated_at = now()
`

  return query, args
}

func FetchRange(ctx context.Context, db *pgxpool.Pool, startDate, endDate time.Time) ([]Row, error) {
  if db == nil {
    return nil, nil
  }
  rows, err := db.Query(ctx, `
select report_date,
  forward_fee_revenue_sats,
  rebalance_fee_cost_sats,
  net_routing_profit_sats,
  forward_count,
  rebalance_count,
  routed_volume_sats,
  onchain_balance_sats,
  lightning_balance_sats,
  total_balance_sats
from reports_daily
where report_date >= $1 and report_date <= $2
order by report_date asc
`, normalizeReportDate(startDate), normalizeReportDate(endDate))
  if err != nil {
    return nil, err
  }
  defer rows.Close()

  var items []Row
  for rows.Next() {
    row, err := scanRow(rows)
    if err != nil {
      return nil, err
    }
    items = append(items, row)
  }
  return items, rows.Err()
}

func FetchAll(ctx context.Context, db *pgxpool.Pool) ([]Row, error) {
  if db == nil {
    return nil, nil
  }
  rows, err := db.Query(ctx, `
select report_date,
  forward_fee_revenue_sats,
  rebalance_fee_cost_sats,
  net_routing_profit_sats,
  forward_count,
  rebalance_count,
  routed_volume_sats,
  onchain_balance_sats,
  lightning_balance_sats,
  total_balance_sats
from reports_daily
order by report_date asc
`)
  if err != nil {
    return nil, err
  }
  defer rows.Close()

  var items []Row
  for rows.Next() {
    row, err := scanRow(rows)
    if err != nil {
      return nil, err
    }
    items = append(items, row)
  }
  return items, rows.Err()
}

func FetchSummaryRange(ctx context.Context, db *pgxpool.Pool, startDate, endDate time.Time) (Summary, error) {
  if db == nil {
    return Summary{}, nil
  }
  var days int64
  totals := Metrics{}
  err := db.QueryRow(ctx, `
select
  count(*),
  coalesce(sum(forward_fee_revenue_sats), 0),
  coalesce(sum(rebalance_fee_cost_sats), 0),
  coalesce(sum(net_routing_profit_sats), 0),
  coalesce(sum(forward_count), 0),
  coalesce(sum(rebalance_count), 0),
  coalesce(sum(routed_volume_sats), 0)
from reports_daily
where report_date >= $1 and report_date <= $2
`, normalizeReportDate(startDate), normalizeReportDate(endDate)).Scan(
    &days,
    &totals.ForwardFeeRevenueSat,
    &totals.RebalanceFeeCostSat,
    &totals.NetRoutingProfitSat,
    &totals.ForwardCount,
    &totals.RebalanceCount,
    &totals.RoutedVolumeSat,
  )
  if err != nil {
    return Summary{}, err
  }

  return Summary{Days: days, Totals: totals, Averages: averageMetrics(totals, days)}, nil
}

func FetchSummaryAll(ctx context.Context, db *pgxpool.Pool) (Summary, error) {
  if db == nil {
    return Summary{}, nil
  }
  var days int64
  totals := Metrics{}
  err := db.QueryRow(ctx, `
select
  count(*),
  coalesce(sum(forward_fee_revenue_sats), 0),
  coalesce(sum(rebalance_fee_cost_sats), 0),
  coalesce(sum(net_routing_profit_sats), 0),
  coalesce(sum(forward_count), 0),
  coalesce(sum(rebalance_count), 0),
  coalesce(sum(routed_volume_sats), 0)
from reports_daily
`).Scan(
    &days,
    &totals.ForwardFeeRevenueSat,
    &totals.RebalanceFeeCostSat,
    &totals.NetRoutingProfitSat,
    &totals.ForwardCount,
    &totals.RebalanceCount,
    &totals.RoutedVolumeSat,
  )
  if err != nil {
    return Summary{}, err
  }

  return Summary{Days: days, Totals: totals, Averages: averageMetrics(totals, days)}, nil
}

func averageMetrics(totals Metrics, days int64) Metrics {
  if days <= 0 {
    return Metrics{}
  }
  return Metrics{
    ForwardFeeRevenueSat: totals.ForwardFeeRevenueSat / days,
    RebalanceFeeCostSat: totals.RebalanceFeeCostSat / days,
    NetRoutingProfitSat: totals.NetRoutingProfitSat / days,
    ForwardCount: totals.ForwardCount / days,
    RebalanceCount: totals.RebalanceCount / days,
    RoutedVolumeSat: totals.RoutedVolumeSat / days,
  }
}

type rowScanner interface {
  Scan(dest ...any) error
}

func scanRow(scanner rowScanner) (Row, error) {
  var reportDate time.Time
  var metrics Metrics
  var onchain pgtype.Int8
  var lightning pgtype.Int8
  var total pgtype.Int8
  err := scanner.Scan(
    &reportDate,
    &metrics.ForwardFeeRevenueSat,
    &metrics.RebalanceFeeCostSat,
    &metrics.NetRoutingProfitSat,
    &metrics.ForwardCount,
    &metrics.RebalanceCount,
    &metrics.RoutedVolumeSat,
    &onchain,
    &lightning,
    &total,
  )
  if err != nil {
    return Row{}, err
  }
  if onchain.Valid {
    val := onchain.Int64
    metrics.OnchainBalanceSat = &val
  }
  if lightning.Valid {
    val := lightning.Int64
    metrics.LightningBalanceSat = &val
  }
  if total.Valid {
    val := total.Int64
    metrics.TotalBalanceSat = &val
  }
  return Row{ReportDate: reportDate, Metrics: metrics}, nil
}

func nullableInt64(value *int64) any {
  if value == nil {
    return nil
  }
  return *value
}

func normalizeReportDate(value time.Time) time.Time {
  return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
}
