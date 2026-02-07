
package server

import (
  "bytes"
  "context"
  "encoding/json"
  "errors"
  "fmt"
  "math"
  "math/rand"
  "net/http"
  "sort"
  "strconv"
  "strings"
  "sync"
  "time"

  "lightningos-light/internal/lndclient"

  "github.com/jackc/pgx/v5"
  "github.com/jackc/pgx/v5/pgtype"
  "github.com/jackc/pgx/v5/pgxpool"
)

const (
  autofeeConfigID = 1
  autofeeMinLookbackDays = 5
  autofeeMaxLookbackDays = 21
)

const superSourceBaseFeeMsatDefault = 1000

type AutofeeConfig struct {
  Enabled bool `json:"enabled"`
  Profile string `json:"profile"`
  LookbackDays int `json:"lookback_days"`
  RunIntervalSec int `json:"run_interval_sec"`
  CooldownUpSec int `json:"cooldown_up_sec"`
  CooldownDownSec int `json:"cooldown_down_sec"`
  AmbossEnabled bool `json:"amboss_enabled"`
  AmbossTokenSet bool `json:"amboss_token_set"`
  InboundPassiveEnabled bool `json:"inbound_passive_enabled"`
  DiscoveryEnabled bool `json:"discovery_enabled"`
  ExplorerEnabled bool `json:"explorer_enabled"`
  SuperSourceEnabled bool `json:"super_source_enabled"`
  SuperSourceBaseFeeMsat int `json:"super_source_base_fee_msat"`
  MinPpm int `json:"min_ppm"`
  MaxPpm int `json:"max_ppm"`
}

type AutofeeConfigUpdate struct {
  Enabled *bool `json:"enabled,omitempty"`
  Profile *string `json:"profile,omitempty"`
  LookbackDays *int `json:"lookback_days,omitempty"`
  RunIntervalSec *int `json:"run_interval_sec,omitempty"`
  CooldownUpSec *int `json:"cooldown_up_sec,omitempty"`
  CooldownDownSec *int `json:"cooldown_down_sec,omitempty"`
  AmbossEnabled *bool `json:"amboss_enabled,omitempty"`
  AmbossToken *string `json:"amboss_token,omitempty"`
  InboundPassiveEnabled *bool `json:"inbound_passive_enabled,omitempty"`
  DiscoveryEnabled *bool `json:"discovery_enabled,omitempty"`
  ExplorerEnabled *bool `json:"explorer_enabled,omitempty"`
  SuperSourceEnabled *bool `json:"super_source_enabled,omitempty"`
  SuperSourceBaseFeeMsat *int `json:"super_source_base_fee_msat,omitempty"`
  MinPpm *int `json:"min_ppm,omitempty"`
  MaxPpm *int `json:"max_ppm,omitempty"`
}

type AutofeeStatus struct {
  Running bool `json:"running"`
  LastRunAt string `json:"last_run_at,omitempty"`
  NextRunAt string `json:"next_run_at,omitempty"`
  LastError string `json:"last_error,omitempty"`
}

type autofeeProfile struct {
  Name string
  StepCap float64
  LowOutThresh float64
  HighOutThresh float64
  SurgeBumpMax float64
  RunIntervalSec int
  CooldownUpSec int
  CooldownDownSec int
  DiscoveryStepCapDown float64
}

type superSourceThresholds struct {
  OutRatioMin float64
  OutAmt1dMult float64
  OutAmt7dMult float64
  MinFwds7d int
  EnterHours int
  ExitHours int
}

var autofeeProfiles = map[string]autofeeProfile{
  "conservative": {
    Name: "conservative",
    StepCap: 0.03,
    LowOutThresh: 0.08,
    HighOutThresh: 0.25,
    SurgeBumpMax: 0.10,
    RunIntervalSec: 8 * 3600,
    CooldownUpSec: 6 * 3600,
    CooldownDownSec: 8 * 3600,
    DiscoveryStepCapDown: 0.10,
  },
  "moderate": {
    Name: "moderate",
    StepCap: 0.05,
    LowOutThresh: 0.10,
    HighOutThresh: 0.20,
    SurgeBumpMax: 0.20,
    RunIntervalSec: 4 * 3600,
    CooldownUpSec: 3 * 3600,
    CooldownDownSec: 4 * 3600,
    DiscoveryStepCapDown: 0.15,
  },
  "aggressive": {
    Name: "aggressive",
    StepCap: 0.08,
    LowOutThresh: 0.12,
    HighOutThresh: 0.18,
    SurgeBumpMax: 0.30,
    RunIntervalSec: 2 * 3600,
    CooldownUpSec: 1 * 3600,
    CooldownDownSec: 2 * 3600,
    DiscoveryStepCapDown: 0.20,
  },
}

var superSourceThresholdsByProfile = map[string]superSourceThresholds{
  "conservative": {
    OutRatioMin: 0.65,
    OutAmt1dMult: 0.70,
    OutAmt7dMult: 5.0,
    MinFwds7d: 15,
    EnterHours: 6,
    ExitHours: 96,
  },
  "moderate": {
    OutRatioMin: 0.60,
    OutAmt1dMult: 0.50,
    OutAmt7dMult: 4.0,
    MinFwds7d: 10,
    EnterHours: 0,
    ExitHours: 72,
  },
  "aggressive": {
    OutRatioMin: 0.55,
    OutAmt1dMult: 0.35,
    OutAmt7dMult: 3.0,
    MinFwds7d: 7,
    EnterHours: 0,
    ExitHours: 48,
  },
}

type AutofeeService struct {
  db *pgxpool.Pool
  lnd *lndclient.Client
  notifier *Notifier
  logger loggerLike

  mu sync.Mutex
  started bool
  running bool
  stop chan struct{}
  lastRunAt time.Time
  nextRunAt time.Time
  lastError string
}

type loggerLike interface {
  Printf(format string, v ...any)
}

func NewAutofeeService(db *pgxpool.Pool, lnd *lndclient.Client, notifier *Notifier, logger loggerLike) *AutofeeService {
  return &AutofeeService{
    db: db,
    lnd: lnd,
    notifier: notifier,
    logger: logger,
  }
}
func (s *AutofeeService) EnsureSchema(ctx context.Context) error {
  if s.db == nil {
    return errors.New("db not configured")
  }

  _, err := s.db.Exec(ctx, `
create table if not exists autofee_config (
  id integer primary key,
  enabled boolean not null default false,
  profile text not null default 'moderate',
  lookback_days integer not null default 7,
  run_interval_sec integer not null default 14400,
  cooldown_up_sec integer not null default 10800,
  cooldown_down_sec integer not null default 14400,
  amboss_enabled boolean not null default false,
  amboss_token text,
  inbound_passive_enabled boolean not null default false,
  discovery_enabled boolean not null default true,
  explorer_enabled boolean not null default true,
  super_source_enabled boolean not null default false,
  super_source_base_fee_msat integer not null default 1000,
  min_ppm integer not null default 10,
  max_ppm integer not null default 2000,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create table if not exists autofee_channel_settings (
  channel_id bigint primary key,
  channel_point text not null,
  enabled boolean not null default true,
  updated_at timestamptz not null default now()
);

create table if not exists autofee_state (
  channel_id bigint primary key,
  last_ppm integer,
  last_inbound_discount_ppm integer,
  last_seed_ppm integer,
  last_outrate_ppm integer,
  last_outrate_ts timestamptz,
  last_rebal_cost_ppm integer,
  last_rebal_cost_ts timestamptz,
  last_ts timestamptz,
  low_streak integer not null default 0,
  baseline_fwd7d integer not null default 0,
  class_label text,
  class_conf real,
  bias_ema real,
  first_seen_ts timestamptz,
  ss_active boolean,
  ss_ok_since timestamptz,
  ss_bad_since timestamptz,
  explorer_state jsonb
);

create index if not exists autofee_channel_settings_enabled_idx on autofee_channel_settings (enabled);

alter table autofee_config add column if not exists super_source_enabled boolean not null default false;
alter table autofee_config add column if not exists super_source_base_fee_msat integer not null default 1000;
alter table autofee_state add column if not exists ss_active boolean;
alter table autofee_state add column if not exists ss_ok_since timestamptz;
alter table autofee_state add column if not exists ss_bad_since timestamptz;
`)
  if err != nil {
    return err
  }

  _, err = s.db.Exec(ctx, `
insert into autofee_config (id)
values ($1)
on conflict (id) do nothing
`, autofeeConfigID)
  return err
}

