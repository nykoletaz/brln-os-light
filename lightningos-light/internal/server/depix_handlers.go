package server

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

func (s *Server) handleDepixConfig(w http.ResponseWriter, r *http.Request) {
	userKey := strings.TrimSpace(r.URL.Query().Get("user_key"))
	timezone := strings.TrimSpace(r.URL.Query().Get("timezone"))

	svc, _ := s.depixService()
	if svc == nil {
		defaults := depixDefaultSettings()
		_, tz := depixResolveTimezone(timezone, defaults.DefaultTimezone)
		writeJSON(w, http.StatusOK, depixConfig{
			Enabled:             defaults.AppInstalled && defaults.AppEnabled,
			Timezone:            tz,
			MinAmountCents:      defaults.MinAmountCents,
			MaxAmountCents:      defaults.MaxAmountCents,
			DailyLimitCents:     defaults.DailyLimitCents,
			DailyUsedCents:      0,
			DailyRemainingCents: defaults.DailyLimitCents,
			EulenFeeCents:       defaults.EulenFeeCents,
			BrlnFeeBPS:          defaults.BrlnFeeBPS,
			BrlnFeePercent:      depixFeePercentText(defaults.BrlnFeeBPS),
		})
		return
	}

	cfg, err := svc.Config(r.Context(), userKey, timezone)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, cfg)
}

func (s *Server) handleDepixCreateOrder(w http.ResponseWriter, r *http.Request) {
	svc, errMsg := s.depixService()
	if svc == nil {
		msg := strings.TrimSpace(errMsg)
		if msg == "" {
			msg = "depix unavailable"
		}
		writeError(w, http.StatusServiceUnavailable, msg)
		return
	}
	installed, enabled, err := svc.AppState(r.Context())
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, err.Error())
		return
	}
	if !installed || !enabled {
		writeError(w, http.StatusForbidden, "Buy DePix is disabled")
		return
	}

	var payload depixCreateRequest
	if err := readJSON(r, &payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	order, err := svc.CreateOrder(r.Context(), payload)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, order)
}

func (s *Server) handleDepixOrdersList(w http.ResponseWriter, r *http.Request) {
	svc, errMsg := s.depixService()
	if svc == nil {
		msg := strings.TrimSpace(errMsg)
		if msg == "" {
			msg = "depix unavailable"
		}
		writeError(w, http.StatusServiceUnavailable, msg)
		return
	}
	installed, enabled, err := svc.AppState(r.Context())
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, err.Error())
		return
	}
	if !installed || !enabled {
		writeError(w, http.StatusForbidden, "Buy DePix is disabled")
		return
	}
	userKey := strings.TrimSpace(r.URL.Query().Get("user_key"))
	if userKey == "" {
		writeError(w, http.StatusBadRequest, "user_key required")
		return
	}
	limit := 30
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			limit = parsed
		}
	}
	items, err := svc.ListOrders(r.Context(), userKey, limit)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (s *Server) handleDepixOrderGet(w http.ResponseWriter, r *http.Request) {
	svc, errMsg := s.depixService()
	if svc == nil {
		msg := strings.TrimSpace(errMsg)
		if msg == "" {
			msg = "depix unavailable"
		}
		writeError(w, http.StatusServiceUnavailable, msg)
		return
	}
	installed, enabled, err := svc.AppState(r.Context())
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, err.Error())
		return
	}
	if !installed || !enabled {
		writeError(w, http.StatusForbidden, "Buy DePix is disabled")
		return
	}
	userKey := strings.TrimSpace(r.URL.Query().Get("user_key"))
	if userKey == "" {
		writeError(w, http.StatusBadRequest, "user_key required")
		return
	}
	idRaw := strings.TrimSpace(chi.URLParam(r, "id"))
	id, err := strconv.ParseInt(idRaw, 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid order id")
		return
	}
	refreshRaw := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("refresh")))
	refresh := refreshRaw == "1" || refreshRaw == "true" || refreshRaw == "yes"
	order, err := svc.GetOrder(r.Context(), userKey, id, refresh)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, order)
}
