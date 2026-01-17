package reports

import (
  "strings"

  "lightningos-light/lnrpc"
)

type RouteEndpoints struct {
  FirstChanID uint64
  LastChanID uint64
  LastHopPubkey string
  HopCount int
}

func ExtractRouteEndpoints(pay *lnrpc.Payment) RouteEndpoints {
  if pay == nil {
    return RouteEndpoints{}
  }
  for _, attempt := range pay.Htlcs {
    if attempt == nil || attempt.Route == nil {
      continue
    }
    if attempt.Status != lnrpc.HTLCAttempt_SUCCEEDED {
      continue
    }
    hops := attempt.Route.Hops
    if len(hops) < 2 {
      return RouteEndpoints{}
    }
    return RouteEndpoints{
      FirstChanID: hops[0].ChanId,
      LastChanID: hops[len(hops)-1].ChanId,
      LastHopPubkey: hops[len(hops)-1].PubKey,
      HopCount: len(hops),
    }
  }
  return RouteEndpoints{}
}

func PaymentSucceeded(pay *lnrpc.Payment) bool {
  if pay == nil {
    return false
  }
  return pay.Status == lnrpc.Payment_SUCCEEDED
}

func IsRebalancePayment(pay *lnrpc.Payment, ourPubkey string, destPubkey string, description string, memoMatch bool) bool {
  if pay == nil {
    return false
  }
  if !PaymentSucceeded(pay) {
    return false
  }

  endpoints := ExtractRouteEndpoints(pay)
  dest := endpoints.LastHopPubkey
  if dest == "" {
    dest = destPubkey
  }

  isRebalance := endpoints.HopCount >= 2 && dest != "" && strings.EqualFold(dest, ourPubkey) && endpoints.FirstChanID != 0 && endpoints.LastChanID != 0 && endpoints.FirstChanID != endpoints.LastChanID
  if !isRebalance && memoMatch {
    if strings.Contains(strings.ToLower(description), "rebalance") {
      isRebalance = true
    }
  }
  return isRebalance
}
