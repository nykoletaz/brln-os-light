package reports

import (
  "context"
  "fmt"
  "strings"

  "lightningos-light/internal/lndclient"
  "lightningos-light/lnrpc"
)

const (
  forwardingPageSize = 50000
  paymentsPageSize = 500
)

func ComputeMetrics(ctx context.Context, lnd *lndclient.Client, tr TimeRange, memoMatch bool) (Metrics, error) {
  if lnd == nil {
    return Metrics{}, fmt.Errorf("lnd client unavailable")
  }
  forwardRevenue, forwardCount, routedVolume, err := fetchForwardingMetrics(ctx, lnd, tr.StartUnix(), tr.EndUnixInclusive())
  if err != nil {
    return Metrics{}, err
  }

  pubkey, err := fetchNodePubkey(ctx, lnd)
  if err != nil {
    return Metrics{}, err
  }

  rebalanceCost, rebalanceCount, err := fetchRebalanceMetrics(ctx, lnd, tr.StartUnix(), tr.EndUnixInclusive(), pubkey, memoMatch)
  if err != nil {
    return Metrics{}, err
  }

  metrics := Metrics{
    ForwardFeeRevenueSat: forwardRevenue,
    RebalanceFeeCostSat: rebalanceCost,
    NetRoutingProfitSat: forwardRevenue - rebalanceCost,
    ForwardCount: forwardCount,
    RebalanceCount: rebalanceCount,
    RoutedVolumeSat: routedVolume,
  }
  return metrics, nil
}

func fetchForwardingMetrics(ctx context.Context, lnd *lndclient.Client, startUnix uint64, endUnix uint64) (int64, int64, int64, error) {
  conn, err := lnd.DialLightning(ctx)
  if err != nil {
    return 0, 0, 0, err
  }
  defer conn.Close()

  client := lnrpc.NewLightningClient(conn)

  var offset uint32
  var revenue int64
  var routedVolume int64
  var count int64

  for {
    resp, err := client.ForwardingHistory(ctx, &lnrpc.ForwardingHistoryRequest{
      StartTime: startUnix,
      EndTime: endUnix,
      IndexOffset: offset,
      NumMaxEvents: forwardingPageSize,
    })
    if err != nil {
      return 0, 0, 0, err
    }
    if resp == nil || len(resp.ForwardingEvents) == 0 {
      break
    }

    for _, evt := range resp.ForwardingEvents {
      if evt == nil {
        continue
      }
      revenue += extractForwardFeeSat(evt)
      routedVolume += extractForwardAmountSat(evt)
      count++
    }

    if resp.LastOffsetIndex <= offset {
      break
    }
    offset = resp.LastOffsetIndex
    if len(resp.ForwardingEvents) < forwardingPageSize {
      break
    }
  }

  return revenue, count, routedVolume, nil
}

func fetchRebalanceMetrics(ctx context.Context, lnd *lndclient.Client, startUnix uint64, endUnix uint64, ourPubkey string, memoMatch bool) (int64, int64, error) {
  conn, err := lnd.DialLightning(ctx)
  if err != nil {
    return 0, 0, err
  }
  defer conn.Close()

  client := lnrpc.NewLightningClient(conn)
  decodeCache := map[string]decodedPayReq{}

  var offset uint64
  var totalFeeMsat int64
  var rebalanceCount int64

  for {
    req := &lnrpc.ListPaymentsRequest{
      IncludeIncomplete: false,
      IndexOffset: offset,
      MaxPayments: paymentsPageSize,
      CreationDateStart: startUnix,
      CreationDateEnd: endUnix,
    }

    resp, err := client.ListPayments(ctx, req)
    if err != nil {
      return 0, 0, err
    }
    if resp == nil || len(resp.Payments) == 0 {
      break
    }

    for _, pay := range resp.Payments {
      if pay == nil {
        continue
      }
      timestamp := extractPaymentTimestamp(pay)
      if timestamp < int64(startUnix) || timestamp > int64(endUnix) {
        continue
      }
      if !PaymentSucceeded(pay) {
        continue
      }

      dest, description := extractDestinationAndDescription(ctx, lnd, pay, decodeCache)
      if IsRebalancePayment(pay, ourPubkey, dest, description, memoMatch) {
        feeMsat := extractPaymentFeeMsat(pay)
        totalFeeMsat += feeMsat
        rebalanceCount++
      }
    }

    nextOffset := resp.LastIndexOffset
    if nextOffset <= offset {
      break
    }
    offset = nextOffset
    if len(resp.Payments) < paymentsPageSize {
      break
    }
  }

  return totalFeeMsat / 1000, rebalanceCount, nil
}

