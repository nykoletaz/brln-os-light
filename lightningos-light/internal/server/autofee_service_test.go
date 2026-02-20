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

func TestBlendTargetWithSeed(t *testing.T) {
  base := 1000
  blended := blendTargetWithSeed(base, 800, 0.20)
  if blended != 960 {
    t.Fatalf("unexpected blend value: got %d want 960", blended)
  }
  if keep := blendTargetWithSeed(base, 0, 0.20); keep != base {
    t.Fatalf("expected base when seed is missing: got %d want %d", keep, base)
  }
  if keep := blendTargetWithSeed(base, 900, 0); keep != base {
    t.Fatalf("expected base when weight is zero: got %d want %d", keep, base)
  }
}

func TestEffectiveLowOutThresholdsDrainedVsFull(t *testing.T) {
  lowDrained, protectDrained, factorDrained := effectiveLowOutThresholds(0.10, 0.10, "drained", 0.20)
  lowFull, protectFull, factorFull := effectiveLowOutThresholds(0.10, 0.10, "full", 0.80)

  if !(lowDrained > 0.10 && protectDrained > 0.10 && factorDrained > 1.0) {
    t.Fatalf("expected drained calibration to increase thresholds: low=%.4f protect=%.4f factor=%.4f", lowDrained, protectDrained, factorDrained)
  }
  if !(lowFull < 0.10 && protectFull < 0.10 && factorFull < 1.0) {
    t.Fatalf("expected full calibration to decrease thresholds: low=%.4f protect=%.4f factor=%.4f", lowFull, protectFull, factorFull)
  }
}

func TestEffectiveLowOutThresholdsUsesRatioGradient(t *testing.T) {
  lowNearBoundary, _, _ := effectiveLowOutThresholds(0.10, 0.10, "drained", 0.24)
  lowVeryDrained, _, _ := effectiveLowOutThresholds(0.10, 0.10, "drained", 0.05)
  if !(lowVeryDrained > lowNearBoundary) {
    t.Fatalf("expected stronger threshold for lower local ratio: near=%.4f very=%.4f", lowNearBoundary, lowVeryDrained)
  }
}

func TestEffectiveLowOutThresholdsFallbackAndClamp(t *testing.T) {
  low, protect, factor := effectiveLowOutThresholds(0, 0, "drained", 0.0)
  if low < lowOutThreshMin || low > lowOutThreshMax {
    t.Fatalf("unexpected low threshold clamp: %.4f", low)
  }
  if protect < lowOutThreshMin || protect > lowOutThreshMax {
    t.Fatalf("unexpected protect threshold clamp: %.4f", protect)
  }
  if factor < lowOutFactorMin || factor > lowOutFactorMax {
    t.Fatalf("unexpected factor clamp: %.4f", factor)
  }
}