func (s *AutofeeService) defaultConfig() AutofeeConfig {
  p := autofeeProfiles["moderate"]
  return AutofeeConfig{
    Enabled: false,
    Profile: p.Name,
    LookbackDays: 7,
    RunIntervalSec: p.RunIntervalSec,
    CooldownUpSec: p.CooldownUpSec,
    CooldownDownSec: p.CooldownDownSec,
    AmbossEnabled: false,
    AmbossTokenSet: false,
    InboundPassiveEnabled: false,
    DiscoveryEnabled: true,
    ExplorerEnabled: true,
    SuperSourceEnabled: false,
    SuperSourceBaseFeeMsat: superSourceBaseFeeMsatDefault,
    MinPpm: 10,
    MaxPpm: 2000,
  }
}

func (s *AutofeeService) GetConfig(ctx context.Context) (AutofeeConfig, error) {
  cfg := s.defaultConfig()
  if s.db == nil {
    return cfg, errors.New("db unavailable")
  }

  var ambossToken pgtype.Text
  err := s.db.QueryRow(ctx, `
select enabled, profile, lookback_days, run_interval_sec, cooldown_up_sec, cooldown_down_sec,
  amboss_enabled, amboss_token, inbound_passive_enabled, discovery_enabled, explorer_enabled,
  super_source_enabled, super_source_base_fee_msat, min_ppm, max_ppm
from autofee_config where id=$1
`, autofeeConfigID).Scan(
    &cfg.Enabled,
    &cfg.Profile,
    &cfg.LookbackDays,
    &cfg.RunIntervalSec,
    &cfg.CooldownUpSec,
    &cfg.CooldownDownSec,
    &cfg.AmbossEnabled,
    &ambossToken,
    &cfg.InboundPassiveEnabled,
    &cfg.DiscoveryEnabled,
    &cfg.ExplorerEnabled,
    &cfg.SuperSourceEnabled,
    &cfg.SuperSourceBaseFeeMsat,
    &cfg.MinPpm,
    &cfg.MaxPpm,
  )
  if err != nil {
    if errors.Is(err, pgx.ErrNoRows) {
      return cfg, nil
    }
    return cfg, err
  }
  cfg.AmbossTokenSet = ambossToken.Valid && strings.TrimSpace(ambossToken.String) != ""
  if cfg.Profile == "" {
    cfg.Profile = "moderate"
  }
  if cfg.LookbackDays < autofeeMinLookbackDays {
    cfg.LookbackDays = autofeeMinLookbackDays
  }
  if cfg.LookbackDays > autofeeMaxLookbackDays {
    cfg.LookbackDays = autofeeMaxLookbackDays
  }
  if cfg.MinPpm <= 0 {
    cfg.MinPpm = 10
  }
  if cfg.MaxPpm <= 0 {
    cfg.MaxPpm = 2000
  }
  return cfg, nil
}

func (s *AutofeeService) UpdateConfig(ctx context.Context, req AutofeeConfigUpdate) (AutofeeConfig, error) {
  current, err := s.GetConfig(ctx)
  if err != nil {
    return current, err
  }

  if req.Enabled != nil {
    current.Enabled = *req.Enabled
  }
  if req.Profile != nil && strings.TrimSpace(*req.Profile) != "" {
    current.Profile = strings.ToLower(strings.TrimSpace(*req.Profile))
    if _, ok := autofeeProfiles[current.Profile]; !ok {
      current.Profile = "moderate"
    }
  }
  if req.LookbackDays != nil {
    current.LookbackDays = *req.LookbackDays
  }
  if req.RunIntervalSec != nil {
    current.RunIntervalSec = *req.RunIntervalSec
  }
  if req.CooldownUpSec != nil {
    current.CooldownUpSec = *req.CooldownUpSec
  }
  if req.CooldownDownSec != nil {
    current.CooldownDownSec = *req.CooldownDownSec
  }
  if req.AmbossEnabled != nil {
    current.AmbossEnabled = *req.AmbossEnabled
  }
  if req.InboundPassiveEnabled != nil {
    current.InboundPassiveEnabled = *req.InboundPassiveEnabled
  }
  if req.DiscoveryEnabled != nil {
    current.DiscoveryEnabled = *req.DiscoveryEnabled
  }
  if req.ExplorerEnabled != nil {
    current.ExplorerEnabled = *req.ExplorerEnabled
  }
  if req.SuperSourceEnabled != nil {
    current.SuperSourceEnabled = *req.SuperSourceEnabled
  }
  if req.SuperSourceBaseFeeMsat != nil {
    current.SuperSourceBaseFeeMsat = *req.SuperSourceBaseFeeMsat
  }
  if req.MinPpm != nil {
    current.MinPpm = *req.MinPpm
  }
  if req.MaxPpm != nil {
    current.MaxPpm = *req.MaxPpm
  }

  if current.LookbackDays < autofeeMinLookbackDays {
    current.LookbackDays = autofeeMinLookbackDays
  }
  if current.LookbackDays > autofeeMaxLookbackDays {
    current.LookbackDays = autofeeMaxLookbackDays
  }
  if current.MinPpm <= 0 {
    current.MinPpm = 10
  }
  if current.MaxPpm <= 0 {
    current.MaxPpm = 2000
  }
  if current.SuperSourceBaseFeeMsat < 0 {
    current.SuperSourceBaseFeeMsat = 0
  }

  if current.RunIntervalSec < 3600 {
    current.RunIntervalSec = 3600
  }
  if current.RunIntervalSec > 86400 {
    current.RunIntervalSec = 86400
  }
  if current.CooldownUpSec < 3600 {
    current.CooldownUpSec = 3600
  }
  if current.CooldownDownSec < 7200 {
    current.CooldownDownSec = 7200
  }

  var ambossToken string
  if req.AmbossToken != nil {
    ambossToken = strings.TrimSpace(*req.AmbossToken)
  } else {
    var raw pgtype.Text
    _ = s.db.QueryRow(ctx, `select amboss_token from autofee_config where id=$1`, autofeeConfigID).Scan(&raw)
    if raw.Valid {
      ambossToken = raw.String
    }
  }

  _, err = s.db.Exec(ctx, `
update autofee_config
set enabled=$2,
  profile=$3,
  lookback_days=$4,
  run_interval_sec=$5,
  cooldown_up_sec=$6,
  cooldown_down_sec=$7,
  amboss_enabled=$8,
  amboss_token=$9,
  inbound_passive_enabled=$10,
  discovery_enabled=$11,
  explorer_enabled=$12,
  super_source_enabled=$13,
  super_source_base_fee_msat=$14,
  min_ppm=$15,
  max_ppm=$16,
  updated_at=now()
where id=$1
`, autofeeConfigID,
    current.Enabled,
    current.Profile,
    current.LookbackDays,
    current.RunIntervalSec,
    current.CooldownUpSec,
    current.CooldownDownSec,
    current.AmbossEnabled,
    ambossToken,
    current.InboundPassiveEnabled,
    current.DiscoveryEnabled,
    current.ExplorerEnabled,
    current.SuperSourceEnabled,
    current.SuperSourceBaseFeeMsat,
    current.MinPpm,
    current.MaxPpm,
  )
  if err != nil {
    return current, err
  }
  current.AmbossTokenSet = strings.TrimSpace(ambossToken) != ""
  return current, nil
}

func (s *AutofeeService) LoadChannelSettings(ctx context.Context) (map[uint64]bool, error) {
  settings := map[uint64]bool{}
  if s.db == nil {
    return settings, errors.New("db unavailable")
  }
  rows, err := s.db.Query(ctx, `select channel_id, enabled from autofee_channel_settings`)
  if err != nil {
    return settings, err
  }
  defer rows.Close()
  for rows.Next() {
    var channelID int64
    var enabled bool
    if err := rows.Scan(&channelID, &enabled); err != nil {
      return settings, err
    }
    settings[uint64(channelID)] = enabled
  }
  return settings, rows.Err()
}

