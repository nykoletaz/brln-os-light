package reports

import (
  "testing"

  "lightningos-light/lnrpc"
)

func TestIsRebalancePaymentStrict(t *testing.T) {
  pay := &lnrpc.Payment{
    Status: lnrpc.Payment_SUCCEEDED,
    Htlcs: []*lnrpc.HTLCAttempt{
      {
        Status: lnrpc.HTLCAttempt_SUCCEEDED,
        Route: &lnrpc.Route{
          Hops: []*lnrpc.Hop{
            {ChanId: 111, PubKey: "peer-out"},
            {ChanId: 222, PubKey: "our-node"},
          },
        },
      },
    },
  }

  if !IsRebalancePayment(pay, "our-node", "", "", false) {
    t.Fatalf("expected rebalance to match")
  }
}

func TestIsRebalancePaymentSameChannel(t *testing.T) {
  pay := &lnrpc.Payment{
    Status: lnrpc.Payment_SUCCEEDED,
    Htlcs: []*lnrpc.HTLCAttempt{
      {
        Status: lnrpc.HTLCAttempt_SUCCEEDED,
        Route: &lnrpc.Route{
          Hops: []*lnrpc.Hop{
            {ChanId: 111, PubKey: "peer-out"},
            {ChanId: 111, PubKey: "our-node"},
          },
        },
      },
    },
  }

  if IsRebalancePayment(pay, "our-node", "", "", false) {
    t.Fatalf("expected rebalance to be false when channel IDs match")
  }
}

func TestIsRebalancePaymentMemoMatch(t *testing.T) {
  pay := &lnrpc.Payment{Status: lnrpc.Payment_SUCCEEDED}
  if !IsRebalancePayment(pay, "our-node", "", "rebalance to self", true) {
    t.Fatalf("expected memo match rebalance")
  }
}
