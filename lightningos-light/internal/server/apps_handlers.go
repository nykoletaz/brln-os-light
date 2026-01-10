package server

import (
  "net/http"

  "github.com/go-chi/chi/v5"
)

func (s *Server) handleAppsList(w http.ResponseWriter, r *http.Request) {
  apps, err := s.appRegistry()
  if err != nil {
    writeError(w, http.StatusInternalServerError, err.Error())
    return
  }
  resp := make([]appInfo, 0, len(apps))
  for _, app := range apps {
    info, infoErr := app.Info(r.Context())
    if infoErr != nil {
      if info.ID == "" {
        info = newAppInfo(app.Definition())
      }
      if info.Installed {
        info.Status = "unknown"
      }
    }
    resp = append(resp, info)
  }
  writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleAppInstall(w http.ResponseWriter, r *http.Request) {
  appID := chi.URLParam(r, "id")
  if appID == "" {
    writeError(w, http.StatusBadRequest, "missing app id")
    return
  }
  app, err := s.appByID(appID)
  if err != nil {
    writeError(w, http.StatusInternalServerError, err.Error())
    return
  }
  if app == nil {
    writeError(w, http.StatusNotFound, "app not found")
    return
  }
  if err := app.Install(r.Context()); err != nil {
    writeError(w, http.StatusInternalServerError, err.Error())
    return
  }
  writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleAppUninstall(w http.ResponseWriter, r *http.Request) {
  appID := chi.URLParam(r, "id")
  if appID == "" {
    writeError(w, http.StatusBadRequest, "missing app id")
    return
  }
  app, err := s.appByID(appID)
  if err != nil {
    writeError(w, http.StatusInternalServerError, err.Error())
    return
  }
  if app == nil {
    writeError(w, http.StatusNotFound, "app not found")
    return
  }
  if err := app.Uninstall(r.Context()); err != nil {
    writeError(w, http.StatusInternalServerError, err.Error())
    return
  }
  writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleAppStart(w http.ResponseWriter, r *http.Request) {
  appID := chi.URLParam(r, "id")
  if appID == "" {
    writeError(w, http.StatusBadRequest, "missing app id")
    return
  }
  app, err := s.appByID(appID)
  if err != nil {
    writeError(w, http.StatusInternalServerError, err.Error())
    return
  }
  if app == nil {
    writeError(w, http.StatusNotFound, "app not found")
    return
  }
  if err := app.Start(r.Context()); err != nil {
    writeError(w, http.StatusInternalServerError, err.Error())
    return
  }
  writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleAppStop(w http.ResponseWriter, r *http.Request) {
  appID := chi.URLParam(r, "id")
  if appID == "" {
    writeError(w, http.StatusBadRequest, "missing app id")
    return
  }
  app, err := s.appByID(appID)
  if err != nil {
    writeError(w, http.StatusInternalServerError, err.Error())
    return
  }
  if app == nil {
    writeError(w, http.StatusNotFound, "app not found")
    return
  }
  if err := app.Stop(r.Context()); err != nil {
    writeError(w, http.StatusInternalServerError, err.Error())
    return
  }
  writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
