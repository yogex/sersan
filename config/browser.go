package config

import (
  "io/ioutil"
  "log"
  "strings"
  "sync"
  "time"

  "gopkg.in/yaml.v2"
)

type Browser struct {
  Image       string `yaml:"image"`
  Port        int32  `yaml:"port"`
  BaseURL     string `yaml:"baseURL"`
  HealthCheck string `yaml:"healthCheck"`
  EntryPoint  string `yaml:"entryPoint"`
  VNCPort     int32  `yaml:"vncPort"`
}

type Versions struct {
  Default  string              `yaml:"default"`
  Versions map[string]*Browser `yaml:"versions"`
}

type BrowserConfig struct {
  lock           sync.RWMutex
  LastReloadTime time.Time
  Browsers       map[string]Versions
}

var browserConfig *BrowserConfig
var browserOnce sync.Once

func GetBrowserConfig() *BrowserConfig {
  browserOnce.Do(func() {
    browserConfig = &BrowserConfig{Browsers: make(map[string]Versions), LastReloadTime: time.Now()}
  })
  return browserConfig
}

func loadBrowserYAML(filename string, v interface{}) error {
  buf, err := ioutil.ReadFile(filename)
  if err != nil {
    return err
  }

  return yaml.Unmarshal(buf, v)
}

func (bc *BrowserConfig) Load(browsers string) error {
  log.Print("Loading configuration file")
  br := make(map[string]Versions)
  err := loadBrowserYAML(browsers, &br)
  if err != nil {
    return err
  }

  bc.lock.Lock()
  defer bc.lock.Unlock()
  bc.Browsers = br
  bc.LastReloadTime = time.Now()
  log.Printf("Loaded configuration from %s", browsers)
  return nil
}

func (bc *BrowserConfig) Find(name string, version string) (*Browser, string, bool) {
  bc.lock.RLock()
  defer bc.lock.RUnlock()
  browser, ok := bc.Browsers[name]
  if !ok {
    return nil, "", false
  }

  if version == "" {
    log.Printf("Using default version %s", browser.Default)
    version = browser.Default
    if version == "" {
      return nil, "", false
    }
  }

  for v, b := range browser.Versions {
    if strings.HasPrefix(v, version) {
      return b, v, true
    }
  }

  return nil, version, false
}
