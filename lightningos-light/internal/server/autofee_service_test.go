package server

import (
  "math"
  "testing"
)

func TestClassifyHTLCFailurePolicy(t *testing.T) {
  entry := htlcManagerFailedEntry{FailureDetail: "AMOUNT BELOW MINIMUM"}
  policy, liquidity := classifyHTLCFailure(entry)
  if !policy {
    t.Fatalf("expected policy=true")
  }
  if liquidity {
    t.Fatalf("expected liquidity=false")
  }
}

func TestClassifyHTLCFailureLiquidity(t *testing.T) {
  entry := htlcManagerFailedEntry{FailureDetail: "TEMPORARY CHANNEL FAILURE"}
  policy, liquidity := classifyHTLCFailure(entry)
  if policy {
    t.Fatalf("expected policy=false")
  }
  if !liquidity {
    t.Fatalf("expected liquidity=true")
  }
}

func TestClassifyHTLCFailureUnknown(t *testing.T) {
  entry := htlcManagerFailedEntry{FailureDetail: "SOMETHING UNKNOWN"}
  policy, liquidity := classifyHTLCFailure(entry)
  if policy || liquidity {
    t.Fatalf("expected policy=false and liquidity=false")
  }
}

func TestForwardRateThresholdScalesWithThresholdFactor(t *testing.T) {
  base := applyHTLCGlobalRateFactor(0.16*htlcForwardSoftRateFactor, htlcGlobalRateFactor, htlcForwardSoftRateFloor)
  if math.Abs(base-0.10) > 0.000001 {
    t.Fatalf("unexpected base threshold: got %.6f want 0.10", base)
  }
  scaled := applyHTLCGlobalRateFactor(base, 1.20, htlcForwardSoftRateFloor)
  if math.Abs(scaled-0.12) > 0.000001 {
    t.Fatalf("unexpected scaled threshold: got %.6f want 0.12", scaled)
  }
}

func TestShouldHoldUpOnRecentRebalance(t *testing.T) {
  if !shouldHoldUpOnRecentRebalance("sink", 0.05, 0.10, 1) {
    t.Fatalf("expected hold-up=true for sink with low out ratio and recent rebalance")
  }
  if shouldHoldUpOnRecentRebalance("sink", 0.20, 0.10, 1) {
    t.Fatalf("expected hold-up=false when out ratio is healthy")
  }
  if shouldHoldUpOnRecentRebalance("router", 0.05, 0.10, 1) {
    t.Fatalf("expected hold-up=false for non-sink channels")
  }
  if shouldHoldUpOnRecentRebalance("sink", 0.05, 0.10, 0) {
    t.Fatalf("expected hold-up=false without recent rebalance")
  }
}