func (s *AutofeeService) SetChannelEnabled(ctx context.Context, channelID uint64, channelPoint string, enabled bool) error {
  if s.db == nil {
    return errors.New("db unavailable")
  }
  if channelID == 0 && strings.TrimSpace(channelPoint) == "" {
    return errors.New("channel_id or channel_point required")
  }
  _, err := s.db.Exec(ctx, `
insert into autofee_channel_settings (channel_id, channel_point, enabled, updated_at)
values ($1, $2, $3, now())
on conflict (channel_id) do update set enabled=excluded.enabled, channel_point=excluded.channel_point, updated_at=excluded.updated_at
`, int64(channelID), strings.TrimSpace(channelPoint), enabled)
  return err
}

func (s *AutofeeService) SetAllChannelsEnabled(ctx context.Context, enabled bool) error {
  if s.db == nil {
    return errors.New("db unavailable")
  }
  if s.lnd == nil {
    return errors.New("lnd unavailable")
  }
  channels, err := s.lnd.ListChannels(ctx)
  if err != nil {
    return err
  }
  batch := &pgx.Batch{}
  for _, ch := range channels {
    batch.Queue(`
insert into autofee_channel_settings (channel_id, channel_point, enabled, updated_at)
values ($1, $2, $3, now())
on conflict (channel_id) do update set enabled=excluded.enabled, channel_point=excluded.channel_point, updated_at=excluded.updated_at
`, int64(ch.ChannelID), ch.ChannelPoint, enabled)
  }
  br := s.db.SendBatch(ctx, batch)
  defer br.Close()
  for range channels {
    if _, err := br.Exec(); err != nil {
      return err
    }
  }
  return nil
}

func (s *AutofeeService) Status() AutofeeStatus {
  s.mu.Lock()
  defer s.mu.Unlock()
  status := AutofeeStatus{
    Running: s.running,
    LastError: s.lastError,
  }
  if !s.lastRunAt.IsZero() {
    status.LastRunAt = s.lastRunAt.UTC().Format(time.RFC3339)
  }
  if !s.nextRunAt.IsZero() {
    status.NextRunAt = s.nextRunAt.UTC().Format(time.RFC3339)
  }
  return status
}

func (s *AutofeeService) Start() {
  s.mu.Lock()
  if s.started {
    s.mu.Unlock()
    return
  }
  s.started = true
  s.stop = make(chan struct{})
  s.mu.Unlock()

  go s.loop()
}

func (s *AutofeeService) Stop() {
  s.mu.Lock()
  if s.stop != nil {
    close(s.stop)
    s.stop = nil
  }
  s.mu.Unlock()
}

func (s *AutofeeService) loop() {
  for {
    cfg, err := s.GetConfig(context.Background())
    if err != nil {
      s.logger.Printf("autofee: config load failed: %v", err)
    }
    interval := time.Duration(cfg.RunIntervalSec) * time.Second
    if interval < time.Hour {
      interval = time.Hour
    }
    if interval > 24*time.Hour {
      interval = 24 * time.Hour
    }
    jitter := time.Duration(rand.Int63n(int64(interval/10)+1)) - time.Duration(int64(interval/20))
    next := time.Now().Add(interval + jitter)
    s.mu.Lock()
    s.nextRunAt = next
    s.mu.Unlock()

    timer := time.NewTimer(time.Until(next))
    select {
    case <-s.stop:
      timer.Stop()
      return
    case <-timer.C:
      _ = s.Run(context.Background(), false, "scheduled")
    }
  }
}

func (s *AutofeeService) Run(ctx context.Context, dryRun bool, reason string) error {
  if s.db == nil {
    return errors.New("db unavailable")
  }
  if s.lnd == nil {
    return errors.New("lnd unavailable")
  }

  s.mu.Lock()
  if s.running {
    s.mu.Unlock()
    return errors.New("autofee already running")
  }
  s.running = true
  s.mu.Unlock()

  defer func() {
    s.mu.Lock()
    s.running = false
    s.lastRunAt = time.Now()
    s.mu.Unlock()
  }()

  cfg, err := s.GetConfig(ctx)
  if err != nil {
    s.setLastError(err)
    return err
  }
  if !cfg.Enabled {
    return nil
  }

  if status, err := s.lnd.GetStatus(ctx); err == nil {
    if !status.SyncedToChain || !status.SyncedToGraph {
      return nil
    }
  }

  engine := newAutofeeEngine(s, cfg)
  if reason == "manual" {
    engine.ignoreCooldown = true
  }
  err = engine.Execute(ctx, dryRun, reason)
  if err != nil {
    s.setLastError(err)
    return err
  }
  s.setLastError(nil)
  return nil
}

func (s *AutofeeService) setLastError(err error) {
  s.mu.Lock()
  defer s.mu.Unlock()
  if err != nil {
    s.lastError = err.Error()
    return
  }
  s.lastError = ""
}
// ===== Engine =====

type autofeeEngine struct {
  svc *AutofeeService
  cfg AutofeeConfig
  profile autofeeProfile
  superSource superSourceThresholds
  ignoreCooldown bool
  now time.Time
}

type autofeeRunSummary struct {
  total int
  inactive int
  disabled int
  eligible int
  applied int
  applyErrors int
  skippedCooldown int
  skippedSmall int
  skippedSame int
  skippedOther int
  seedAmboss int
  seedAmbossMissing int
  seedAmbossError int
  seedAmbossEmpty int
  seedOutrate int
  seedMem int
  seedDefault int
  superSource int
  inboundDiscount int
}

func (s *autofeeRunSummary) addTags(tags []string) {
  for _, tag := range tags {
    switch tag {
    case "seed:amboss":
      s.seedAmboss++
    case "seed:amboss-missing":
      s.seedAmbossMissing++
    case "seed:amboss-error":
      s.seedAmbossError++
    case "seed:amboss-empty":
      s.seedAmbossEmpty++
    case "seed:outrate":
      s.seedOutrate++
    case "seed:mem":
      s.seedMem++
    case "seed:default":
      s.seedDefault++
    case "cooldown":
      s.skippedCooldown++
    case "hold-small":
      s.skippedSmall++
    case "same-ppm":
      s.skippedSame++
    }
  }
}

func newAutofeeEngine(svc *AutofeeService, cfg AutofeeConfig) *autofeeEngine {
  p := autofeeProfiles[cfg.Profile]
  if p.Name == "" {
    p = autofeeProfiles["moderate"]
  }
  ss := superSourceThresholdsByProfile[p.Name]
  if ss.OutRatioMin == 0 {
    ss = superSourceThresholdsByProfile["moderate"]
  }
  return &autofeeEngine{
    svc: svc,
    cfg: cfg,
    profile: p,
    superSource: ss,
    now: time.Now().UTC(),
  }
}

type autofeeChannelState struct {
  ChannelID uint64
  LastPpm int
  LastInboundDiscount int
  LastSeed int
  LastOutrate int
  LastOutrateTs time.Time
  LastRebalCost int
  LastRebalCostTs time.Time
  LastTs time.Time
  LowStreak int
  BaselineFwd7d int
  ClassLabel string
  ClassConf float64
  BiasEma float64
  FirstSeen time.Time
  SuperSourceActive bool
  SuperSourceOkSince time.Time
  SuperSourceBadSince time.Time
  ExplorerState explorerState
}

type explorerState struct {
  Active bool `json:"active"`
  StartedTs int64 `json:"started_ts"`
  Rounds int `json:"rounds"`
  FwdsAtStart int `json:"fwds_at_start"`
  LastExitTs int64 `json:"last_exit_ts"`
  Seen bool `json:"seen"`
}

