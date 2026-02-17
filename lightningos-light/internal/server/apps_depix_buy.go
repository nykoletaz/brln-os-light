package server

import (
	"context"
	"errors"
	"strings"
)

const depixBuyAppID = "depixbuy"

type depixBuyApp struct {
	server *Server
}

func newDepixBuyApp(s *Server) appHandler {
	return depixBuyApp{server: s}
}

func depixBuyDefinition() appDefinition {
	return appDefinition{
		ID:          depixBuyAppID,
		Name:        "Buy DePix",
		Description: "Buy DePix via PIX with BRLN split lock and checkout tracking.",
		Port:        0,
	}
}

func (a depixBuyApp) Definition() appDefinition {
	return depixBuyDefinition()
}

func (a depixBuyApp) Info(ctx context.Context) (appInfo, error) {
	info := newAppInfo(a.Definition())
	svc, errMsg := a.server.depixService()
	if svc == nil {
		if strings.TrimSpace(errMsg) == "" {
			return info, nil
		}
		return info, errors.New(strings.TrimSpace(errMsg))
	}
	installed, enabled, err := svc.AppState(ctx)
	if err != nil {
		return info, err
	}
	if !installed {
		return info, nil
	}
	info.Installed = true
	if enabled {
		info.Status = "running"
	} else {
		info.Status = "stopped"
	}
	return info, nil
}

func (a depixBuyApp) Install(ctx context.Context) error {
	svc, errMsg := a.server.depixService()
	if svc == nil {
		msg := strings.TrimSpace(errMsg)
		if msg == "" {
			msg = "depix unavailable"
		}
		return errors.New(msg)
	}
	return svc.SetAppInstalled(ctx, true, true)
}

func (a depixBuyApp) Uninstall(ctx context.Context) error {
	svc, errMsg := a.server.depixService()
	if svc == nil {
		msg := strings.TrimSpace(errMsg)
		if msg == "" {
			msg = "depix unavailable"
		}
		return errors.New(msg)
	}
	return svc.SetAppInstalled(ctx, false, false)
}

func (a depixBuyApp) Start(ctx context.Context) error {
	svc, errMsg := a.server.depixService()
	if svc == nil {
		msg := strings.TrimSpace(errMsg)
		if msg == "" {
			msg = "depix unavailable"
		}
		return errors.New(msg)
	}
	installed, _, err := svc.AppState(ctx)
	if err != nil {
		return err
	}
	if !installed {
		return errors.New("Buy DePix is not installed")
	}
	return svc.SetAppEnabled(ctx, true)
}

func (a depixBuyApp) Stop(ctx context.Context) error {
	svc, errMsg := a.server.depixService()
	if svc == nil {
		msg := strings.TrimSpace(errMsg)
		if msg == "" {
			msg = "depix unavailable"
		}
		return errors.New(msg)
	}
	installed, _, err := svc.AppState(ctx)
	if err != nil {
		return err
	}
	if !installed {
		return errors.New("Buy DePix is not installed")
	}
	return svc.SetAppEnabled(ctx, false)
}
