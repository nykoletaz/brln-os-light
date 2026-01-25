package reports

import (
  "context"
  "fmt"
  "time"

  "lightningos-light/internal/lndclient"
  "lightningos-light/lnrpc"
)

const rebalanceScanPageSize = 5000
const rebalanceScanMaxPages = 200000

func FetchRebalanceFeesByDay(ctx context.Context, lnd *lndclient.Client, startUnix uint64, endUnix uint64, loc *time.Location) (map[time.Time]RebalanceOverride, error) {
  if lnd == nil {
    return nil, fmt.Errorf("lnd client unavailable")
  }
  if loc == nil {
    loc = time.Local
  }

  pubkey, err := fetchNodePubkey(ctx, lnd)
  if err != nil {
    return nil, err
  }

  conn, err := lnd.DialLightning(ctx)
  if err != nil {
    return nil, err
  }
  defer conn.Close()

  client := lnrpc.NewLightningClient(conn)
  results := make(map[time.Time]RebalanceOverride)

  var indexOffset uint64
  var pages int
  var lastOffset uint64

  for {
    if pages > rebalanceScanMaxPages {
      break
    }
    pages++

    req := &lnrpc.ListPaymentsRequest{
      IncludeIncomplete: false,
      Reversed: true,
      IndexOffset: indexOffset,
      MaxPayments: rebalanceScanPageSize,
    }
    resp, err := client.ListPayments(ctx, req)
    if err != nil {
      return nil, err
    }
    if resp == nil || len(resp.Payments) == 0 {
      break
    }

    minIndex := uint64(0)
    maxTs := int64(0)
    minTs := int64(1<<63 - 1)

    for _, pay := range resp.Payments {
      if pay == nil {
        continue
      }
      if pay.PaymentIndex > 0 {
        if minIndex == 0 || pay.PaymentIndex < minIndex {
          minIndex = pay.PaymentIndex
        }
      }

      ts := extractPaymentTimestamp(pay)
      if ts > maxTs {
        maxTs = ts
      }
      if ts < minTs {
        minTs = ts
      }

      if ts < int64(startUnix) || ts > int64(endUnix) {
        continue
      }
      if !PaymentSucceeded(pay) {
        continue
      }
      if !IsRebalancePayment(pay, pubkey, "", "", false) {
        continue
      }

      feeMsat := extractPaymentFeeMsat(pay)
      local := time.Unix(ts, 0).In(loc)
      dayKey := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, loc)
      current := results[dayKey]
      current.FeeMsat += feeMsat
      current.Count++
      results[dayKey] = current
    }

    if maxTs < int64(startUnix) {
      break
    }
    if minIndex == 0 {
      break
    }
    if indexOffset != 0 && minIndex >= indexOffset {
      break
    }
    if lastOffset == minIndex {
      break
    }
    lastOffset = minIndex
    indexOffset = minIndex

    if len(resp.Payments) < rebalanceScanPageSize && minTs < int64(startUnix) {
      break
    }
  }

  return results, nil
}