func (e *autofeeEngine) Execute(ctx context.Context, dryRun bool, reason string) error {
  channels, err := e.svc.lnd.ListChannels(ctx)
  if err != nil {
    return err
  }

  settings, err := e.svc.LoadChannelSettings(ctx)
  if err != nil {
    return err
  }

  state, err := e.loadState(ctx)
  if err != nil {
    return err
  }

  forwardStats, err := e.fetchForwardStats(ctx, e.cfg.LookbackDays)
  if err != nil {
    return err
  }
  forwardStats1d, err := e.fetchForwardStats(ctx, 1)
  if err != nil {
    return err
  }
  forwardStats7d := forwardStats
  if e.cfg.LookbackDays != 7 {
    if stats7d, err := e.fetchForwardStats(ctx, 7); err == nil {
      forwardStats7d = stats7d
    }
  }
  inboundStats, err := e.fetchInboundStats(ctx, e.cfg.LookbackDays)
  if err != nil {
    return err
  }
  rebalStats, err := e.fetchRebalanceStats(ctx, e.cfg.LookbackDays)
  if err != nil {
    return err
  }

  totalOutFeeMsat := int64(0)
  for _, item := range forwardStats {
    totalOutFeeMsat += item.FeeMsat
  }
  rebalGlobal := rebalStats.Global
  rebalGlobalPpm := ppmMsat(rebalGlobal.FeeMsat, rebalGlobal.AmtMsat)

  summary := autofeeRunSummary{total: len(channels)}
  updates := 0
  for _, ch := range channels {
    if !ch.Active {
      summary.inactive++
      continue
    }
    if ch.ChannelID == 0 {
      continue
    }
    enabled, ok := settings[ch.ChannelID]
    if ok && !enabled {
      summary.disabled++
      continue
    }

    st := state[ch.ChannelID]
    decision := e.evaluateChannel(ch, st, forwardStats, forwardStats1d, forwardStats7d, inboundStats, rebalStats, totalOutFeeMsat, rebalGlobalPpm)
    if decision == nil {
      continue
    }
    summary.eligible++
    summary.addTags(decision.Tags)
    if decision.SuperSourceActive {
      summary.superSource++
    }
    if decision.InboundDiscount > 0 {
      summary.inboundDiscount++
    }

    e.persistState(ctx, decision.State)

    if decision.Apply {
      updates++
      summary.applied++
      if dryRun {
        _ = e.logDecision(ctx, "dry-run", decision)
        continue
      }
      if err := e.applyDecision(ctx, ch, decision); err != nil {
        summary.applyErrors++
        _ = e.logDecision(ctx, "error", decision.withError(err))
        continue
      }
      _ = e.logDecision(ctx, "apply", decision)
    } else {
      hasSkipTag := false
      for _, tag := range decision.Tags {
        switch tag {
        case "cooldown", "hold-small", "same-ppm":
          hasSkipTag = true
        }
        if hasSkipTag {
          break
        }
      }
      if !hasSkipTag {
        summary.skippedOther++
      }
    }
  }

  summaryText := fmt.Sprintf(
    "channels=%d eligible=%d applied=%d errors=%d skip{cooldown=%d small=%d same=%d other=%d disabled=%d inactive=%d} seed{amboss=%d missing=%d err=%d empty=%d outrate=%d mem=%d default=%d} super_source=%d inbound_disc=%d",
    summary.total, summary.eligible, summary.applied, summary.applyErrors,
    summary.skippedCooldown, summary.skippedSmall, summary.skippedSame, summary.skippedOther, summary.disabled, summary.inactive,
    summary.seedAmboss, summary.seedAmbossMissing, summary.seedAmbossError, summary.seedAmbossEmpty, summary.seedOutrate, summary.seedMem, summary.seedDefault,
    summary.superSource, summary.inboundDiscount,
  )
  if e.ignoreCooldown {
    summaryText = summaryText + " cooldown_ignored=1"
  }
  _ = e.logSummary(ctx, dryRun, reason, summaryText)
  return nil
}

type forwardStat struct {
  FeeMsat int64
  AmtMsat int64
  Count int64
}

type inboundStat struct {
  AmtMsat int64
  Count int64
}

type rebalStat struct {
  FeeMsat int64
  AmtMsat int64
}

type rebalStats struct {
  ByChannel map[uint64]rebalStat
  Global rebalStat
}

func (e *autofeeEngine) fetchForwardStats(ctx context.Context, lookback int) (map[uint64]forwardStat, error) {
  rows, err := e.svc.db.Query(ctx, `
select chan_id_out, coalesce(sum(fee_msat), 0), coalesce(sum(amount_out_msat), 0), count(*)
from notifications
where type='forward' and occurred_at >= now() - ($1 * interval '1 day')
  and chan_id_out is not null
group by chan_id_out
`, lookback)
  if err != nil {
    return nil, err
  }
  defer rows.Close()
  out := map[uint64]forwardStat{}
  for rows.Next() {
    var chanID int64
    var feeMsat int64
    var amtMsat int64
    var count int64
    if err := rows.Scan(&chanID, &feeMsat, &amtMsat, &count); err != nil {
      return nil, err
    }
    out[uint64(chanID)] = forwardStat{FeeMsat: feeMsat, AmtMsat: amtMsat, Count: count}
  }
  return out, rows.Err()
}

func (e *autofeeEngine) fetchInboundStats(ctx context.Context, lookback int) (map[uint64]inboundStat, error) {
  rows, err := e.svc.db.Query(ctx, `
select chan_id_in, coalesce(sum(amount_in_msat), 0), count(*)
from notifications
where type='forward' and occurred_at >= now() - ($1 * interval '1 day')
  and chan_id_in is not null
group by chan_id_in
`, lookback)
  if err != nil {
    return nil, err
  }
  defer rows.Close()
  out := map[uint64]inboundStat{}
  for rows.Next() {
    var chanID int64
    var amtMsat int64
    var count int64
    if err := rows.Scan(&chanID, &amtMsat, &count); err != nil {
      return nil, err
    }
    out[uint64(chanID)] = inboundStat{AmtMsat: amtMsat, Count: count}
  }
  return out, rows.Err()
}

func (e *autofeeEngine) fetchRebalanceStats(ctx context.Context, lookback int) (rebalStats, error) {
  stats := rebalStats{ByChannel: map[uint64]rebalStat{}}
  rows, err := e.svc.db.Query(ctx, `
select rebal_target_chan_id, coalesce(sum(fee_msat), 0), coalesce(sum(amount_sat), 0)
from notifications
where type='rebalance' and occurred_at >= now() - ($1 * interval '1 day')
  and rebal_target_chan_id is not null
group by rebal_target_chan_id
`, lookback)
  if err != nil {
    return stats, err
  }
  defer rows.Close()
  for rows.Next() {
    var chanID int64
    var feeMsat int64
    var amtSat int64
    if err := rows.Scan(&chanID, &feeMsat, &amtSat); err != nil {
      return stats, err
    }
    stats.ByChannel[uint64(chanID)] = rebalStat{FeeMsat: feeMsat, AmtMsat: amtSat * 1000}
  }
  if err := rows.Err(); err != nil {
    return stats, err
  }

  err = e.svc.db.QueryRow(ctx, `
select coalesce(sum(fee_msat), 0), coalesce(sum(amount_sat), 0)
from notifications
where type='rebalance' and occurred_at >= now() - ($1 * interval '1 day')
`, lookback).Scan(&stats.Global.FeeMsat, &stats.Global.AmtMsat)
  if err != nil && !errors.Is(err, pgx.ErrNoRows) {
    return stats, err
  }
  stats.Global.AmtMsat = stats.Global.AmtMsat * 1000
  return stats, nil
}

