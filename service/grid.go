package service

import (
    "log"
    "net/url"

    "github.com/salestock/sersan/config"
)

// GridBase Grid base
type GridBase struct {
    Grid *config.Browser
}

// StartedGrid Started grid
type StartedGrid struct {
    Name   string
    URL    *url.URL
    Grid   GridBase
    Cancel func()
}

// GridStarter Grid starter
type GridStarter interface {
    StartWithCancel() (*StartedGrid, error)
}

// Manager Grid manager
type Manager interface {
    Find(caps Caps, requestId uint64) (GridStarter, bool)
}

// DefaultManager Grid default manager
type DefaultManager struct {
    BrowserConfig *config.BrowserConfig
}

// Find Find grid matching capabilities
func (m *DefaultManager) Find(caps Caps) (GridStarter, bool) {
    browserName := caps.Name
    version := caps.Version
    log.Printf("Locating grid with browser %s and version %s", browserName, version)
    grid, version, ok := m.BrowserConfig.Find(browserName, version)
    gridBase := GridBase{Grid: grid}
    if !ok {
        log.Printf("Grid not found")
        return nil, false
    }
    return &Kubernetes{
        GridBase: gridBase,
        Caps:     caps,
    }, true
}
