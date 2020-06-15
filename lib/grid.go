package lib

import (
	"io/ioutil"
	"log"
	"net/url"
	"strings"
	"sync"
	"time"

	yaml "gopkg.in/yaml.v2"
)

type Grid struct {
	Image         string `yaml:"image"`
	Port          int32  `yaml:"port"`
	BaseURL       string `yaml:"baseURL"`
	HealthCheck   string `yaml:"healthCheck"`
	EntryPoint    string `yaml:"entryPoint"`
	VNCPort       int32  `yaml:"vncPort"`
	Engine        string `yaml:"engine"`
	MachineType   string `yaml:"machineType"`
	CPURequest    string `yaml:"cpuRequest"`
	MemoryRequest string `yaml:"memoryRequest"`
	CPULimit      string `yaml:"cpuLimit"`
	MemoryLimit   string `yaml:"memoryLimit"`
}

type Versions struct {
	Default  string           `yaml:"default"`
	Versions map[string]*Grid `yaml:"versions"`
}

type GridConfig struct {
	lock           sync.RWMutex
	LastReloadTime time.Time
	Grids          map[string]Versions
}

var gridConfig *GridConfig
var gridOnce sync.Once

func GetGridConfig() *GridConfig {
	gridOnce.Do(func() {
		gridConfig = &GridConfig{Grids: make(map[string]Versions), LastReloadTime: time.Now()}
	})
	return gridConfig
}

func loadGridYAML(filename string, v interface{}) error {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(buf, v)
}

func (gc *GridConfig) Load(grids string) error {
	log.Printf("INIT - Loading grid configuration file")
	grid := make(map[string]Versions)
	err := loadGridYAML(grids, &grid)
	if err != nil {
		return err
	}

	gc.lock.Lock()
	defer gc.lock.Unlock()
	gc.Grids = grid
	gc.LastReloadTime = time.Now()
	log.Printf("INIT - Loaded grid configuration from %s", grids)
	return nil
}

func (gc *GridConfig) Find(name string, version string) (*Grid, string, bool) {
	gc.lock.RLock()
	defer gc.lock.RUnlock()
	grid, ok := gc.Grids[name]
	if !ok {
		return nil, "", false
	}

	if version == "" || version == "ANY" {
		log.Printf("Using default version %s", grid.Default)
		version = grid.Default
		if version == "" {
			log.Printf("Default version of %s is not found", name)
			return nil, "", false
		}
	}

	for v, g := range grid.Versions {
		if strings.HasPrefix(v, version) {
			return g, v, true
		}
	}

	return nil, version, false
}

// GridBase Grid base
type GridBase struct {
	Grid    *Grid
	Timeout int
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
	Find(caps Caps, requestID uint64) (GridStarter, bool)
}

// DefaultManager Grid default manager
type DefaultManager struct {
	GridConfig *GridConfig
}

// Find Find grid matching capabilities
func (m *DefaultManager) Find(caps Caps) (GridStarter, bool) {
	gridName := strings.ToLower(caps.Name)
	version := strings.ToLower(caps.Version)
	if gridName == "" {
		gridName = strings.ToLower(caps.PlatformName)
	}

	if version == "" {
		version = strings.ToLower(caps.PlatformVersion)
	}

	log.Printf("Locating grid %s-%s", gridName, version)
	grid, version, ok := m.GridConfig.Find(gridName, version)
	gridBase := GridBase{Grid: grid, Timeout: caps.GridTimeout}
	if !ok {
		log.Printf("Grid %s-%s not found", gridName, version)
		return nil, false
	}

	return GetGridStarter(grid.Engine, gridBase, caps), true
}