func (e *autofeeEngine) loadState(ctx context.Context) (map[uint64]*autofeeChannelState, error) {
  items := map[uint64]*autofeeChannelState{}
  rows, err := e.svc.db.Query(ctx, `
select channel_id, last_ppm, last_inbound_discount_ppm, last_seed_ppm, last_outrate_ppm, last_outrate_ts,
  last_rebal_cost_ppm, last_rebal_cost_ts, last_ts, low_streak, baseline_fwd7d, class_label, class_conf, bias_ema,
  first_seen_ts, ss_active, ss_ok_since, ss_bad_since, explorer_state
from autofee_state
`)
  if err != nil {
    return items, err
  }
  defer rows.Close()

  for rows.Next() {
    var channelID int64
    var st autofeeChannelState
    var lastPpm pgtype.Int4
    var lastInb pgtype.Int4
    var lastSeed pgtype.Int4
    var lastOut pgtype.Int4
    var lastOutTs pgtype.Timestamptz
    var lastRebal pgtype.Int4
    var lastRebalTs pgtype.Timestamptz
    var lastTs pgtype.Timestamptz
    var lowStreak int
    var baseline int
    var classLabel pgtype.Text
    var classConf pgtype.Float8
    var biasEma pgtype.Float8
    var firstSeen pgtype.Timestamptz
    var ssActive pgtype.Bool
    var ssOkSince pgtype.Timestamptz
    var ssBadSince pgtype.Timestamptz
    var explorerRaw []byte
    if err := rows.Scan(&channelID, &lastPpm, &lastInb, &lastSeed, &lastOut, &lastOutTs, &lastRebal, &lastRebalTs, &lastTs,
      &lowStreak, &baseline, &classLabel, &classConf, &biasEma, &firstSeen, &ssActive, &ssOkSince, &ssBadSince, &explorerRaw); err != nil {
      return items, err
    }
    st.ChannelID = uint64(channelID)
    if lastPpm.Valid {
      st.LastPpm = int(lastPpm.Int32)
    }
    if lastInb.Valid {
      st.LastInboundDiscount = int(lastInb.Int32)
    }
    if lastSeed.Valid {
      st.LastSeed = int(lastSeed.Int32)
    }
    if lastOut.Valid {
      st.LastOutrate = int(lastOut.Int32)
    }
    if lastOutTs.Valid {
      st.LastOutrateTs = lastOutTs.Time
    }
    if lastRebal.Valid {
      st.LastRebalCost = int(lastRebal.Int32)
    }
    if lastRebalTs.Valid {
      st.LastRebalCostTs = lastRebalTs.Time
    }
    if lastTs.Valid {
      st.LastTs = lastTs.Time
    }
    st.LowStreak = lowStreak
    st.BaselineFwd7d = baseline
    if classLabel.Valid {
      st.ClassLabel = classLabel.String
    }
    if classConf.Valid {
      st.ClassConf = classConf.Float64
    }
    if biasEma.Valid {
      st.BiasEma = biasEma.Float64
    }
    if firstSeen.Valid {
      st.FirstSeen = firstSeen.Time
    }
    if ssActive.Valid {
      st.SuperSourceActive = ssActive.Bool
    }
    if ssOkSince.Valid {
      st.SuperSourceOkSince = ssOkSince.Time
    }
    if ssBadSince.Valid {
      st.SuperSourceBadSince = ssBadSince.Time
    }
    if len(explorerRaw) > 0 {
      _ = json.Unmarshal(explorerRaw, &st.ExplorerState)
    }
    items[uint64(channelID)] = &st
  }
  return items, rows.Err()
}

func (e *autofeeEngine) persistState(ctx context.Context, st *autofeeChannelState) {
  if st == nil {
    return
  }
  rawExplorer, _ := json.Marshal(st.ExplorerState)
  _, _ = e.svc.db.Exec(ctx, `
insert into autofee_state (
  channel_id, last_ppm, last_inbound_discount_ppm, last_seed_ppm, last_outrate_ppm, last_outrate_ts,
  last_rebal_cost_ppm, last_rebal_cost_ts, last_ts, low_streak, baseline_fwd7d, class_label, class_conf, bias_ema,
  first_seen_ts, ss_active, ss_ok_since, ss_bad_since, explorer_state
) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)
`+
`on conflict (channel_id) do update set
  last_ppm=excluded.last_ppm,
  last_inbound_discount_ppm=excluded.last_inbound_discount_ppm,
  last_seed_ppm=excluded.last_seed_ppm,
  last_outrate_ppm=excluded.last_outrate_ppm,
  last_outrate_ts=excluded.last_outrate_ts,
  last_rebal_cost_ppm=excluded.last_rebal_cost_ppm,
  last_rebal_cost_ts=excluded.last_rebal_cost_ts,
  last_ts=excluded.last_ts,
  low_streak=excluded.low_streak,
  baseline_fwd7d=excluded.baseline_fwd7d,
  class_label=excluded.class_label,
  class_conf=excluded.class_conf,
  bias_ema=excluded.bias_ema,
  first_seen_ts=excluded.first_seen_ts,
  ss_active=excluded.ss_active,
  ss_ok_since=excluded.ss_ok_since,
  ss_bad_since=excluded.ss_bad_since,
  explorer_state=excluded.explorer_state
`, int64(st.ChannelID), nullableInt(int64(st.LastPpm)), nullableInt(int64(st.LastInboundDiscount)),
    nullableInt(int64(st.LastSeed)), nullableInt(int64(st.LastOutrate)), nullableTime(st.LastOutrateTs),
    nullableInt(int64(st.LastRebalCost)), nullableTime(st.LastRebalCostTs), nullableTime(st.LastTs),
    st.LowStreak, st.BaselineFwd7d, nullableString(st.ClassLabel), nullableFloat(st.ClassConf),
    nullableFloat(st.BiasEma), nullableTime(st.FirstSeen), st.SuperSourceActive,
    nullableTime(st.SuperSourceOkSince), nullableTime(st.SuperSourceBadSince), rawExplorer,
  )
}
// ===== decisions =====

type decision struct {
  ChannelID uint64
  ChannelPoint string
  LocalPpm int
  NewPpm int
  Target int
  Floor int
  Tags []string
  InboundDiscount int
  SuperSourceActive bool
  Apply bool
  Error error
  State *autofeeChannelState
}

func (d *decision) withError(err error) *decision {
  d.Error = err
  return d
}

// ===== evaluation =====