func fetchNodePubkey(ctx context.Context, lnd *lndclient.Client) (string, error) {
  cached := strings.TrimSpace(lnd.CachedPubkey())
  if cached != "" {
    return cached, nil
  }
  status, err := lnd.GetStatus(ctx)
  if err != nil {
    return "", err
  }
  if strings.TrimSpace(status.Pubkey) == "" {
    return "", fmt.Errorf("lnd pubkey unavailable")
  }
  return status.Pubkey, nil
}

type decodedPayReq struct {
  Destination string
  Description string
  Ready bool
}

func extractDestinationAndDescription(ctx context.Context, lnd *lndclient.Client, pay *lnrpc.Payment, cache map[string]decodedPayReq) (string, string) {
  if pay == nil {
    return "", ""
  }

  dest := ""
  description := ""

  for _, htlc := range pay.Htlcs {
    if htlc == nil || htlc.Route == nil || len(htlc.Route.Hops) == 0 {
      continue
    }
    if htlc.Status == lnrpc.HTLCAttempt_SUCCEEDED {
      dest = htlc.Route.Hops[len(htlc.Route.Hops)-1].PubKey
      break
    }
  }

  payreq := strings.TrimSpace(pay.PaymentRequest)
  if payreq == "" {
    return dest, description
  }

  if cached, ok := cache[payreq]; ok && cached.Ready {
    if dest == "" {
      dest = cached.Destination
    }
    description = cached.Description
    return dest, description
  }

  decoded, err := lnd.DecodeInvoice(ctx, payreq)
  if err != nil {
    cache[payreq] = decodedPayReq{Ready: true}
    return dest, description
  }

  entry := decodedPayReq{
    Destination: decoded.Destination,
    Description: decoded.Memo,
    Ready: true,
  }
  cache[payreq] = entry
  if dest == "" {
    dest = entry.Destination
  }
  description = entry.Description
  return dest, description
}

func extractForwardFeeSat(evt *lnrpc.ForwardingEvent) int64 {
  if evt == nil {
    return 0
  }
  if evt.Fee != 0 {
    return int64(evt.Fee)
  }
  if evt.FeeMsat != 0 {
    return int64(evt.FeeMsat / 1000)
  }
  return 0
}

func extractForwardAmountSat(evt *lnrpc.ForwardingEvent) int64 {
  if evt == nil {
    return 0
  }
  if evt.AmtOut != 0 {
    return int64(evt.AmtOut)
  }
  if evt.AmtOutMsat != 0 {
    return int64(evt.AmtOutMsat / 1000)
  }
  return 0
}

func extractPaymentTimestamp(pay *lnrpc.Payment) int64 {
  if pay == nil {
    return 0
  }
  if pay.CreationDate != 0 {
    return int64(pay.CreationDate)
  }
  if pay.CreationTimeNs != 0 {
    return int64(pay.CreationTimeNs / 1_000_000_000)
  }
  return 0
}

func extractPaymentFeeMsat(pay *lnrpc.Payment) int64 {
  if pay == nil {
    return 0
  }
  if pay.FeeMsat != 0 {
    return int64(pay.FeeMsat)
  }
  if pay.FeeSat != 0 {
    return int64(pay.FeeSat) * 1000
  }
  return 0
}
