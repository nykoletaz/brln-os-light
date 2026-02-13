package server

import "testing"

func TestWithinMaxHTLCHysteresis(t *testing.T) {
  tests := []struct {
    name        string
    currentSat  int64
    targetSat   int64
    capacitySat int64
    want        bool
  }{
    {
      name:        "small delta under threshold",
      currentSat:  981511,
      targetSat:   981557,
      capacitySat: 1000000,
      want:        true,
    },
    {
      name:        "large delta over threshold",
      currentSat:  981000,
      targetSat:   983000,
      capacitySat: 1000000,
      want:        false,
    },
    {
      name:        "tiny channel still uses base threshold",
      currentSat:  5000,
      targetSat:   5090,
      capacitySat: 10000,
      want:        true,
    },
    {
      name:        "tiny channel over base threshold",
      currentSat:  5000,
      targetSat:   5200,
      capacitySat: 10000,
      want:        false,
    },
  }

  for _, tc := range tests {
    tc := tc
    t.Run(tc.name, func(t *testing.T) {
      currentMsat, err := satToMsat(tc.currentSat)
      if err != nil {
        t.Fatalf("satToMsat(currentSat) failed: %v", err)
      }
      targetMsat, err := satToMsat(tc.targetSat)
      if err != nil {
        t.Fatalf("satToMsat(targetSat) failed: %v", err)
      }
      if got := withinMaxHTLCHysteresis(currentMsat, targetMsat, tc.capacitySat); got != tc.want {
        t.Fatalf("withinMaxHTLCHysteresis(%d, %d, %d) = %v, want %v", currentMsat, targetMsat, tc.capacitySat, got, tc.want)
      }
    })
  }
}

func TestMinutesToHoursCeil(t *testing.T) {
  tests := []struct {
    in   int
    want int
  }{
    {in: 0, want: 0},
    {in: 1, want: 1},
    {in: 59, want: 1},
    {in: 60, want: 1},
    {in: 61, want: 2},
    {in: 240, want: 4},
  }
  for _, tc := range tests {
    if got := minutesToHoursCeil(tc.in); got != tc.want {
      t.Fatalf("minutesToHoursCeil(%d) = %d, want %d", tc.in, got, tc.want)
    }
  }
}

func TestValidateHTLCManagerConfigIntervalMinutes(t *testing.T) {
  cfg := defaultHTLCManagerConfig()
  cfg.IntervalMinutes = 30
  if err := validateHTLCManagerConfig(cfg); err != nil {
    t.Fatalf("expected valid config for 30 minutes: %v", err)
  }

  cfg.IntervalMinutes = 0
  if err := validateHTLCManagerConfig(cfg); err == nil {
    t.Fatalf("expected validation error for zero minutes")
  }
}