func (e *autofeeEngine) evaluateChannel(ch lndclient.ChannelInfo, st *autofeeChannelState, forwardStats map[uint64]forwardStat,
  forwardStats1d map[uint64]forwardStat, forwardStats7d map[uint64]forwardStat, inboundStats map[uint64]inboundStat,
  rebalStats rebalStats, totalOutFeeMsat int64, rebalGlobalPpm int) *decision {

  localPpm := 0
  if ch.FeeRatePpm != nil {
    localPpm = int(*ch.FeeRatePpm)
  } else if st != nil && st.LastPpm > 0 {
    localPpm = st.LastPpm
  }

  if localPpm <= 0 {
    localPpm = e.cfg.MinPpm
  }

  outRatio := 0.5
  if ch.CapacitySat > 0 {
    outRatio = float64(ch.LocalBalanceSat) / float64(ch.CapacitySat)
  }

  fwd := forwardStats[ch.ChannelID]
  fwd1d := forwardStats1d[ch.ChannelID]
  fwd7d := forwardStats7d[ch.ChannelID]
  inb := inboundStats[ch.ChannelID]
  outPpm7d := ppmMsat(fwd.FeeMsat, fwd.AmtMsat)
  fwdCount := int(fwd.Count)
  fwdCount7d := int(fwd7d.Count)

  outAmtSat := fwd.AmtMsat / 1000
  inAmtSat := inb.AmtMsat / 1000
  outAmt1dSat := fwd1d.AmtMsat / 1000
  outAmt7dSat := fwd7d.AmtMsat / 1000

  totalVal := outAmtSat + inAmtSat
  biasRaw := 0.0
  if totalVal > 0 {
    biasRaw = float64(outAmtSat-inAmtSat) / float64(totalVal)
  }

  if st == nil {
    st = &autofeeChannelState{ChannelID: ch.ChannelID}
  }
  if st.FirstSeen.IsZero() {
    st.FirstSeen = e.now
  }

  biasEma := biasRaw
  if st.BiasEma != 0 {
    biasEma = (1.0-0.45)*st.BiasEma + 0.45*biasRaw
  }
  st.BiasEma = biasEma

  classLabel, classConf := classifyChannel(biasEma, outRatio, inb.Count, fwd.Count, st.ClassLabel, st.ClassConf)
  st.ClassLabel = classLabel
  st.ClassConf = classConf

  superSourceActive := false
  superSourceLike := false
  if e.cfg.SuperSourceEnabled {
    superSourceLike = classLabel == "router"
    ssRatio1d := 0.0
    ssRatio7d := 0.0
    if ch.CapacitySat > 0 {
      ssRatio1d = float64(outAmt1dSat) / float64(ch.CapacitySat)
      ssRatio7d = float64(outAmt7dSat) / float64(ch.CapacitySat)
    }
    ssVol1d := ssRatio1d >= e.superSource.OutAmt1dMult
    ssVol7d := ssRatio7d >= e.superSource.OutAmt7dMult
    ssOk := (classLabel == "source" || classLabel == "router") &&
      outRatio >= e.superSource.OutRatioMin &&
      ch.CapacitySat > 0 &&
      (ssVol1d || ssVol7d) &&
      fwdCount7d >= e.superSource.MinFwds7d

    okSince := st.SuperSourceOkSince
    badSince := st.SuperSourceBadSince
    active := st.SuperSourceActive
    if ssOk {
      badSince = time.Time{}
      if okSince.IsZero() {
        okSince = e.now
      }
      if e.now.Sub(okSince) >= time.Duration(e.superSource.EnterHours)*time.Hour {
        active = true
      }
    } else {
      okSince = time.Time{}
      if badSince.IsZero() {
        badSince = e.now
      }
      if active && e.now.Sub(badSince) >= time.Duration(e.superSource.ExitHours)*time.Hour {
        active = false
      }
    }
    superSourceActive = active
    st.SuperSourceActive = active
    st.SuperSourceOkSince = okSince
    st.SuperSourceBadSince = badSince
  }

  seed, seedTags := e.seedForChannel(ch.RemotePubkey, st)
  if seed <= 0 {
    seed = 200
  }
  st.LastSeed = int(seed)

  target := int(seed) + 25

  if outRatio < 0.10 {
    st.LowStreak++
  } else {
    st.LowStreak = 0
  }
  if st.LowStreak >= 1 {
    bumpAcc := math.Min(0.25, float64(st.LowStreak)*0.05)
    if target <= localPpm {
      target = int(math.Ceil(float64(localPpm) * (1.0 + bumpAcc)))
    } else {
      target = int(math.Ceil(float64(target) * (1.0 + bumpAcc)))
    }
  }

  if outRatio < e.profile.LowOutThresh {
    target = int(math.Ceil(float64(target) * 1.02))
  } else if outRatio > e.profile.HighOutThresh {
    target = int(math.Floor(float64(target) * 0.98))
    if fwdCount == 0 && outRatio > 0.60 {
      target = int(math.Floor(float64(target) * 0.985))
    }
  }

  tags := []string{}
  if superSourceActive {
    tags = append(tags, "super-source")
    if superSourceLike {
      tags = append(tags, "super-source-like")
    }
  }
  if outRatio < 0.10 {
    lack := (0.10 - outRatio) / 0.10
    bump := math.Min(e.profile.SurgeBumpMax, 0.5*lack)
    if bump > 0 {
      target = int(math.Ceil(float64(target) * (1.0 + bump)))
      tags = append(tags, fmt.Sprintf("surge+%d%%", int(bump*100)))
    }
  }
  revShare := 0.0
  if totalOutFeeMsat > 0 {
    revShare = float64(fwd.FeeMsat) / float64(totalOutFeeMsat)
  }
  if revShare >= 0.20 && outRatio < 0.30 {
    target = int(math.Ceil(float64(target) * 1.12))
    tags = append(tags, "top-rev")
  }

  baseCostPpm := rebalGlobalPpm
  rebal := rebalStats.ByChannel[ch.ChannelID]
  if rebal.AmtMsat >= 200_000*1000 && rebal.AmtMsat > 0 {
    baseCostPpm = ppmMsat(rebal.FeeMsat, rebal.AmtMsat)
    st.LastRebalCost = baseCostPpm
    st.LastRebalCostTs = e.now
  } else if st.LastRebalCost > 0 && e.now.Sub(st.LastRebalCostTs) <= 21*24*time.Hour {
    baseCostPpm = st.LastRebalCost
  }
  if baseCostPpm < e.cfg.MinPpm {
    baseCostPpm = e.cfg.MinPpm
  }
  marginPpm7d := outPpm7d - int(float64(baseCostPpm)*1.10)
  if marginPpm7d < 0 && fwdCount >= 5 {
    target = int(math.Ceil(float64(target) * 1.05))
    tags = append(tags, "neg-margin")
  }

  discoveryHit := false
  explorerActive := false
  if e.cfg.ExplorerEnabled {
    explorerActive = e.evalExplorer(st, outRatio, fwdCount, ch.LocalBalanceSat, ch.CapacitySat, localPpm)
    if explorerActive {
      tags = append(tags, "explorer")
    }
  }
  if e.cfg.DiscoveryEnabled && fwdCount == 0 && outRatio > 0.40 {
    discoveryHit = true
    tags = append(tags, "discovery")
  }

  if outRatio < 0.10 && target < localPpm {
    target = localPpm
    tags = append(tags, "no-down-low")
  }

  if superSourceActive {
    target = e.cfg.MinPpm
  }

  target = clampInt(target, e.cfg.MinPpm, e.cfg.MaxPpm)

  capFrac := e.profile.StepCap
  if outRatio < 0.03 {
    capFrac = math.Max(capFrac, 0.10)
  } else if outRatio < 0.05 {
    capFrac = math.Max(capFrac, 0.07)
  }
  if fwdCount == 0 && outRatio > 0.60 {
    capFrac = math.Max(capFrac, 0.12)
  }
  if discoveryHit {
    capFrac = math.Max(capFrac, e.profile.DiscoveryStepCapDown)
  }
  if explorerActive {
    capFrac = math.Max(capFrac, e.profile.DiscoveryStepCapDown)
  }

  rawStep := applyStepCap(localPpm, target, capFrac, 5)

  floor := int(math.Ceil(float64(baseCostPpm) * 1.10))
  if outPpm7d > 0 && fwdCount >= 4 {
    peg := int(math.Ceil(float64(outPpm7d) * 1.05))
    if peg > floor {
      floor = peg
      tags = append(tags, "peg")
    }
  }

  if superSourceActive {
    floor = e.cfg.MinPpm
  }

  finalPpm := clampInt(maxInt(rawStep, floor), e.cfg.MinPpm, e.cfg.MaxPpm)

  inboundDiscount := 0
  if e.cfg.InboundPassiveEnabled && classLabel == "sink" && outRatio <= 0.10 && fwdCount >= 5 && marginPpm7d >= 200 {
    anchor := int(math.Ceil(float64(baseCostPpm) * 1.002))
    maxDiscount := int(math.Ceil(float64(finalPpm) * 0.90))
    gap := finalPpm - anchor
    if gap > 0 {
      inboundDiscount = minInt(gap, maxDiscount)
    }
  }

  apply := true
  delta := int(math.Abs(float64(finalPpm - localPpm)))
  rel := float64(delta) / float64(maxInt(1, localPpm))
  if delta < 15 && rel < 0.04 {
    apply = false
    tags = append(tags, "hold-small")
  }
  if finalPpm != localPpm && !e.ignoreCooldown {
    fwdsSince := fwdCount - st.BaselineFwd7d
    cooldownHours := float64(e.cfg.CooldownDownSec) / 3600.0
    if finalPpm > localPpm {
      cooldownHours = float64(e.cfg.CooldownUpSec) / 3600.0
    }
    if !st.LastTs.IsZero() {
      hoursSince := e.now.Sub(st.LastTs).Hours()
      if hoursSince < cooldownHours && fwdsSince < 2 {
        apply = false
        tags = append(tags, "cooldown")
      }
    }
  }

  if fwdCount > 0 {
    if st.BaselineFwd7d > 0 {
      st.BaselineFwd7d = int(math.Round(0.7*float64(st.BaselineFwd7d) + 0.3*float64(fwdCount)))
    } else {
      st.BaselineFwd7d = fwdCount
    }
  }
  st.LastPpm = finalPpm
  if inboundDiscount > 0 {
    st.LastInboundDiscount = inboundDiscount
  }
  if outPpm7d > 0 {
    st.LastOutrate = outPpm7d
    st.LastOutrateTs = e.now
  }

  if finalPpm == localPpm {
    tags = append(tags, "same-ppm")
  }
  if apply && finalPpm != localPpm {
    st.LastTs = e.now
  }

  if explorerActive && finalPpm < localPpm {
    st.ExplorerState.Rounds++
  }

  tags = append(tags, seedTags...)
  return &decision{
    ChannelID: ch.ChannelID,
    ChannelPoint: ch.ChannelPoint,
    LocalPpm: localPpm,
    NewPpm: finalPpm,
    Target: target,
    Floor: floor,
    Tags: tags,
    InboundDiscount: inboundDiscount,
    SuperSourceActive: superSourceActive,
    Apply: apply && finalPpm != localPpm,
    State: st,
  }
}

