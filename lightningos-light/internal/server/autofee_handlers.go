package server

import (
  "context"
  "errors"
  "net/http"
  "strconv"
  "strings"
  "time"
)

func (s *Server) handleAutofeeConfigGet(w http.ResponseWriter, r *http.Request) {
  svc, errMsg := s.autofeeService()
  if svc == nil {
    if errMsg == "" {
      errMsg = "autofee unavailable"
    }
    writeError(w, http.StatusServiceUnavailable, errMsg)
    return
  }

  ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
  defer cancel()

  cfg, err := svc.GetConfig(ctx)
  if err != nil {
    writeError(w, http.StatusInternalServerError, err.Error())
    return
  }
  writeJSON(w, http.StatusOK, cfg)
}

func (s *Server) handleAutofeeConfigPost(w http.ResponseWriter, r *http.Request) {
  svc, errMsg := s.autofeeService()
  if svc == nil {
    if errMsg == "" {
      errMsg = "autofee unavailable"
    }
    writeError(w, http.StatusServiceUnavailable, errMsg)
    return
  }

  var req AutofeeConfigUpdate
  if err := readJSON(r, &req); err != nil {
    writeError(w, http.StatusBadRequest, "invalid json")
    return
  }

  ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
  defer cancel()

  cfg, err := svc.UpdateConfig(ctx, req)
  if err != nil {
    writeError(w, http.StatusInternalServerError, err.Error())
    return
  }
  writeJSON(w, http.StatusOK, cfg)
}

func (s *Server) handleAutofeeChannelsGet(w http.ResponseWriter, r *http.Request) {
  svc, errMsg := s.autofeeService()
  if svc == nil {
    if errMsg == "" {
      errMsg = "autofee unavailable"
    }
    writeError(w, http.StatusServiceUnavailable, errMsg)
    return
  }

  ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
  defer cancel()

  settings, err := svc.LoadChannelSettings(ctx)
  if err != nil {
    writeError(w, http.StatusInternalServerError, err.Error())
    return
  }

  payload := []map[string]any{}
  for channelID, enabled := range settings {
    payload = append(payload, map[string]any{
      "channel_id": channelID,
      "enabled": enabled,
    })
  }

  writeJSON(w, http.StatusOK, map[string]any{"settings": payload})
}

func (s *Server) handleAutofeeChannelsPost(w http.ResponseWriter, r *http.Request) {
  svc, errMsg := s.autofeeService()
  if svc == nil {
    if errMsg == "" {
      errMsg = "autofee unavailable"
    }
    writeError(w, http.StatusServiceUnavailable, errMsg)
    return
  }

  var req struct {
    ApplyAll bool `json:"apply_all"`
    Enabled *bool `json:"enabled"`
    ChannelID *uint64 `json:"channel_id"`
    ChannelPoint string `json:"channel_point"`
  }
  if err := readJSON(r, &req); err != nil {
    writeError(w, http.StatusBadRequest, "invalid json")
    return
  }
  if req.Enabled == nil {
    writeError(w, http.StatusBadRequest, "enabled required")
    return
  }

  ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
  defer cancel()

  if req.ApplyAll {
    if err := svc.SetAllChannelsEnabled(ctx, *req.Enabled); err != nil {
      writeError(w, http.StatusInternalServerError, err.Error())
      return
    }
    writeJSON(w, http.StatusOK, map[string]any{"ok": true})
    return
  }

  if req.ChannelID == nil && strings.TrimSpace(req.ChannelPoint) == "" {
    writeError(w, http.StatusBadRequest, "channel_id or channel_point required")
    return
  }
  channelID := uint64(0)
  if req.ChannelID != nil {
    channelID = *req.ChannelID
  }
  if err := svc.SetChannelEnabled(ctx, channelID, req.ChannelPoint, *req.Enabled); err != nil {
    writeError(w, http.StatusInternalServerError, err.Error())
    return
  }
  writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleAutofeeRun(w http.ResponseWriter, r *http.Request) {
  svc, errMsg := s.autofeeService()
  if svc == nil {
    if errMsg == "" {
      errMsg = "autofee unavailable"
    }
    writeError(w, http.StatusServiceUnavailable, errMsg)
    return
  }
  var req struct {
    DryRun bool `json:"dry_run"`
  }
  if err := readJSON(r, &req); err != nil {
    writeError(w, http.StatusBadRequest, "invalid json")
    return
  }

  ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
  defer cancel()

  if err := svc.Run(ctx, req.DryRun, "manual"); err != nil {
    writeError(w, http.StatusInternalServerError, err.Error())
    return
  }
  writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleAutofeeStatus(w http.ResponseWriter, r *http.Request) {
  svc, errMsg := s.autofeeService()
  if svc == nil {
    if errMsg == "" {
      errMsg = "autofee unavailable"
    }
    writeError(w, http.StatusServiceUnavailable, errMsg)
    return
  }
  writeJSON(w, http.StatusOK, svc.Status())
}

// Hook into /api/logs when service=autofee
func (s *Server) readAutofeeLogLines(ctx context.Context, limit int) ([]string, error) {
  if s.db == nil {
    return nil, errors.New("db unavailable")
  }
  if limit <= 0 {
    limit = 200
  }
  if limit > 1000 {
    limit = 1000
  }
  rows, err := s.db.Query(ctx, `
select occurred_at, line
from autofee_logs
order by id desc
limit $1
`, limit)
  if err != nil {
    return nil, err
  }
  defer rows.Close()

  lines := []string{}
  for rows.Next() {
    var ts time.Time
    var memo string
    if err := rows.Scan(&ts, &memo); err != nil {
      return nil, err
    }
    line := ts.UTC().Format(time.RFC3339) + " | " + memo
    lines = append(lines, line)
  }
  return lines, rows.Err()
}

func (s *Server) handleAutofeeResults(w http.ResponseWriter, r *http.Request) {
  svc, errMsg := s.autofeeService()
  if svc == nil {
    if errMsg == "" {
      errMsg = "autofee unavailable"
    }
    writeError(w, http.StatusServiceUnavailable, errMsg)
    return
  }
  limit := parseAutofeeLimit(r.URL.Query().Get("lines"))
  if limit <= 0 {
    limit = 50
  }
  ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
  defer cancel()
  rows, err := s.db.Query(ctx, `
select line
from autofee_logs
order by coalesce(run_id, '0')::bigint desc, seq asc
limit $1
`, limit)
  if err != nil {
    writeError(w, http.StatusInternalServerError, err.Error())
    return
  }
  defer rows.Close()
  out := []string{}
  for rows.Next() {
    var line string
    if err := rows.Scan(&line); err != nil {
      writeError(w, http.StatusInternalServerError, err.Error())
      return
    }
    out = append(out, line)
  }
  writeJSON(w, http.StatusOK, map[string]any{"lines": out})
}

func parseAutofeeLimit(raw string) int {
  if raw == "" {
    return 200
  }
  v, err := strconv.Atoi(raw)
  if err != nil {
    return 200
  }
  return v
}
