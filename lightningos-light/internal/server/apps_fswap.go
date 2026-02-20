package server

import (
	"context"
)

const (
	fswapAppID = "fswap"
)

type fswapApp struct {
	server *Server
}

func newFswapApp(s *Server) appHandler {
	return fswapApp{server: s}
}

func fswapDefinition() appDefinition {
	return appDefinition{
		ID:          fswapAppID,
		Name:        "FSwap",
		Description: "Pague boletos e contas usando sats do seu node Lightning.",
	}
}

func (a fswapApp) Definition() appDefinition {
	return fswapDefinition()
}

func (a fswapApp) Info(_ context.Context) (appInfo, error) {
	info := newAppInfo(a.Definition())
	info.Installed = true
	info.Status = "running"
	return info, nil
}

func (a fswapApp) Install(_ context.Context) error  { return nil }
func (a fswapApp) Uninstall(_ context.Context) error { return nil }
func (a fswapApp) Start(_ context.Context) error     { return nil }
func (a fswapApp) Stop(_ context.Context) error      { return nil }