func (e *autofeeEngine) evalExplorer(st *autofeeChannelState, outRatio float64, fwdCount int, localBal int64, capacity int64, localPpm int) bool {
  nowTs := e.now.Unix()
  active := st.ExplorerState.Active
  if !active {
    daysSince := 999.0
    if !st.LastTs.IsZero() {
      daysSince = e.now.Sub(st.LastTs).Hours() / 24.0
    }
    if daysSince >= 7 && outRatio >= 0.50 && fwdCount <= 5 {
      st.ExplorerState = explorerState{
        Active: true,
        StartedTs: nowTs,
        Rounds: 0,
        FwdsAtStart: fwdCount,
        Seen: true,
      }
      return true
    }
    return false
  }

  hoursSince := float64(nowTs-st.ExplorerState.StartedTs) / 3600.0
  fwdsSince := fwdCount - st.ExplorerState.FwdsAtStart
  if fwdsSince >= 1 || hoursSince >= 48 || st.ExplorerState.Rounds >= 3 {
    st.ExplorerState.Active = false
    st.ExplorerState.LastExitTs = nowTs
    return false
  }
  return true
}
func (e *autofeeEngine) seedForChannel(pubkey string, st *autofeeChannelState) (float64, []string) {
  tags := []string{}
  if e.cfg.AmbossEnabled {
    token, err := e.fetchAmbossToken(context.Background())
    if err != nil {
      tags = append(tags, "seed:amboss-error")
    } else if token == "" {
      tags = append(tags, "seed:amboss-missing")
    } else if pubkey != "" {
      seed, seedTags, err := e.fetchAmbossSeed(pubkey, token)
      if err != nil {
        tags = append(tags, "seed:amboss-error")
      } else if seed > 0 {
        return seed, append(tags, seedTags...)
      } else {
        tags = append(tags, "seed:amboss-empty")
      }
    }
  }

  if st.LastOutrate > 0 && !st.LastOutrateTs.IsZero() && e.now.Sub(st.LastOutrateTs) <= 21*24*time.Hour {
    return float64(st.LastOutrate), append(tags, "seed:outrate")
  }
  if st.LastSeed > 0 {
    return float64(st.LastSeed), append(tags, "seed:mem")
  }
  return 200.0, append(tags, "seed:default")
}

func (e *autofeeEngine) fetchAmbossToken(ctx context.Context) (string, error) {
  var raw pgtype.Text
  err := e.svc.db.QueryRow(ctx, `select amboss_token from autofee_config where id=$1`, autofeeConfigID).Scan(&raw)
  if err != nil && !errors.Is(err, pgx.ErrNoRows) {
    return "", err
  }
  if raw.Valid {
    return strings.TrimSpace(raw.String), nil
  }
  return "", nil
}

type ambossSeriesResp struct {
  Data struct {
    GetNodeMetrics struct {
      HistoricalSeries [][]any `json:"historical_series"`
    } `json:"getNodeMetrics"`
  } `json:"data"`
}

func (e *autofeeEngine) fetchAmbossSeed(pubkey string, token string) (float64, []string, error) {
  vals, err := fetchAmbossSeries(pubkey, token, e.cfg.LookbackDays, "incoming_fee_rate_metrics", "weighted_corrected_mean")
  if err != nil {
    return 0, nil, err
  }
  if len(vals) == 0 {
    return 0, nil, nil
  }
  p65 := percentile(vals, 0.65)
  p95 := percentile(vals, 0.95)
  seed := p65
  tags := []string{}

  incMedian, _ := ambossAvgSeries(pubkey, token, e.cfg.LookbackDays, "incoming_fee_rate_metrics", "median")
  incMean, _ := ambossAvgSeries(pubkey, token, e.cfg.LookbackDays, "incoming_fee_rate_metrics", "mean")
  incStd, _ := ambossAvgSeries(pubkey, token, e.cfg.LookbackDays, "incoming_fee_rate_metrics", "std")

  if incMedian > 0 {
    seed = (1.0-0.30)*seed + 0.30*incMedian
    tags = append(tags, "seed:med")
  }
  if incMean > 0 && incStd > 0 {
    sigmaMu := incStd / incMean
    pen := math.Min(0.15, 0.25*sigmaMu)
    if pen > 0 {
      seed = seed * (1.0 - pen)
      tags = append(tags, fmt.Sprintf("seed:vol-%d%%", int(math.Round(pen*100))))
    }
  }

  incWcorr, _ := ambossAvgSeries(pubkey, token, e.cfg.LookbackDays, "incoming_fee_rate_metrics", "weighted_corrected_mean")
  outWcorr, _ := ambossAvgSeries(pubkey, token, e.cfg.LookbackDays, "outgoing_fee_rate_metrics", "weighted_corrected_mean")
  if incWcorr > 0 && outWcorr > 0 {
    ratio := outWcorr / incWcorr
    f := 1.0 + 0.20*(ratio-1.0)
    if f < 0.80 {
      f = 0.80
    } else if f > 1.50 {
      f = 1.50
    }
    if math.Abs(f-1.0) > 0.001 {
      seed = seed * f
      tags = append(tags, fmt.Sprintf("seed:ratio×%.2f", f))
    }
  }

  if p95 > 0 && seed > p95 {
    seed = p95
    tags = append(tags, "seed:p95cap")
  }
  if seed > 1600 {
    seed = 1600
    tags = append(tags, "seed:absmax")
  }
  tags = append(tags, "seed:amboss")
  return seed, tags, nil
}

func ambossAvgSeries(pubkey string, token string, lookbackDays int, metric string, submetric string) (float64, error) {
  vals, err := fetchAmbossSeries(pubkey, token, lookbackDays, metric, submetric)
  if err != nil {
    return 0, err
  }
  if len(vals) == 0 {
    return 0, nil
  }
  total := 0.0
  for _, v := range vals {
    total += v
  }
  return total / float64(len(vals)), nil
}

func fetchAmbossSeries(pubkey string, token string, lookbackDays int, metric string, submetric string) ([]float64, error) {
  if pubkey == "" || token == "" {
    return nil, nil
  }
  fromDate := time.Now().UTC().Add(-time.Duration(lookbackDays) * 24 * time.Hour).Format("2006-01-02")
  payload := map[string]any{
    "query": `
        query GetNodeMetrics($from: String!, $metric: NodeMetricsKeys!, $pubkey: String!, $submetric: ChannelMetricsKeys) {
          getNodeMetrics(pubkey: $pubkey) {
            historical_series(from: $from, metric: $metric, submetric: $submetric)
          }
        }`,
    "variables": map[string]any{
      "from": fromDate,
      "metric": metric,
      "pubkey": pubkey,
      "submetric": submetric,
    },
  }
  body, _ := json.Marshal(payload)
  req, err := http.NewRequest("POST", "https://api.amboss.space/graphql", bytes.NewReader(body))
  if err != nil {
    return nil, err
  }
  req.Header.Set("Content-Type", "application/json")
  req.Header.Set("Authorization", "Bearer "+token)
  client := &http.Client{Timeout: 20 * time.Second}
  resp, err := client.Do(req)
  if err != nil {
    return nil, err
  }
  defer resp.Body.Close()
  if resp.StatusCode < 200 || resp.StatusCode >= 300 {
    return nil, fmt.Errorf("amboss status %d", resp.StatusCode)
  }
  var result ambossSeriesResp
  if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
    return nil, err
  }
  rows := result.Data.GetNodeMetrics.HistoricalSeries
  vals := make([]float64, 0, len(rows))
  for _, row := range rows {
    if len(row) != 2 {
      continue
    }
    if v, ok := ambossValueToFloat(row[1]); ok {
      vals = append(vals, v)
    }
  }
  return vals, nil
}

