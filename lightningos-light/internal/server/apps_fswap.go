package server

import (
	"context"
	"errors"
	"os"
	"strconv"
	"strings"
)

const (
	fswapAppID           = "fswap"
	fswapAppInstalledEnv = "FSWAP_APP_INSTALLED"
	fswapAppEnabledEnv   = "FSWAP_APP_ENABLED"
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

func parseBoolSetting(raw string, fallback bool) bool {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case "1", "true", "t", "yes", "y", "on":
		return true
	case "0", "false", "f", "no", "n", "off":
		return false
	default:
		return fallback
	}
}

func fswapReadBoolSetting(key string, fallback bool) bool {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return parseBoolSetting(value, fallback)
	}
	value, err := readEnvFileValue(secretsEnvPath, key)
	if err != nil {
		return fallback
	}
	return parseBoolSetting(value, fallback)
}

func fswapAppState() (installed bool, enabled bool) {
	installed = fswapReadBoolSetting(fswapAppInstalledEnv, false)
	enabled = fswapReadBoolSetting(fswapAppEnabledEnv, false)
	if !installed {
		enabled = false
	}
	return installed, enabled
}

func fswapSetAppState(installed bool, enabled bool) error {
	if !installed {
		enabled = false
	}
	installedValue := strconv.FormatBool(installed)
	enabledValue := strconv.FormatBool(enabled)
	if err := writeEnvFileValue(secretsEnvPath, fswapAppInstalledEnv, installedValue); err != nil {
		return err
	}
	if err := writeEnvFileValue(secretsEnvPath, fswapAppEnabledEnv, enabledValue); err != nil {
		return err
	}
	_ = os.Setenv(fswapAppInstalledEnv, installedValue)
	_ = os.Setenv(fswapAppEnabledEnv, enabledValue)
	return nil
}

func (a fswapApp) Info(_ context.Context) (appInfo, error) {
	info := newAppInfo(a.Definition())
	installed, enabled := fswapAppState()
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

func (a fswapApp) Install(_ context.Context) error {
	return fswapSetAppState(true, true)
}

func (a fswapApp) Uninstall(_ context.Context) error {
	return fswapSetAppState(false, false)
}

func (a fswapApp) Start(_ context.Context) error {
	installed, _ := fswapAppState()
	if !installed {
		return errors.New("FSwap is not installed")
	}
	return fswapSetAppState(true, true)
}

func (a fswapApp) Stop(_ context.Context) error {
	installed, _ := fswapAppState()
	if !installed {
		return errors.New("FSwap is not installed")
	}
	return fswapSetAppState(true, false)
}
