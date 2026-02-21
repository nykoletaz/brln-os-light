package server

import (
	"context"
	"fmt"
	"os"
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
	apiKey := fswapNodeAPIKey()
	if apiKey == "" {
		// Not activated yet — show as not installed
		return info, nil
	}
	info.Installed = true
	info.Status = "running"
	return info, nil
}

func (a fswapApp) Install(_ context.Context) error {
	// Install is a no-op; the actual activation happens through the
	// /api/boleto/activate flow which saves the API key automatically.
	// Returning nil lets the App Store navigate to the pay-boleto page.
	return nil
}

func (a fswapApp) Uninstall(_ context.Context) error {
	// Remove the API key from secrets.env and process environment
	if err := writeEnvFileValue(secretsEnvPath, "FSWAP_NODE_API_KEY", ""); err != nil {
		return fmt.Errorf("failed to remove FSwap API key: %w", err)
	}
	os.Unsetenv("FSWAP_NODE_API_KEY")
	return nil
}

func (a fswapApp) Start(_ context.Context) error { return nil }
func (a fswapApp) Stop(_ context.Context) error  { return nil }