func ambossValueToFloat(raw any) (float64, bool) {
  switch v := raw.(type) {
  case float64:
    return v, true
  case float32:
    return float64(v), true
  case int:
    return float64(v), true
  case int64:
    return float64(v), true
  case json.Number:
    f, err := v.Float64()
    if err == nil {
      return f, true
    }
  case string:
    f, err := strconv.ParseFloat(v, 64)
    if err == nil {
      return f, true
    }
  }
  return 0, false
}
func (e *autofeeEngine) applyDecision(ctx context.Context, ch lndclient.ChannelInfo, d *decision) error {
  if ch.ChannelPoint == "" {
    return errors.New("channel_point missing")
  }
  baseFee := int64(0)
  feeRate := int64(d.NewPpm)
  timeLock := int64(0)
  inboundEnabled := false
  inboundRate := int64(0)

  if ch.BaseFeeMsat != nil {
    baseFee = *ch.BaseFeeMsat
  }
  if d.InboundDiscount > 0 {
    inboundEnabled = true
    inboundRate = int64(-absInt(d.InboundDiscount))
  }

  if baseFee == 0 || timeLock == 0 {
    policy, err := e.svc.lnd.GetChannelPolicy(ctx, ch.ChannelPoint)
    if err == nil {
      baseFee = policy.BaseFeeMsat
      timeLock = policy.TimeLockDelta
    }
  }
  if timeLock <= 0 {
    timeLock = 144
  }
  if d.SuperSourceActive && e.cfg.SuperSourceBaseFeeMsat > 0 {
    baseFee = int64(e.cfg.SuperSourceBaseFeeMsat)
  }

  return e.svc.lnd.UpdateChannelFees(ctx, ch.ChannelPoint, false, baseFee, feeRate, timeLock, inboundEnabled, 0, inboundRate)
}

func (e *autofeeEngine) logDecision(ctx context.Context, action string, d *decision) error {
  if d == nil {
    return nil
  }
  memo := fmt.Sprintf("autofee %s: %s %d->%d ppm | target=%d floor=%d | %s",
    action,
    d.ChannelPoint,
    d.LocalPpm,
    d.NewPpm,
    d.Target,
    d.Floor,
    strings.Join(d.Tags, " "),
  )
  evt := Notification{
    OccurredAt: time.Now().UTC(),
    Type: "autofee",
    Action: action,
    Direction: "neutral",
    Status: "SETTLED",
    AmountSat: 0,
    FeeSat: 0,
    FeeMsat: 0,
    ChannelID: int64(d.ChannelID),
    ChannelPoint: d.ChannelPoint,
    Memo: memo,
  }
  eventKey := fmt.Sprintf("autofee:%d:%d", d.ChannelID, time.Now().UnixNano())
  if e.svc.notifier != nil {
    _, _ = e.svc.notifier.upsertNotification(ctx, eventKey, evt)
    return nil
  }
  _, err := e.svc.db.Exec(ctx, `
insert into notifications (event_key, occurred_at, type, action, direction, status, amount_sat, fee_sat, fee_msat, channel_id, channel_point, memo)
values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
on conflict (event_key) do nothing
`, eventKey, evt.OccurredAt, evt.Type, evt.Action, evt.Direction, evt.Status, evt.AmountSat, evt.FeeSat, evt.FeeMsat, evt.ChannelID, evt.ChannelPoint, evt.Memo)
  return err
}

func (e *autofeeEngine) logSummary(ctx context.Context, dryRun bool, reason, summary string) error {
  action := "summary"
  if dryRun {
    action = "summary-dry"
  }
  memo := fmt.Sprintf("autofee summary (%s): %s", reason, summary)
  evt := Notification{
    OccurredAt: time.Now().UTC(),
    Type: "autofee",
    Action: action,
    Direction: "neutral",
    Status: "SETTLED",
    Memo: memo,
  }
  eventKey := fmt.Sprintf("autofee:summary:%d", time.Now().UnixNano())
  if e.svc.notifier != nil {
    _, _ = e.svc.notifier.upsertNotification(ctx, eventKey, evt)
    return nil
  }
  _, err := e.svc.db.Exec(ctx, `
insert into notifications (event_key, occurred_at, type, action, direction, status, amount_sat, fee_sat, fee_msat, memo)
values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
on conflict (event_key) do nothing
`, eventKey, evt.OccurredAt, evt.Type, evt.Action, evt.Direction, evt.Status, evt.AmountSat, evt.FeeSat, evt.FeeMsat, evt.Memo)
  return err
}

// ===== utils =====

func clampInt(v int, min int, max int) int {
  if v < min {
    return min
  }
  if v > max {
    return max
  }
  return v
}

func maxInt(a, b int) int {
  if a > b {
    return a
  }
  return b
}

func minInt(a, b int) int {
  if a < b {
    return a
  }
  return b
}

func absInt(v int) int {
  if v < 0 {
    return -v
  }
  return v
}

func applyStepCap(current int, target int, capFrac float64, minStep int) int {
  if current <= 0 {
    return target
  }
  cap := int(math.Max(float64(minStep), math.Abs(float64(current))*capFrac))
  delta := target - current
  if delta > cap {
    return current + cap
  }
  if delta < -cap {
    return current - cap
  }
  return target
}

func ppmMsat(feeMsat int64, amtMsat int64) int {
  if amtMsat <= 0 {
    return 0
  }
  return int(math.Round(float64(feeMsat) / float64(amtMsat) * 1_000_000))
}

func percentile(vals []float64, q float64) float64 {
  if len(vals) == 0 {
    return 0
  }
  sort.Float64s(vals)
  if len(vals) == 1 {
    return vals[0]
  }
  pos := q * float64(len(vals)-1)
  lo := int(math.Floor(pos))
  hi := int(math.Ceil(pos))
  if lo == hi {
    return vals[lo]
  }
  return vals[lo]*(float64(hi)-pos) + vals[hi]*(pos-float64(lo))
}

func classifyChannel(biasEma float64, outRatio float64, inCount int64, outCount int64, prevLabel string, prevConf float64) (string, float64) {
  label := "unknown"
  conf := 0.0
  if (inCount+outCount) >= 4 {
    if biasEma >= 0.50 && outRatio < 0.15 {
      label = "sink"
      conf = math.Min(1.0, (biasEma-0.50)/(1.0-0.50)+0.3)
    } else if biasEma <= -0.35 && outRatio > 0.55 {
      label = "source"
      conf = math.Min(1.0, ((-biasEma)-0.35)/(1.0-0.35)+0.3)
    } else if math.Abs(biasEma) <= 0.30 && inCount > 0 && outCount > 0 {
      label = "router"
      conf = math.Min(1.0, (0.30-math.Abs(biasEma))/0.30+0.3)
    }
  }
  if label == "unknown" {
    return prevLabel, prevConf
  }
  if prevLabel == "" || prevLabel == "unknown" {
    return label, conf
  }
  if label != prevLabel {
    if conf >= prevConf+0.10 {
      return label, conf
    }
    return prevLabel, prevConf
  }
  return label, math.Min(1.0, 0.5*prevConf+0.5*conf)
}

func nullableFloat(val float64) any {
  if val == 0 {
    return nil
  }
  return val
}

func nullableTime(t time.Time) any {
  if t.IsZero() {
    return nil
  }
  return t
}
