package server

import (
  "context"
  "errors"
  "net/http"
  "strconv"
  "time"

  "github.com/go-chi/chi/v5"
)

func (s *Server) handleShortcutsGet(w http.ResponseWriter, r *http.Request) {
  svc, errMsg := s.shortcutsService()
  if svc == nil {
    if errMsg == "" {
      errMsg = "shortcuts unavailable"
    }
    writeError(w, http.StatusServiceUnavailable, errMsg)
    return
  }

  ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
  defer cancel()

  items, err := svc.List(ctx)
  if err != nil {
    writeError(w, http.StatusInternalServerError, err.Error())
    return
  }
  writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (s *Server) handleShortcutsPost(w http.ResponseWriter, r *http.Request) {
  svc, errMsg := s.shortcutsService()
  if svc == nil {
    if errMsg == "" {
      errMsg = "shortcuts unavailable"
    }
    writeError(w, http.StatusServiceUnavailable, errMsg)
    return
  }

  var req struct {
    URL string `json:"url"`
    Emoji string `json:"emoji"`
  }
  if err := readJSON(r, &req); err != nil {
    writeError(w, http.StatusBadRequest, "invalid json")
    return
  }

  ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
  defer cancel()

  item, err := svc.Create(ctx, req.URL, req.Emoji)
  if err != nil {
    switch {
    case errors.Is(err, ErrShortcutInvalidURL), errors.Is(err, ErrShortcutInvalidEmoji):
      writeError(w, http.StatusBadRequest, err.Error())
    case errors.Is(err, ErrShortcutExists):
      writeError(w, http.StatusConflict, err.Error())
    default:
      writeError(w, http.StatusInternalServerError, err.Error())
    }
    return
  }
  writeJSON(w, http.StatusCreated, item)
}

func (s *Server) handleShortcutsDelete(w http.ResponseWriter, r *http.Request) {
  svc, errMsg := s.shortcutsService()
  if svc == nil {
    if errMsg == "" {
      errMsg = "shortcuts unavailable"
    }
    writeError(w, http.StatusServiceUnavailable, errMsg)
    return
  }

  idRaw := chi.URLParam(r, "id")
  id, err := strconv.ParseInt(idRaw, 10, 64)
  if err != nil || id <= 0 {
    writeError(w, http.StatusBadRequest, "invalid shortcut id")
    return
  }

  ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
  defer cancel()

  if err := svc.Delete(ctx, id); err != nil {
    if errors.Is(err, ErrShortcutNotFound) {
      writeError(w, http.StatusNotFound, err.Error())
      return
    }
    if errors.Is(err, ErrShortcutProtected) {
      writeError(w, http.StatusForbidden, err.Error())
      return
    }
    writeError(w, http.StatusInternalServerError, err.Error())
    return
  }
  writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
